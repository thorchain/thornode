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

	"gitlab.com/thorchain/thornode/bifrost/binance"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

// Observer observer service
type Observer struct {
	cfg             config.ObserverConfiguration
	logger          zerolog.Logger
	blockScanner    *BinanceBlockScanner
	storage         TxInStorage
	stopChan        chan struct{}
	thorchainBridge *thorclient.ThorchainBridge
	m               *metrics.Metrics
	wg              *sync.WaitGroup
	errCounter      *prometheus.CounterVec
	addrMgr         *AddressManager
}

// NewObserver create a new instance of Observer
func NewObserver(cfg config.ObserverConfiguration, thorchainBridge *thorclient.ThorchainBridge, addrMgr *AddressManager, bnb *binance.Binance, m *metrics.Metrics) (*Observer, error) {
	scanStorage, err := NewBinanceChanBlockScannerStorage(cfg.ObserverDbPath)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create scan storage")
	}

	logger := log.Logger.With().Str("module", "observer").Logger()

	if !cfg.BlockScanner.EnforceBlockHeight {
		startBlockHeight, err := thorchainBridge.GetBinanceChainStartHeight()
		if nil != err {
			return nil, errors.Wrap(err, "fail to get start block height from thorchain")
		}

		if startBlockHeight > 0 {
			cfg.BlockScanner.StartBlockHeight = startBlockHeight
			logger.Info().Int64("height", cfg.BlockScanner.StartBlockHeight).Msg("resume from last block height known by thorchain")
		} else {
			cfg.BlockScanner.StartBlockHeight, err = bnb.GetHeight()
			if err != nil {
				return nil, errors.Wrap(err, "fail to get binance height")
			}

			logger.Info().Int64("height", cfg.BlockScanner.StartBlockHeight).Msg("Current block height is indeterminate; using current height from Binance.")
		}
	}

	blockScanner, err := NewBinanceBlockScanner(cfg.BlockScanner, scanStorage, bnb.IsTestNet, addrMgr, m)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create block scanner")
	}
	return &Observer{
		cfg:             cfg,
		logger:          logger,
		blockScanner:    blockScanner,
		wg:              &sync.WaitGroup{},
		stopChan:        make(chan struct{}),
		thorchainBridge: thorchainBridge,
		storage:         scanStorage,
		m:               m,
		errCounter:      m.GetCounterVec(metrics.ObserverError),
		addrMgr:         addrMgr,
	}, nil
}

func (o *Observer) Start() error {
	o.wg.Add(1)
	go o.txinsProcessor(o.blockScanner.GetMessages(), 1)
	o.retryAllTx()
	o.wg.Add(1)
	go o.retryTxProcessor()
	o.blockScanner.Start()
	return nil
}

func (o *Observer) retryAllTx() {
	txIns, err := o.storage.GetTxInForRetry(false)
	if nil != err {
		o.logger.Error().Err(err).Msg("fail to get txin for retry")
		o.errCounter.WithLabelValues("fail_get_txin_for_retry", "").Inc()
		return
	}
	for _, item := range txIns {
		o.processOneTxIn(item)
	}
}

func (o *Observer) retryTxProcessor() {
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
			txIns, err := o.storage.GetTxInForRetry(true)
			if nil != err {
				o.errCounter.WithLabelValues("fail_to_get_txin_for_retry", "").Inc()
				o.logger.Error().Err(err).Msg("fail to get txin for retry")
				continue
			}
			for _, item := range txIns {
				select {
				case <-o.stopChan:
					return
				default:
					o.processOneTxIn(item)
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
	if err := o.storage.SetTxInStatus(txIn, types.Processing); nil != err {
		o.errCounter.WithLabelValues("fail_save_txin_local", txIn.BlockHeight).Inc()
		o.logger.Error().Err(err).Msg("fail to save TxIn to local store")
		return
	}
	if err := o.signAndSendToThorchain(txIn); nil != err {
		o.logger.Error().Err(err).Msg("fail to send to thorchain")
		o.errCounter.WithLabelValues("fail_send_to_thorchain", txIn.BlockHeight).Inc()
		if err := o.storage.SetTxInStatus(txIn, types.Failed); nil != err {
			o.logger.Error().Err(err).Msg("fail to save TxIn to local store")
			return
		}
	}
	if err := o.storage.RemoveTxIn(txIn); nil != err {
		o.errCounter.WithLabelValues("fail_remove_from_local_store", txIn.BlockHeight).Inc()
		o.logger.Error().Err(err).Msg("fail to remove txin from local store")
		return
	}
}

func (o *Observer) signAndSendToThorchain(txIn types.TxIn) error {
	txs, err := o.getThorchainTxIns(txIn)
	if nil != err {
		return errors.Wrap(err, "fail to convert txin to thorchain txin")
	}
	stdTx, err := o.thorchainBridge.GetObservationsStdTx(txs)
	if nil != err {
		o.errCounter.WithLabelValues("fail_to_sign", txIn.BlockHeight).Inc()
		return errors.Wrap(err, "fail to sign the tx")
	}
	txID, err := o.thorchainBridge.Send(*stdTx, types.TxSync)
	if nil != err {
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
		if nil != err {
			o.errCounter.WithLabelValues("fail_to_parse_tx_hash", txIn.BlockHeight).Inc()
			return nil, errors.Wrapf(err, "fail to parse tx hash, %s is invalid ", item.Tx)
		}
		sender, err := common.NewAddress(item.Sender)
		if nil != err {
			o.errCounter.WithLabelValues("fail_to_parse_sender", item.Sender).Inc()
			return nil, errors.Wrapf(err, "fail to parse sender,%s is invalid sender address", item.Sender)
		}

		to, err := common.NewAddress(item.To)
		if nil != err {
			o.errCounter.WithLabelValues("fail_to_parse_sender", item.Sender).Inc()
			return nil, errors.Wrapf(err, "fail to parse sender,%s is invalid sender address", item.Sender)
		}

		h, err := strconv.ParseInt(txIn.BlockHeight, 10, 64)
		if nil != err {
			o.errCounter.WithLabelValues("fail to parse block height", txIn.BlockHeight).Inc()
			return nil, errors.Wrapf(err, "fail to parse block height")
		}
		observedPoolPubKey, err := common.NewPubKey(item.ObservedPoolAddress)
		if nil != err {
			o.errCounter.WithLabelValues("fail to parse observed pool address", item.ObservedPoolAddress).Inc()
			return nil, errors.Wrapf(err, "fail to parse observed pool address: %s", item.ObservedPoolAddress)
		}
		txs[i] = stypes.NewObservedTx(
			common.NewTx(txID, sender, to, item.Coins, common.GetBNBGasFee(uint64(len(item.Coins))), item.Memo),
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

	if err := o.blockScanner.Stop(); nil != err {
		o.logger.Error().Err(err).Msg("fail to close block scanner")
	}

	close(o.stopChan)
	o.wg.Wait()
	if err := o.addrMgr.Stop(); nil != err {
		o.logger.Error().Err(err).Msg("fail to stop pool address manager")
	}
	return o.m.Stop()
}
