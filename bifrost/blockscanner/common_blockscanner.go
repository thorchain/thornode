package blockscanner

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

type CommonBlockScannerSupplemental interface {
	GetTxs(height int64) (stypes.TxIn, bool, error)
}

// CommonBlockScanner is used to discover block height
// since both binance and thorchain use cosmos, so this part logic should be the same
type CommonBlockScanner struct {
	cfg            config.BlockScannerConfiguration
	rpcHost        string
	logger         zerolog.Logger
	wg             *sync.WaitGroup
	scanChan       chan Block
	stopChan       chan struct{}
	globalTxsQueue chan stypes.TxIn
	httpClient     *http.Client
	scannerStorage ScannerStorage
	metrics        *metrics.Metrics
	previousBlock  int64
	errorCounter   *prometheus.CounterVec
	supplemental   CommonBlockScannerSupplemental
}

type Block struct {
	Height int64
	Txs    stypes.TxIn
}

// NewCommonBlockScanner create a new instance of CommonBlockScanner
func NewCommonBlockScanner(cfg config.BlockScannerConfiguration, startBlockHeight int64, scannerStorage ScannerStorage, m *metrics.Metrics, supplemental CommonBlockScannerSupplemental) (*CommonBlockScanner, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("host is empty")
	}
	rpcHost := cfg.RPCHost
	if !strings.HasPrefix(rpcHost, "http") {
		rpcHost = fmt.Sprintf("http://%s", rpcHost)
	}

	// check that we can parse our host url
	_, err := url.Parse(rpcHost)
	if err != nil {
		return nil, err
	}

	if scannerStorage == nil {
		return nil, errors.New("scannerStorage is nil")
	}
	if m == nil {
		return nil, errors.New("metrics instance is nil")
	}
	return &CommonBlockScanner{
		cfg:      cfg,
		logger:   log.Logger.With().Str("module", "commonblockscanner").Logger(),
		rpcHost:  rpcHost,
		wg:       &sync.WaitGroup{},
		stopChan: make(chan struct{}),
		scanChan: make(chan Block, cfg.BlockScanProcessors),
		httpClient: &http.Client{
			Timeout: cfg.HttpRequestTimeout,
		},
		scannerStorage: scannerStorage,
		metrics:        m,
		previousBlock:  startBlockHeight,
		errorCounter:   m.GetCounterVec(metrics.CommonBlockScannerError),
		supplemental:   supplemental,
	}, nil
}

// GetHttpClient return the http client used internal to ourside world
// right now we need to use this for test
func (b *CommonBlockScanner) GetHttpClient() *http.Client {
	return b.httpClient
}

// GetMessages return the channel
func (b *CommonBlockScanner) GetMessages() <-chan Block {
	return b.scanChan
}

// Start block scanner
func (b *CommonBlockScanner) Start() {
	b.wg.Add(1)
	go b.scanBlocks()
	b.wg.Add(1)
	go b.retryFailedBlocks()
}

// retryFailedBlocks , if somehow we failed to process a block , it will be retried
func (b *CommonBlockScanner) retryFailedBlocks() {
	b.logger.Debug().Msg("start to retry failed blocks")
	defer b.logger.Debug().Msg("stop retry failed blocks")
	defer b.wg.Done()
	t := time.NewTicker(b.cfg.BlockRetryInterval)
	for {
		select {
		case <-b.stopChan:
			return // bail
		case <-t.C:
			b.retryBlocks(true)
		}
	}
}

func (b *CommonBlockScanner) retryBlocks(failedonly bool) {
	// start up to grab those blocks that we didn't finished
	blocks, err := b.scannerStorage.GetBlocksForRetry(failedonly)
	if err != nil {
		b.errorCounter.WithLabelValues("fail_get_blocks_for_retry", "").Inc()
		b.logger.Error().Err(err).Msg("fail to get blocks for retry")
	}
	b.logger.Debug().Msgf("find %v blocks need to retry", blocks)
	for _, item := range blocks {
		select {
		case <-b.stopChan:
			return // need to bail
		case b.scanChan <- item:
			b.metrics.GetCounter(metrics.TotalRetryBlocks).Inc()
		}
	}
}

// scanBlocks
func (b *CommonBlockScanner) scanBlocks() {
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
	// start up to grab those blocks that we didn't finished
	b.retryBlocks(false)
	for {
		select {
		case <-b.stopChan:
			return
		default:
			currentBlock := b.previousBlock + 1
			txIn, ok, err := b.supplemental.GetTxs(currentBlock)
			if err != nil {
				// don't log an error if we didn't get the block ok
				if !ok {
					b.errorCounter.WithLabelValues("fail_get_block", "").Inc()
					b.logger.Error().Err(err).Msg("fail to get RPCBlock")
				}
				continue
			}
			block := Block{Height: currentBlock, Txs: txIn}
			b.logger.Debug().Int64("current block height", currentBlock).Int64("block height", b.previousBlock).Msgf("Chain %s get block height", b.cfg.ChainID)
			b.previousBlock++
			b.metrics.GetCounter(metrics.TotalBlockScanned).Inc()
			if err := b.scannerStorage.SetBlockScanStatus(block, NotStarted); err != nil {
				b.logger.Error().Err(err).Msg("fail to set block status")
				b.errorCounter.WithLabelValues("fail_set_block_status", strconv.FormatInt(b.previousBlock, 10)).Inc()
				return
			}
			select {
			case <-b.stopChan:
				return
			case b.scanChan <- block:
			}
			b.metrics.GetCounter(metrics.CurrentPosition).Inc()
			if err := b.scannerStorage.SetScanPos(b.previousBlock); err != nil {
				b.errorCounter.WithLabelValues("fail_save_block_pos", strconv.FormatInt(b.previousBlock, 10)).Inc()
				b.logger.Error().Err(err).Msg("fail to save block scan pos")
				// alert!!
				return
			}
		}
	}
}

func (b *CommonBlockScanner) Stop() error {
	b.logger.Debug().Msg("receive stop request")
	defer b.logger.Debug().Msg("common block scanner stopped")
	close(b.stopChan)
	b.wg.Wait()
	return nil
}
