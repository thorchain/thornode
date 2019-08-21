package observer

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/thorchain/bepswap/common"
	stypes "gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"

	"gitlab.com/thorchain/bepswap/observe/config"
	"gitlab.com/thorchain/bepswap/observe/x/statechain"
	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

// Observer observer service
type Observer struct {
	cfg              config.Configuration
	logger           zerolog.Logger
	WebSocket        *WebSocket
	blockScanner     *BinanceBlockScanner
	storage          TxInStorage
	stopChan         chan struct{}
	stateChainBridge *statechain.StateChainBridge
	wg               *sync.WaitGroup
}

// NewObserver create a new instance of Observer
func NewObserver(cfg config.Configuration) (*Observer, error) {
	webSocket, err := NewWebSocket(cfg)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create web socket instance")
	}
	scanStorage, err := NewBinanceChanBlockScannerStorage(cfg.ObserverDbPath)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create scan storage")
	}
	blockScanner, err := NewBinanceBlockScanner(cfg.BlockScannerConfiguration, scanStorage, cfg.DEXHost, cfg.PoolAddress)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create block scanner")
	}
	stateChainBridge, err := statechain.NewStateChainBridge(cfg.StateChainConfiguration)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create new state chain bridge")
	}
	return &Observer{
		cfg:              cfg,
		logger:           log.Logger.With().Str("module", "observer").Logger(),
		WebSocket:        webSocket,
		blockScanner:     blockScanner,
		wg:               &sync.WaitGroup{},
		stopChan:         make(chan struct{}),
		stateChainBridge: stateChainBridge,
		storage:          scanStorage,
	}, nil
}

func (o *Observer) Start() error {
	for idx := 1; idx <= o.cfg.MessageProcessor; idx++ {
		o.wg.Add(1)
		go o.txinsProcessor(o.WebSocket.GetMessages(), idx)
	}
	for idx := o.cfg.MessageProcessor; idx <= o.cfg.MessageProcessor*2; idx++ {
		o.wg.Add(1)
		go o.txinsProcessor(o.blockScanner.GetMessages(), idx)
	}
	o.retryAllTx()
	o.wg.Add(1)
	go o.retryTxProcessor()
	o.blockScanner.Start()
	return o.WebSocket.Start()
}

func (o *Observer) retryAllTx() {
	txIns, err := o.storage.GetTxInForRetry(false)
	if nil != err {
		o.logger.Error().Err(err).Msg("fail to get txin for retry")
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
	t := time.NewTicker(o.cfg.ObserverRetryInterval)
	defer t.Stop()
	for {
		select {
		case <-o.stopChan:
			return
		case <-t.C:
			txIns, err := o.storage.GetTxInForRetry(true)
			if nil != err {
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
				o.logger.Debug().Msg("nothing need to forward to statechain")
				continue
			}
			o.processOneTxIn(txIn)
		}
	}
}
func (o *Observer) processOneTxIn(txIn types.TxIn) {
	if err := o.storage.SetTxInStatus(txIn, types.Processing); nil != err {
		o.logger.Error().Err(err).Msg("fail to save TxIn to local store")
		return
	}
	if err := o.signAndSendToStatechain(txIn); nil != err {
		o.logger.Error().Err(err).Msg("fail to send to statechain")
		if err := o.storage.SetTxInStatus(txIn, types.Failed); nil != err {
			o.logger.Error().Err(err).Msg("fail to save TxIn to local store")
			return
		}
	}
	if err := o.storage.RemoveTxIn(txIn); nil != err {
		o.logger.Error().Err(err).Msg("fail to remove txin from local store")
		return
	}
}
func (o *Observer) signAndSendToStatechain(txIn types.TxIn) error {
	txs, err := o.getStateChainTxIns(txIn)
	if nil != err {
		return errors.Wrap(err, "fail to convert txin to statechain txin")
	}
	signed, err := o.stateChainBridge.Sign(txs)
	if nil != err {
		return errors.Wrap(err, "fail to sign the tx")
	}
	txID, err := o.stateChainBridge.Send(*signed, types.TxSync)
	if nil != err {
		return errors.Wrap(err, "fail to send the tx to statechain")
	}
	o.logger.Info().Str("block", txIn.BlockHeight).Str("statechain hash", txID.String()).Msg("sign and send to statechain successfully")
	return nil
}

// getStateChainTxIns convert to the type statechain expected
// maybe in later we can just refactor this to use the type in statechain
func (o *Observer) getStateChainTxIns(txIn types.TxIn) ([]stypes.TxIn, error) {
	txs := make([]stypes.TxIn, len(txIn.TxArray))
	for i, item := range txIn.TxArray {
		o.logger.Debug().Str("tx-hash", item.Tx).Msg("txInItem")
		txID, err := common.NewTxID(item.Tx)
		if nil != err {
			return nil, errors.Wrapf(err, "fail to parse tx hash, %s is invalid ", item.Tx)
		}
		bnbAddr, err := common.NewBnbAddress(item.Sender)
		if nil != err {
			return nil, errors.Wrapf(err, "fail to parse sender,%s is invalid sender address", item.Sender)
		}
		txs[i] = stypes.NewTxIn(
			txID,
			item.Coins,
			item.Memo,
			bnbAddr,
		)
	}
	return txs, nil
}

// Stop the observer
func (o *Observer) Stop() error {
	o.logger.Debug().Msg("request to stop observer")
	defer o.logger.Debug().Msg("observer stopped")
	if err := o.WebSocket.Stop(); nil != err {
		o.logger.Error().Err(err).Msg("fail to stop websocket")
	}
	if err := o.blockScanner.Stop(); nil != err {
		o.logger.Error().Err(err).Msg("fail to close block scanner")
	}
	close(o.stopChan)
	o.wg.Wait()
	return nil
}
