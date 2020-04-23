package blockscanner

import (
	"errors"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	btypes "gitlab.com/thorchain/thornode/bifrost/blockscanner/types"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
)

type BlockScannerFetcher interface {
	FetchTxs(height int64) (types.TxIn, error)
}

type Block struct {
	Height int64
	Txs    []string
}

// BlockScanner is used to discover block height
type BlockScanner struct {
	cfg             config.BlockScannerConfiguration
	logger          zerolog.Logger
	wg              *sync.WaitGroup
	scanChan        chan int64
	stopChan        chan struct{}
	scannerStorage  ScannerStorage
	metrics         *metrics.Metrics
	previousBlock   int64
	globalTxsQueue  chan types.TxIn
	errorCounter    *prometheus.CounterVec
	thorchainBridge *thorclient.ThorchainBridge
	chainScanner    BlockScannerFetcher
}

// NewBlockScanner create a new instance of BlockScanner
func NewBlockScanner(cfg config.BlockScannerConfiguration, scannerStorage ScannerStorage, m *metrics.Metrics, thorchainBridge *thorclient.ThorchainBridge, chainScanner BlockScannerFetcher) (*BlockScanner, error) {
	var err error
	if scannerStorage == nil {
		return nil, errors.New("scannerStorage is nil")
	}
	if m == nil {
		return nil, errors.New("metrics instance is nil")
	}
	if thorchainBridge == nil {
		return nil, errors.New("thorchain bridge is nil")
	}

	logger := log.Logger.With().Str("module", "blockscanner").Str("chain", cfg.ChainID.String()).Logger()
	scanner := &BlockScanner{
		cfg:             cfg,
		logger:          logger,
		wg:              &sync.WaitGroup{},
		stopChan:        make(chan struct{}),
		scanChan:        make(chan int64),
		scannerStorage:  scannerStorage,
		metrics:         m,
		errorCounter:    m.GetCounterVec(metrics.CommonBlockScannerError),
		thorchainBridge: thorchainBridge,
		chainScanner:    chainScanner,
	}

	scanner.previousBlock, err = scanner.FetchLastHeight()
	return scanner, err
}

// GetMessages return the channel
func (b *BlockScanner) GetMessages() <-chan int64 {
	return b.scanChan
}

// Start block scanner
func (b *BlockScanner) Start(globalTxsQueue chan types.TxIn) {
	b.globalTxsQueue = globalTxsQueue
	b.wg.Add(1)
	go b.scanBlocks()
}

// scanBlocks
func (b *BlockScanner) scanBlocks() {
	b.logger.Debug().Msg("start to scan blocks")
	defer b.logger.Debug().Msg("stop scan blocks")
	defer b.wg.Done()
	currentPos, err := b.scannerStorage.GetScanPos()
	if err != nil {
		b.errorCounter.WithLabelValues("fail_get_scan_pos", "").Inc()
		b.logger.Error().Err(err).Msgf("fail to get current block scan pos, %s will start from %d", b.cfg.ChainID, b.previousBlock)
	} else {
		b.previousBlock = currentPos
	}
	b.metrics.GetCounter(metrics.CurrentPosition).Add(float64(currentPos))

	// start up to grab those blocks
	for {
		select {
		case <-b.stopChan:
			return
		default:
			currentBlock := b.previousBlock + 1
			txIn, err := b.chainScanner.FetchTxs(currentBlock)
			if err != nil {
				// don't log an error if its because the block doesn't exist yet
				if !errors.Is(err, btypes.UnavailableBlock) {
					b.errorCounter.WithLabelValues("fail_get_block", "").Inc()
					b.logger.Error().Err(err).Msg("fail to get RPCBlock")
				}
				continue
			}
			b.logger.Debug().Int64("block height", currentBlock).Int("txs", len(txIn.TxArray))
			b.previousBlock++
			b.metrics.GetCounter(metrics.TotalBlockScanned).Inc()
			if len(txIn.TxArray) == 0 {
				continue
			}
			select {
			case <-b.stopChan:
				return
			case b.globalTxsQueue <- txIn:
			}
			b.metrics.GetCounter(metrics.CurrentPosition).Inc()
			if err := b.scannerStorage.SetScanPos(b.previousBlock); err != nil {
				b.errorCounter.WithLabelValues("fail_save_block_pos", strconv.FormatInt(b.previousBlock, 10)).Inc()
				b.logger.Error().Err(err).Msg("fail to save block scan pos")
				// alert!!
				continue
			}
		}
	}
}

func (b *BlockScanner) FetchLastHeight() (int64, error) {
	// If we've already started scanning, begin where we left off
	currentPos, _ := b.scannerStorage.GetScanPos() // ignore error
	if currentPos > 0 {
		return currentPos, nil
	}

	// if we've configured a starting height, use that
	if b.cfg.StartBlockHeight > 0 {
		return b.cfg.StartBlockHeight, nil
	}

	// attempt to find the height from thorchain
	// wait for thorchain to be caught up first
	if err := b.thorchainBridge.WaitToCatchUp(); err != nil {
		return 0, err
	}

	if b.thorchainBridge != nil {
		var height int64
		if b.cfg.ChainID.Equals(common.THORChain) {
			height, _ = b.thorchainBridge.GetBlockHeight()
		} else {
			height, _ = b.thorchainBridge.GetLastObservedInHeight(b.cfg.ChainID)
		}
		if height > 0 {
			return height, nil
		}
	}

	// TODO: get current block height from RPC chain node

	return 0, nil
}

func (b *BlockScanner) Stop() {
	b.logger.Debug().Msg("receive stop request")
	defer b.logger.Debug().Msg("common block scanner stopped")
	close(b.stopChan)
	b.wg.Wait()
}
