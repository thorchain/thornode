package observer

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/syndtr/goleveldb/leveldb"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/bepswap/observe/config"

	"gitlab.com/thorchain/bepswap/observe/x/statechain"
	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"

	"gitlab.com/thorchain/bepswap/common"
	stypes "gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

// Observer observer service
type Observer struct {
	cfg       config.Configuration
	logger    zerolog.Logger
	Db        *leveldb.DB
	WebSocket *WebSocket
	// TODO not sure how block scan suppose to work, I don't believe we should scan all the blocks
	BlockScan *BlockScan
	stopChan  chan struct{}
	wg        *sync.WaitGroup
}

// NewObserver create a new instance of Observer
func NewObserver(cfg config.Configuration) (*Observer, error) {

	if len(cfg.ObserverDbPath) == 0 {
		return nil, errors.New("observer db path is empty")
	}
	db, err := leveldb.OpenFile(cfg.ObserverDbPath, nil)
	if nil != err {
		return nil, errors.Wrapf(err, "fail to open level db %s", cfg.ObserverDbPath)
	}

	webSocket, err := NewWebSocket(cfg)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create web socket instance")
	}
	return &Observer{
		cfg:       cfg,
		logger:    log.Logger.With().Str("module", "observer").Logger(),
		Db:        db,
		WebSocket: webSocket,
		//BlockScan:    NewBlockScan(blockTxChan),
		wg:       &sync.WaitGroup{},
		stopChan: make(chan struct{}),
	}, nil
}

func (o *Observer) Start() error {
	for idx := 1; idx <= o.cfg.MessageProcessor; idx++ {
		o.wg.Add(1)
		go o.processTxnIn(o.WebSocket.GetMessages(), idx)
	}

	return o.WebSocket.Start()
}

func (o *Observer) processTxnIn(ch <-chan types.TxIn, idx int) {
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
			mode := types.TxSync

			addr, err := sdk.AccAddressFromBech32(o.cfg.RuneAddress)
			if err != nil {
				log.Error().Msgf("Error: %v", err)
			}

			txs := make([]stypes.TxIn, len(txIn.TxArray))
			for i, item := range txIn.TxArray {
				// TODO: don't ignore this error
				txID, _ := common.NewTxID(item.Tx)
				bnbAddr, _ := common.NewBnbAddress(item.Sender)
				txs[i] = stypes.NewTxIn(
					txID,
					item.Coins,
					item.Memo,
					bnbAddr,
				)
			}

			// TODO if the following two step failed , we should retry and
			signed, err := statechain.Sign(txs, addr, o.cfg)
			if nil != err {
				o.logger.Error().Err(err).Msg("fail to sign the tx")
				continue
			}
			txID, err := statechain.Send(signed, mode)
			if nil != err {
				o.logger.Error().Err(err).Msg("fail to send the tx to statechain")
				continue
			}
			o.logger.Debug().Str("txid", txID.String()).Msg("send to statechain successfully")

		}

	}
}

// Stop the observer
func (o *Observer) Stop() error {
	o.logger.Debug().Msg("request to stop observer")
	defer o.logger.Debug().Msg("observer stopped")
	if err := o.WebSocket.Stop(); nil != err {
		o.logger.Error().Err(err).Msg("fail to stop websocket")
	}
	close(o.stopChan)
	o.wg.Wait()
	return o.Db.Close()
}
