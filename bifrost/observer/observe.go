package observer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
	stypes "gitlab.com/thorchain/thornode/x/thorchain/types"
)

// Observer observer service
type Observer struct {
	logger          zerolog.Logger
	chains          map[common.Chain]chainclients.ChainClient
	stopChan        chan struct{}
	pubkeyMgr       pubkeymanager.PubKeyValidator
	globalTxsQueue  chan types.TxIn
	m               *metrics.Metrics
	errCounter      *prometheus.CounterVec
	thorchainBridge *thorclient.ThorchainBridge
}

// NewObserver create a new instance of Observer for chain
func NewObserver(pubkeyMgr pubkeymanager.PubKeyValidator, chains map[common.Chain]chainclients.ChainClient, thorchainBridge *thorclient.ThorchainBridge, m *metrics.Metrics) (*Observer, error) {
	logger := log.Logger.With().Str("module", "observer").Logger()
	return &Observer{
		logger:          logger,
		chains:          chains,
		stopChan:        make(chan struct{}),
		m:               m,
		pubkeyMgr:       pubkeyMgr,
		globalTxsQueue:  make(chan types.TxIn),
		errCounter:      m.GetCounterVec(metrics.ObserverError),
		thorchainBridge: thorchainBridge,
	}, nil
}

func (o *Observer) getChain(chainID common.Chain) (chainclients.ChainClient, error) {
	chain, ok := o.chains[chainID]
	if !ok {
		o.logger.Debug().Str("chain", chainID.String()).Msg("is not supported yet")
		return nil, errors.New("Not supported")
	}
	return chain, nil
}

func (o *Observer) Start() error {
	for _, chain := range o.chains {
		chain.Start(o.globalTxsQueue)
	}
	go o.processTxIns()
	return nil
}

func (o *Observer) processTxIns() {
	for {
		select {
		case <-o.stopChan:
			return
		case txIn := <-o.globalTxsQueue:
			txIn.TxArray = o.filterObservations(txIn.Chain, txIn.TxArray)
			if err := o.signAndSendToThorchain(txIn); err != nil {
				o.logger.Error().Err(err).Msg("fail to send to thorchain")
				o.errCounter.WithLabelValues("fail_send_to_thorchain", txIn.BlockHeight).Inc()
			}
			// check if chain client has OnObservedTxIn method then call it
			chainClient, err := o.getChain(txIn.Chain)
			if err != nil {
				o.logger.Error().Err(err).Msg("fail to retrieve chain client")
				continue
			}
			i, ok := chainClient.(interface{ OnObservedTxIn(txIn types.TxIn) })
			if ok {
				i.OnObservedTxIn(txIn)
			}
		}
	}
}

func (o *Observer) filterObservations(chain common.Chain, items []types.TxInItem) (txs []types.TxInItem) {
	for _, txInItem := range items {
		if ok := o.MatchedAddress(chain, txInItem); !ok {
			continue
		}

		// NOTE: the following could result in the same tx being added
		// twice, which is expected. We want to make sure we generate both
		// a inbound and outbound txn, if we both apply.

		// check if the from address is a valid pool
		if ok, cpi := o.pubkeyMgr.IsValidPoolAddress(txInItem.Sender, chain); ok {
			txInItem.ObservedPoolPubKey = cpi.PubKey
			txs = append(txs, txInItem)
		}
		// check if the to address is a valid pool address
		if ok, cpi := o.pubkeyMgr.IsValidPoolAddress(txInItem.To, chain); ok {
			txInItem.ObservedPoolPubKey = cpi.PubKey
			txs = append(txs, txInItem)
		} else {
			// Apparently we don't recognize where we are sending funds to.
			// Lets check if we should because its an internal transaction
			// moving funds between vaults (for example). If it is, lets
			// manually trigger an update of pubkeys, then check again...
			switch strings.ToLower(txInItem.Memo) {
			case "migrate", "yggdrasil-", "yggdrasil+":
				o.pubkeyMgr.FetchPubKeys()
				if ok, cpi := o.pubkeyMgr.IsValidPoolAddress(txInItem.To, chain); ok {
					txInItem.ObservedPoolPubKey = cpi.PubKey
					txs = append(txs, txInItem)
				}
			}
		}
	}
	return
}

func (o *Observer) MatchedAddress(chain common.Chain, txInItem types.TxInItem) bool {
	// Check if we are migrating our funds...
	if ok := o.isMigration(chain, txInItem.Sender, txInItem.Memo); ok {
		o.logger.Debug().Str("memo", txInItem.Memo).Msg("migrate")
		return true
	}

	// Check if our pool is registering a new yggdrasil pool. Ie
	// sending the staked assets to the user
	if ok := o.isRegisterYggdrasil(chain, txInItem.Sender, txInItem.Memo); ok {
		o.logger.Debug().Str("memo", txInItem.Memo).Msg("yggdrasil+")
		return true
	}

	// Check if out pool is de registering a yggdrasil pool. Ie sending
	// the bond back to the user
	if ok := o.isDeregisterYggdrasil(chain, txInItem.Sender, txInItem.Memo); ok {
		o.logger.Debug().Str("memo", txInItem.Memo).Msg("yggdrasil-")
		return true
	}

	// Check if THORNode are sending from a yggdrasil address
	if ok := o.isYggdrasil(chain, txInItem.Sender); ok {
		o.logger.Debug().Str("assets sent from yggdrasil pool", txInItem.Memo).Msg("fill order")
		return true
	}

	// Check if THORNode are sending to a yggdrasil address
	if ok := o.isYggdrasil(chain, txInItem.To); ok {
		o.logger.Debug().Str("assets to yggdrasil pool", txInItem.Memo).Msg("refill")
		return true
	}

	// outbound message from pool, when it is outbound, it does not matter how much coins THORNode send to customer for now
	if ok := o.isOutboundMsg(chain, txInItem.Sender, txInItem.Memo); ok {
		o.logger.Debug().Str("memo", txInItem.Memo).Msg("outbound")
		return true
	}

	return false
}

// Check if memo is for registering an Asgard vault
func (o *Observer) isMigration(chain common.Chain, addr, memo string) bool {
	return o.isAddrWithMemo(chain, addr, memo, "migrate")
}

// Check if memo is for registering a Yggdrasil vault
func (o *Observer) isRegisterYggdrasil(chain common.Chain, addr, memo string) bool {
	return o.isAddrWithMemo(chain, addr, memo, "yggdrasil+")
}

// Check if memo is for de registering a Yggdrasil vault
func (o *Observer) isDeregisterYggdrasil(chain common.Chain, addr, memo string) bool {
	return o.isAddrWithMemo(chain, addr, memo, "yggdrasil-")
}

// Check if THORNode have an outbound yggdrasil transaction
func (o *Observer) isYggdrasil(chain common.Chain, addr string) bool {
	ok, _ := o.pubkeyMgr.IsValidPoolAddress(addr, chain)
	return ok
}

func (o *Observer) isOutboundMsg(chain common.Chain, addr, memo string) bool {
	return o.isAddrWithMemo(chain, addr, memo, "outbound")
}

func (o *Observer) isAddrWithMemo(chain common.Chain, addr, memo, targetMemo string) bool {
	match, _ := o.pubkeyMgr.IsValidPoolAddress(addr, chain)
	if !match {
		return false
	}
	lowerMemo := strings.ToLower(memo)
	if strings.HasPrefix(lowerMemo, targetMemo) {
		return true
	}
	return false
}

func (o *Observer) signAndSendToThorchain(txIn types.TxIn) error {
	txs, err := o.getThorchainTxIns(txIn)
	if err != nil {
		return fmt.Errorf("fail to convert txin to thorchain txin: %w", err)
	}
	stdTx, err := o.thorchainBridge.GetObservationsStdTx(txs)
	if err != nil {
		o.errCounter.WithLabelValues("fail_to_sign", txIn.BlockHeight).Inc()
		return fmt.Errorf("fail to sign the tx: %w", err)
	}
	txID, err := o.thorchainBridge.Broadcast(*stdTx, types.TxSync)
	if err != nil {
		o.errCounter.WithLabelValues("fail_to_send_to_thorchain", txIn.BlockHeight).Inc()
		return fmt.Errorf("fail to send the tx to thorchain: %w", err)
	}
	o.logger.Info().Str("block", txIn.BlockHeight).Str("thorchain hash", txID.String()).Msg("sign and send to thorchain successfully")
	return nil
}

// getThorchainTxIns convert to the type thorchain expected
// maybe in later THORNode can just refactor this to use the type in thorchain
func (o *Observer) getThorchainTxIns(txIn types.TxIn) (stypes.ObservedTxs, error) {
	txs := make(stypes.ObservedTxs, len(txIn.TxArray))
	o.logger.Debug().Msgf("len %d", len(txIn.TxArray))
	for i, item := range txIn.TxArray {
		o.logger.Debug().Str("tx-hash", item.Tx).Msg("txInItem")
		txID, err := common.NewTxID(item.Tx)
		if err != nil {
			o.errCounter.WithLabelValues("fail_to_parse_tx_hash", txIn.BlockHeight).Inc()
			return nil, fmt.Errorf("fail to parse tx hash, %s is invalid: %w", item.Tx, err)
		}
		sender, err := common.NewAddress(item.Sender)
		if err != nil {
			o.errCounter.WithLabelValues("fail_to_parse_sender", item.Sender).Inc()
			return nil, fmt.Errorf("fail to parse sender,%s is invalid sender address: %w", item.Sender, err)
		}

		to, err := common.NewAddress(item.To)
		if err != nil {
			o.errCounter.WithLabelValues("fail_to_parse_sender", item.Sender).Inc()
			return nil, fmt.Errorf("fail to parse sender,%s is invalid sender address: %w", item.Sender, err)
		}

		h, err := strconv.ParseInt(txIn.BlockHeight, 10, 64)
		if err != nil {
			o.errCounter.WithLabelValues("fail to parse block height", txIn.BlockHeight).Inc()
			return nil, fmt.Errorf("fail to parse block height: %w", err)
		}
		o.logger.Debug().Msgf("pool pubkey %s", item.ObservedPoolPubKey)
		chainAddr, _ := item.ObservedPoolPubKey.GetAddress(txIn.Chain)
		o.logger.Debug().Msgf("%s address %s", txIn.Chain.String(), chainAddr)
		if err != nil {
			o.errCounter.WithLabelValues("fail to parse observed pool address", item.ObservedPoolPubKey.String()).Inc()
			return nil, fmt.Errorf("fail to parse observed pool address: %s: %w", item.ObservedPoolPubKey.String(), err)
		}
		txs[i] = stypes.NewObservedTx(
			common.NewTx(txID, sender, to, item.Coins, item.Gas, item.Memo),
			h,
			item.ObservedPoolPubKey,
		)
	}
	return txs, nil
}

// Stop the observer
func (o *Observer) Stop() error {
	o.logger.Debug().Msg("request to stop observer")
	defer o.logger.Debug().Msg("observer stopped")

	for _, chain := range o.chains {
		chain.Stop()
	}

	close(o.stopChan)
	if err := o.pubkeyMgr.Stop(); err != nil {
		o.logger.Error().Err(err).Msg("fail to stop pool address manager")
	}
	return o.m.Stop()
}
