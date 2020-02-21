package observer

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/prometheus/client_golang/prometheus"

	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
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
		err := chain.Start(o.globalTxsQueue, o.pubkeyMgr, o.m)
		if err != nil {
			o.logger.Error().Err(err).Str("chain", chain.GetChain().String()).Msg("fail to start")
			return err
		}
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
			if err := o.signAndSendToThorchain(txIn); err != nil {
				o.logger.Error().Err(err).Msg("fail to send to thorchain")
				o.errCounter.WithLabelValues("fail_send_to_thorchain", txIn.BlockHeight).Inc()
			}
		}
	}
}

func (o *Observer) signAndSendToThorchain(txIn types.TxIn) error {
	txs, err := o.getThorchainTxIns(txIn)
	if err != nil {
		return errors.Wrap(err, "fail to convert txin to thorchain txin")
	}
	stdTx, err := o.thorchainBridge.GetObservationsStdTx(txs)
	if err != nil {
		o.errCounter.WithLabelValues("fail_to_sign", txIn.BlockHeight).Inc()
		return errors.Wrap(err, "fail to sign the tx")
	}
	txID, err := o.thorchainBridge.Broadcast(*stdTx, types.TxSync)
	if err != nil {
		o.errCounter.WithLabelValues("fail_to_send_to_thorchain", txIn.BlockHeight).Inc()
		return errors.Wrap(err, "fail to send the tx to thorchain")
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
			return nil, errors.Wrapf(err, "fail to parse tx hash, %s is invalid ", item.Tx)
		}
		sender, err := common.NewAddress(item.Sender)
		if err != nil {
			o.errCounter.WithLabelValues("fail_to_parse_sender", item.Sender).Inc()
			return nil, errors.Wrapf(err, "fail to parse sender,%s is invalid sender address", item.Sender)
		}

		to, err := common.NewAddress(item.To)
		if err != nil {
			o.errCounter.WithLabelValues("fail_to_parse_sender", item.Sender).Inc()
			return nil, errors.Wrapf(err, "fail to parse sender,%s is invalid sender address", item.Sender)
		}

		h, err := strconv.ParseInt(txIn.BlockHeight, 10, 64)
		if err != nil {
			o.errCounter.WithLabelValues("fail to parse block height", txIn.BlockHeight).Inc()
			return nil, errors.Wrapf(err, "fail to parse block height")
		}
		o.logger.Debug().Msgf("pool address %s", item.ObservedPoolAddress)
		observedPoolPubKey, err := common.NewPubKey(item.ObservedPoolAddress)
		o.logger.Debug().Msgf("poool address %s", observedPoolPubKey.String())
		bnbAddr, _ := observedPoolPubKey.GetAddress(common.BNBChain)
		o.logger.Debug().Msgf("bnb address %s", bnbAddr)
		if err != nil {
			o.errCounter.WithLabelValues("fail to parse observed pool address", item.ObservedPoolAddress).Inc()
			return nil, errors.Wrapf(err, "fail to parse observed pool address: %s", item.ObservedPoolAddress)
		}
		chain, _ := o.getChain(txIn.Chain)
		txs[i] = stypes.NewObservedTx(
			common.NewTx(txID, sender, to, item.Coins, chain.GetGasFee(uint64(len(item.Coins))), item.Memo),
			h,
			observedPoolPubKey,
		)
	}
	return txs, nil
}

// Stop the observer
func (o *Observer) Stop() error {
	o.logger.Debug().Msg("request to stop observer")
	defer o.logger.Debug().Msg("observer stopped")

	for _, chain := range o.chains {
		if err := chain.Stop(); err != nil {
			o.logger.Error().Err(err).Msgf("fail to close %s block scanner", chain.GetChain().String())
			return err
		}
	}

	close(o.stopChan)	
	if err := o.pubkeyMgr.Stop(); err != nil {
		o.logger.Error().Err(err).Msg("fail to stop pool address manager")
	}
	return o.m.Stop()
}
