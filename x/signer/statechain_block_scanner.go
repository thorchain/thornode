package signer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/thorchain/bepswap/common"

	"gitlab.com/thorchain/bepswap/observe/config"
	"gitlab.com/thorchain/bepswap/observe/x/blockscanner"
	stypes "gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

type StateChainBlockScan struct {
	logger             zerolog.Logger
	wg                 *sync.WaitGroup
	stopChan           chan struct{}
	txOutChan          chan stypes.TxOut
	cfg                config.BlockScannerConfiguration
	scannerStorage     blockscanner.ScannerStorage
	commonBlockScanner *blockscanner.CommonBlockScanner
	chainHost          string
}

// NewStateChainBlockScan create a new instance of statechain block scanner
func NewStateChainBlockScan(cfg config.BlockScannerConfiguration, scanStorage blockscanner.ScannerStorage, chainHost string) (*StateChainBlockScan, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("rpc host is empty")
	}
	if nil == scanStorage {
		return nil, errors.New("scanStorage is nil")
	}
	commonBlockScanner, err := blockscanner.NewCommonBlockScanner(cfg, scanStorage)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create common block scanner")
	}
	return &StateChainBlockScan{
		logger:             log.With().Str("module", "statechainblockscanner").Logger(),
		wg:                 &sync.WaitGroup{},
		stopChan:           make(chan struct{}),
		txOutChan:          make(chan stypes.TxOut),
		cfg:                cfg,
		scannerStorage:     scanStorage,
		commonBlockScanner: commonBlockScanner,
		chainHost:          chainHost,
	}, nil
}

// GetMessages return the channel
func (b *StateChainBlockScan) GetMessages() <-chan stypes.TxOut {
	return b.txOutChan
}

// Start to scan blocks
func (b *StateChainBlockScan) Start() error {
	for idx := 1; idx < b.cfg.BlockScanProcessors; idx++ {
		b.wg.Add(1)
		go b.processBlocks(idx)
	}
	b.commonBlockScanner.Start()
	return nil
}

func (b *StateChainBlockScan) processABlock(blockHeight int64) error {
	uri := url.URL{
		Scheme: "http",
		Host:   b.chainHost,
		Path:   fmt.Sprintf("/swapservice/txoutarray/%v", blockHeight),
	}
	buf, err := b.commonBlockScanner.GetFromHttpWithRetry(uri.String())
	if nil != err {
		return errors.Wrap(err, "fail to get tx out from a block")
	}
	var txOut stypes.TxOut
	if err := json.Unmarshal(buf, &txOut); err != nil {
		return errors.Wrap(err, "fail to unmarshal TxOut")
	}
	if len(txOut.TxArray) == 0 {
		b.logger.Debug().Int64("block", blockHeight).Msg("nothing to process")
		return nil
	}

	for i, txArr := range txOut.TxArray {

		for j, coin := range txArr.Coins {
			amt := coin.Amount.Float64()
			txOut.TxArray[i].Coins[j].Amount = common.Amount(fmt.Sprintf("%.0f", amt))
		}
	}
	b.txOutChan <- txOut
	return nil
}

func (b *StateChainBlockScan) processBlocks(idx int) {
	b.logger.Debug().Int("idx", idx).Msg("start searching tx out in a block")
	defer b.logger.Debug().Int("idx", idx).Msg("stop searching tx out in a block")
	defer b.wg.Done()

	for {
		select {
		case <-b.stopChan: // time to get out
			return
		case block, more := <-b.commonBlockScanner.GetMessages():
			if !more {
				return
			}
			b.logger.Debug().Int64("block", block).Msg("processing block")
			if err := b.processABlock(block); nil != err {
				if errStatus := b.scannerStorage.SetBlockScanStatus(block, blockscanner.Failed); nil != errStatus {
					b.logger.Error().Err(err).Int64("height", block).Msg("fail to set block to fail status")
				}
				b.logger.Error().Err(err).Int64("height", block).Msg("fail to search tx in block")
				// we will have a retry go routine to check it.
				continue
			}
			// set a block as success
			if err := b.scannerStorage.RemoveBlockStatus(block); nil != err {
				b.logger.Error().Err(err).Int64("block", block).Msg("fail to remove block status from data store, thus block will be re processed")
			}
		}
	}
}

// Stop the scanner
func (b *StateChainBlockScan) Stop() error {
	b.logger.Info().Msg("received request to stop state chain block scanner")
	defer b.logger.Info().Msg("statechain block scanner stopped successfully")
	close(b.stopChan)
	b.wg.Wait()
	return nil
}
