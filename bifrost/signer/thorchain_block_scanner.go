package signer

import (
	"strconv"
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

type ThorchainBlockScan struct {
	logger             zerolog.Logger
	wg                 *sync.WaitGroup
	stopChan           chan struct{}
	txOutChan          chan types.TxOut
	keygensChan        chan types.Keygens
	cfg                config.BlockScannerConfiguration
	scannerStorage     blockscanner.ScannerStorage
	commonBlockScanner *blockscanner.CommonBlockScanner
	thorchain          *thorclient.ThorchainBridge
	m                  *metrics.Metrics
	errCounter         *prometheus.CounterVec
	pkm                *PubKeyManager
	cdc                *codec.Codec
}

// NewThorchainBlockScan create a new instance of thorchain block scanner
func NewThorchainBlockScan(cfg config.BlockScannerConfiguration, scanStorage blockscanner.ScannerStorage, thorchain *thorclient.ThorchainBridge, m *metrics.Metrics, pkm *PubKeyManager) (*ThorchainBlockScan, error) {
	if nil == scanStorage {
		return nil, errors.New("scanStorage is nil")
	}
	if nil == m {
		return nil, errors.New("metric is nil")
	}
	commonBlockScanner, err := blockscanner.NewCommonBlockScanner(cfg, scanStorage, m)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create txOut block scanner")
	}
	return &ThorchainBlockScan{
		logger:             log.With().Str("module", "thorchainblockscanner").Logger(),
		wg:                 &sync.WaitGroup{},
		stopChan:           make(chan struct{}),
		txOutChan:          make(chan types.TxOut),
		keygensChan:        make(chan types.Keygens),
		cfg:                cfg,
		scannerStorage:     scanStorage,
		commonBlockScanner: commonBlockScanner,
		thorchain:          thorchain,
		errCounter:         m.GetCounterVec(metrics.ThorchainBlockScannerError),
		pkm:                pkm,
		cdc:                codec.New(),
	}, nil
}

// GetMessages return the channel
func (b *ThorchainBlockScan) GetTxOutMessages() <-chan types.TxOut {
	return b.txOutChan
}

func (b *ThorchainBlockScan) GetKeygenMessages() <-chan types.Keygens {
	return b.keygensChan
}

// Start to scan blocks
func (b *ThorchainBlockScan) Start() error {
	b.wg.Add(1)
	go b.processBlocks(1)
	b.commonBlockScanner.Start()
	return nil
}

func (b *ThorchainBlockScan) processKeygenBlock(blockHeight int64) error {
	for _, pk := range b.pkm.pks {
		keygens, err := b.thorchain.GetKeygens(blockHeight, pk.String())
		if err != nil {
			return errors.Wrap(err, "fail to get keygens from block scanner")
		}
		b.keygensChan <- *keygens
	}
	return nil
}

func (b *ThorchainBlockScan) processTxOutBlock(blockHeight int64) error {
	for _, pk := range b.pkm.pks {
		if len(pk.String()) == 0 {
			continue
		}
		tx, err := b.thorchain.GetKeysign(blockHeight, pk.String())
		if err != nil {
			return errors.Wrap(err, "fail to get keysign from block scanner")
		}

		for c, out := range tx.Chains {
			b.logger.Debug().Str("chain", c.String()).Msg("chain")
			if len(out.TxArray) == 0 {
				b.logger.Debug().Int64("block", blockHeight).Msg("nothing to process")
				b.m.GetCounter(metrics.BlockNoTxOut).Inc()
				return nil
			}
			b.txOutChan <- out
		}
	}
	return nil
}

func (b *ThorchainBlockScan) processBlocks(idx int) {
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
			if err := b.processTxOutBlock(block); nil != err {
				if errStatus := b.scannerStorage.SetBlockScanStatus(block, blockscanner.Failed); nil != errStatus {
					b.errCounter.WithLabelValues("fail_set_block_Status", strconv.FormatInt(block, 10))
					b.logger.Error().Err(err).Int64("height", block).Msg("fail to set block to fail status")
				}
				b.errCounter.WithLabelValues("fail_search_tx", strconv.FormatInt(block, 10))
				b.logger.Error().Err(err).Int64("height", block).Msg("fail to search tx in block")
				// THORNode will have a retry go routine to check it.
				continue
			}

			// set a block as success
			if err := b.scannerStorage.RemoveBlockStatus(block); nil != err {
				b.errCounter.WithLabelValues("fail_remove_block_Status", strconv.FormatInt(block, 10))
				b.logger.Error().Err(err).Int64("block", block).Msg("fail to remove block status from data store, thus block will be re processed")
			}

			// Intentionally not covering this before the block is marked as
			// success. This is because we don't care if keygen is successful
			// or not.
			b.logger.Debug().Int64("block", block).Msg("processing keygen block")
			if err := b.processKeygenBlock(block); nil != err {
				b.errCounter.WithLabelValues("fail_process_keygen", strconv.FormatInt(block, 10))
				b.logger.Error().Err(err).Int64("height", block).Msg("fail to process keygen")
			}
		}
	}
}

// Stop the scanner
func (b *ThorchainBlockScan) Stop() error {
	b.logger.Info().Msg("received request to stop thorchain block scanner")
	defer b.logger.Info().Msg("thorchain block scanner stopped successfully")
	close(b.stopChan)
	b.wg.Wait()
	return nil
}
