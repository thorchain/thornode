package observer

import (
	"errors"
	"fmt"
	"strconv"

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

const maxTxArrayLen = 100

// Observer observer service
type Observer struct {
	logger            zerolog.Logger
	chains            map[common.Chain]chainclients.ChainClient
	stopChan          chan struct{}
	pubkeyMgr         pubkeymanager.PubKeyValidator
	globalTxsQueue    chan types.TxIn
	globalErrataQueue chan types.ErrataBlock
	m                 *metrics.Metrics
	errCounter        *prometheus.CounterVec
	thorchainBridge   *thorclient.ThorchainBridge
}

// NewObserver create a new instance of Observer for chain
func NewObserver(pubkeyMgr pubkeymanager.PubKeyValidator, chains map[common.Chain]chainclients.ChainClient, thorchainBridge *thorclient.ThorchainBridge, m *metrics.Metrics) (*Observer, error) {
	logger := log.Logger.With().Str("module", "observer").Logger()
	return &Observer{
		logger:            logger,
		chains:            chains,
		stopChan:          make(chan struct{}),
		m:                 m,
		pubkeyMgr:         pubkeyMgr,
		globalTxsQueue:    make(chan types.TxIn),
		globalErrataQueue: make(chan types.ErrataBlock),
		errCounter:        m.GetCounterVec(metrics.ObserverError),
		thorchainBridge:   thorchainBridge,
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
		chain.Start(o.globalTxsQueue, o.globalErrataQueue)
	}
	go o.processTxIns()
	go o.processErrataTx()
	return nil
}

func (o *Observer) processTxIns() {
	for {
		select {
		case <-o.stopChan:
			return
		case txIn := <-o.globalTxsQueue:
			txIn.TxArray = o.filterObservations(txIn.Chain, txIn.TxArray)
			for _, txIn := range o.chunkify(txIn) {
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

				i, ok := chainClient.(interface {
					OnObservedTxIn(txIn types.TxInItem, blockHeight int64)
				})
				if ok {
					height, err := strconv.ParseInt(txIn.BlockHeight, 10, 64)
					if err != nil {
						o.logger.Error().Err(err).Msg("fail to parse block height")
						continue
					}
					for _, item := range txIn.TxArray {
						if o.isOutboundMsg(txIn.Chain, item.Sender) {
							continue
						}
						i.OnObservedTxIn(item, height)
					}
				}
			}
		}
	}
}

func (o *Observer) isOutboundMsg(chain common.Chain, fromAddr string) bool {
	matchOutbound, _ := o.pubkeyMgr.IsValidPoolAddress(fromAddr, chain)
	if matchOutbound {
		return true
	}
	return false
}

// chunkify - breaks the observations into 100 transactions per observation
func (o *Observer) chunkify(txIn types.TxIn) (result []types.TxIn) {
	for len(txIn.TxArray) > 0 {
		newTx := types.TxIn{
			BlockHeight: txIn.BlockHeight,
			Chain:       txIn.Chain,
		}
		if len(txIn.TxArray) > maxTxArrayLen {
			newTx.Count = fmt.Sprintf("%d", maxTxArrayLen)
			newTx.TxArray = txIn.TxArray[:maxTxArrayLen]
			txIn.TxArray = txIn.TxArray[maxTxArrayLen:]
		} else {
			newTx.Count = fmt.Sprintf("%d", len(txIn.TxArray))
			newTx.TxArray = txIn.TxArray
			txIn.TxArray = nil
		}
		result = append(result, newTx)
	}
	return result
}

func (o *Observer) filterObservations(chain common.Chain, items []types.TxInItem) (txs []types.TxInItem) {
	for _, txInItem := range items {
		// NOTE: the following could result in the same tx being added
		// twice, which is expected. We want to make sure we generate both
		// a inbound and outbound txn, if we both apply.

		// check if the from address is a valid pool
		if ok, cpi := o.pubkeyMgr.IsValidPoolAddress(txInItem.Sender, chain); ok {
			txInItem.ObservedVaultPubKey = cpi.PubKey
			txs = append(txs, txInItem)
		}
		// check if the to address is a valid pool address
		if ok, cpi := o.pubkeyMgr.IsValidPoolAddress(txInItem.To, chain); ok {
			txInItem.ObservedVaultPubKey = cpi.PubKey
			txs = append(txs, txInItem)
		}
	}
	return
}

func (o *Observer) processErrataTx() {
	for {
		select {
		case <-o.stopChan:
			return
		case errataBlock, more := <-o.globalErrataQueue:
			if !more {
				return
			}
			o.logger.Info().Msgf("Received a errata block %+v from the Thorchain", errataBlock.Height)
			for _, errataTx := range errataBlock.Txs {
				if err := o.sendErrataTxToThorchain(errataBlock.Height, errataTx.TxID, errataTx.Chain); err != nil {
					o.errCounter.WithLabelValues("fail_to_broadcast_errata_tx", "").Inc()
					o.logger.Error().Err(err).Msg("fail to broadcast errata tx")
				}
			}
		}
	}
}

func (o *Observer) sendErrataTxToThorchain(height int64, txID common.TxID, chain common.Chain) error {
	stdTx, err := o.thorchainBridge.GetErrataStdTx(txID, chain)
	strHeight := strconv.FormatInt(height, 10)
	if err != nil {
		o.errCounter.WithLabelValues("fail_to_sign", strHeight).Inc()
		return fmt.Errorf("fail to sign the tx: %w", err)
	}
	txID, err = o.thorchainBridge.Broadcast(*stdTx, types.TxSync)
	if err != nil {
		o.errCounter.WithLabelValues("fail_to_send_to_thorchain", strHeight).Inc()
		return fmt.Errorf("fail to send the tx to thorchain: %w", err)
	}
	o.logger.Info().Int64("block", height).Str("thorchain hash", txID.String()).Msg("sign and send to thorchain successfully")
	return nil
}

func (o *Observer) signAndSendToThorchain(txIn types.TxIn) error {
	nodeStatus, err := o.thorchainBridge.FetchNodeStatus()
	if err != nil {
		return fmt.Errorf("failed to get node status: %w", err)
	}
	if nodeStatus != stypes.Active {
		return nil
	}
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
		o.logger.Debug().Msgf("pool pubkey %s", item.ObservedVaultPubKey)
		chainAddr, _ := item.ObservedVaultPubKey.GetAddress(txIn.Chain)
		o.logger.Debug().Msgf("%s address %s", txIn.Chain.String(), chainAddr)
		if err != nil {
			o.errCounter.WithLabelValues("fail to parse observed pool address", item.ObservedVaultPubKey.String()).Inc()
			return nil, fmt.Errorf("fail to parse observed pool address: %s: %w", item.ObservedVaultPubKey.String(), err)
		}
		txs[i] = stypes.NewObservedTx(
			common.NewTx(txID, sender, to, item.Coins, item.Gas, item.Memo),
			h,
			item.ObservedVaultPubKey,
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
