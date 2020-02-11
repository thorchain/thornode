package observer

import (
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/common"
	stypes "gitlab.com/thorchain/thornode/x/thorchain/types"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

// Observer observer service
type Observer struct {
	cfg             config.ObserverConfiguration
	logger          zerolog.Logger
	chains          map[common.Chain]chainclients.ChainClient
	stopChan        chan struct{}
	thorchainBridge *thorclient.ThorchainBridge
	m               *metrics.Metrics
	wg              *sync.WaitGroup
	errCounter      *prometheus.CounterVec
	pubkeyMgr       pubkeymanager.PubKeyValidator
	txQueue         chan types.TxIn
}

// NewObserver create a new instance of Observer for chain
func NewObserver(cfg config.ObserverConfiguration, thorchainBridge *thorclient.ThorchainBridge, pubkeyMgr pubkeymanager.PubKeyValidator, chains map[common.Chain]chainclients.ChainClient, m *metrics.Metrics) (*Observer, error) {
	logger := log.Logger.With().Str("module", "observer").Logger()

	idx := 0
	for _, chain := range chains {
		if !cfg.BlockScanners[idx].EnforceBlockHeight {
			startBlockHeight, err := thorchainBridge.GetLastObservedInHeight(chain.GetChain())
			if err != nil {
				return nil, errors.Wrap(err, "fail to get start block height from thorchain")
			}

			if startBlockHeight > 0 {
				cfg.BlockScanners[idx].StartBlockHeight = startBlockHeight
				logger.Info().Int64("height", cfg.BlockScanners[idx].StartBlockHeight).Msg("resume from last block height known by thorchain")
			} else {
				cfg.BlockScanners[idx].StartBlockHeight, err = chain.GetHeight()
				if err != nil {
					return nil, errors.Wrap(err, "fail to get binance height")
				}

				logger.Info().Int64("height", cfg.BlockScanners[idx].StartBlockHeight).Msg("Current block height is indeterminate; using current height from Binance.")
			}
		}
		err := chain.InitBlockScanner(cfg.BlockScanners[idx], pubkeyMgr, m)
		if err != nil {
			return nil, err
		}
		idx++
	}

	return &Observer{
		cfg:             cfg,
		logger:          logger,
		chains:          chains,
		wg:              &sync.WaitGroup{},
		stopChan:        make(chan struct{}),
		thorchainBridge: thorchainBridge,
		m:               m,
		errCounter:      m.GetCounterVec(metrics.ObserverError),
		pubkeyMgr:       pubkeyMgr,
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

func (o *Observer) Start() {
	for _, chain := range o.chains {
		o.startChain(chain)
	}
	o.wg.Add(1)
	go o.processTxIns()
}

func (o *Observer) startChain(chain chainclients.ChainClient) {
	o.wg.Add(1)
	go o.txinsProcessor(chain.GetMessages(), 1)
	o.retryAllTx(chain)
	o.wg.Add(1)
	go o.retryTxProcessor(chain)
	chain.Start()
}

func (o *Observer) processTxIns() {
	for {
		select {
		case <-o.stopChan:
			return
		case txIn := <-o.txQueue:
			o.processOneTxIn(txIn)
		}
	}
}

func (o *Observer) retryAllTx(chain chainclients.ChainClient) {
	txIns, err := chain.GetTxInForRetry(false)
	if err != nil {
		o.logger.Error().Err(err).Msg("fail to get txin for retry")
		o.errCounter.WithLabelValues("fail_get_txin_for_retry", "").Inc()
		return
	}
	for _, item := range txIns {
		o.txQueue <- item
	}
}

func (o *Observer) retryTxProcessor(chain chainclients.ChainClient) {
	o.logger.Info().Msg("start retry process")
	defer o.logger.Info().Msg("stop retry process")
	defer o.wg.Done()
	// retry all
	t := time.NewTicker(o.cfg.RetryInterval)
	defer t.Stop()
	for {
		select {
		case <-o.stopChan:
			return
		case <-t.C:
			txIns, err := chain.GetTxInForRetry(true)
			if err != nil {
				o.errCounter.WithLabelValues("fail_to_get_txin_for_retry", "").Inc()
				o.logger.Error().Err(err).Msg("fail to get txin for retry")
				continue
			}
			for _, item := range txIns {
				select {
				case <-o.stopChan:
					return
				default:
					o.txQueue <- item
				}
			}
		}
	}
}

func (o *Observer) txinsProcessor(ch <-chan types.TxIn, idx int) {
	o.logger.Info().Int("idx", idx).Msg("start to process tx in")
	defer o.logger.Info().Int("idx", idx).Msg("stop to process tx in")
	defer o.wg.Done()
	for {
		select {
		case <-o.stopChan:
			return
		case txIn, more := <-ch:
			if !more {
				// channel closed
				return
			}
			if len(txIn.TxArray) == 0 {
				o.logger.Debug().Msg("nothing need to forward to thorchain")
				continue
			}
			o.processOneTxIn(txIn)
		}
	}
}

func (o *Observer) processOneTxIn(txIn types.TxIn) {
	chain, _ := o.getChain(txIn.Chain)
	if err := chain.SetTxInStatus(txIn, types.Processing); err != nil {
		o.errCounter.WithLabelValues("fail_save_txin_local", txIn.BlockHeight).Inc()
		o.logger.Error().Err(err).Msg("fail to save TxIn to local store")
		return
	}
	if err := o.signAndSendToThorchain(txIn); err != nil {
		o.logger.Error().Err(err).Msg("fail to send to thorchain")
		o.errCounter.WithLabelValues("fail_send_to_thorchain", txIn.BlockHeight).Inc()
		if err := chain.SetTxInStatus(txIn, types.Failed); err != nil {
			o.logger.Error().Err(err).Msg("fail to save TxIn to local store")
			return
		}
	}
	if err := chain.RemoveTxIn(txIn); err != nil {
		o.errCounter.WithLabelValues("fail_remove_from_local_store", txIn.BlockHeight).Inc()
		o.logger.Error().Err(err).Msg("fail to remove txin from local store")
		return
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
		observedPoolPubKey, err := common.NewPubKey(item.ObservedPoolAddress)
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
		}
	}

	close(o.stopChan)
	o.wg.Wait()
	if err := o.pubkeyMgr.Stop(); err != nil {
		o.logger.Error().Err(err).Msg("fail to stop pool address manager")
	}
	return o.m.Stop()
}
