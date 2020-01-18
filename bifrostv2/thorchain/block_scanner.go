package thorchain

import (
	"strconv"
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/blockscanner"
	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/metrics"
	"gitlab.com/thorchain/thornode/bifrostv2/vaultmanager"
	stypes "gitlab.com/thorchain/thornode/x/thorchain/types"
)

// BlockScanner represents a block scanner specific to Thorchain
type BlockScanner struct {
	logger                zerolog.Logger
	wg                    *sync.WaitGroup
	stopChan              chan struct{}
	txOutChan             chan stypes.TxOut
	keygensChan           chan stypes.Keygens
	cfg                   config.BlockScannerConfiguration
	scannerStorage        blockscanner.ScannerStorage
	commonBlockScannerner *blockscanner.CommonBlockScanner
	thorchain             *Client
	metrics               *metrics.Metrics
	errCounter            *prometheus.CounterVec
	vaultMgr              *vaultmanager.VaultManager
	cdc                   *codec.Codec
}

// NewBlockScanner create a new instance of thorchain block scanner
func NewBlockScanner(cfg config.BlockScannerConfiguration, scanStorage blockscanner.ScannerStorage, thorchain *Client, m *metrics.Metrics, vaultMgr *vaultmanager.VaultManager) (*BlockScanner, error) {
	if nil == scanStorage {
		return nil, errors.New("scanStorage is nil")
	}
	if nil == m {
		return nil, errors.New("metric is nil")
	}
	commonBlockScannerner, err := blockscanner.NewCommonBlockScanner(cfg, scanStorage, m)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create txOut block scanner")
	}
	return &BlockScanner{
		logger:                log.With().Str("module", "thorchain_block_scanner").Logger(),
		wg:                    &sync.WaitGroup{},
		stopChan:              make(chan struct{}),
		txOutChan:             make(chan stypes.TxOut),
		keygensChan:           make(chan stypes.Keygens),
		cfg:                   cfg,
		scannerStorage:        scanStorage,
		commonBlockScannerner: commonBlockScannerner,
		thorchain:             thorchain,
		errCounter:            m.GetCounterVec(metrics.ThorchainBlockScannerError),
		metrics:               m,
		vaultMgr:              vaultMgr,
		cdc:                   codec.New(),
	}, nil
}

// GetTxOutMessages return the channel with TxOut messages from thorchain
func (b *BlockScanner) GetTxOutMessages() <-chan stypes.TxOut {
	return b.txOutChan
}

// GetKeygenMessages return the channel with keygen messages from thorchain
func (b *BlockScanner) GetKeygenMessages() <-chan stypes.Keygens {
	return b.keygensChan
}

// Start to scan blocks
func (b *BlockScanner) Start() error {
	// Start block scanner
	b.wg.Add(1)
	go b.processBlocks(1)
	b.commonBlockScannerner.Start()
	return nil
}

// processTxOutBlock retrieve txout from this block height and pass results to
// txOutChan
func (b *BlockScanner) processTxOutBlock(blockHeight int64) error {
	for _, pk := range b.vaultMgr.GetPubKeys() {
		if len(pk.String()) == 0 {
			continue
		}
		txOut, err := b.thorchain.GetKeysign(blockHeight, pk.String())
		if err != nil {
			return err
		}
		b.txOutChan <- *txOut
	}
	return nil
}

// processKeygenBlock retrieve keygen from this block height and pass results to
// keygensChan
func (b *BlockScanner) processKeygenBlock(blockHeight int64) error {
	for _, pk := range b.vaultMgr.GetPubKeys() {
		if len(pk.String()) == 0 {
			continue
		}
		keygens, err := b.thorchain.GetKeygens(blockHeight, pk.String())
		if err != nil {
			return err
		}
		b.keygensChan <- *keygens
	}
	return nil
}

func (b *BlockScanner) processBlocks(idx int) {
	b.logger.Debug().Int("idx", idx).Msg("start searching tx out in a block")
	defer b.logger.Debug().Int("idx", idx).Msg("stop searching tx out in a block")
	defer b.wg.Done()

	for {
		select {
		case <-b.stopChan: // time to get out
			return
		case block, more := <-b.commonBlockScannerner.GetMessages():
			if !more {
				return
			}
			b.logger.Debug().Int64("block", block).Msg("processing block")
			if err := b.processTxOutBlock(block); nil != err {
				if errStatus := b.scannerStorage.SetBlockScannerStatus(block, blockscanner.Failed); nil != errStatus {
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
func (b *BlockScanner) Stop() error {
	b.logger.Info().Msg("received request to stop thorchain block scanner")
	defer b.logger.Info().Msg("thorchain block scanner stopped successfully")
	close(b.stopChan)
	b.wg.Wait()
	return nil
}
