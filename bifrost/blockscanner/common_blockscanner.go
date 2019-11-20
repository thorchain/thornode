package blockscanner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	btypes "gitlab.com/thorchain/bepswap/thornode/bifrost/binance/types"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/config"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/metrics"
)

// CommonBlockScanner is used to discover block height
// since both binance and thorchain use cosmos, so this part logic should be the same
type CommonBlockScanner struct {
	cfg            config.BlockScannerConfiguration
	logger         zerolog.Logger
	wg             *sync.WaitGroup
	scanChan       chan int64
	stopChan       chan struct{}
	httpClient     *http.Client
	scannerStorage ScannerStorage
	metrics        *metrics.Metrics
	previousBlock  int64
	errorCounter   *prometheus.CounterVec
}

// NewCommonBlockScanner create a new instance of CommonBlockScanner
func NewCommonBlockScanner(cfg config.BlockScannerConfiguration, scannerStorage ScannerStorage, m *metrics.Metrics) (*CommonBlockScanner, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("host is empty")
	}
	if nil == scannerStorage {
		return nil, errors.New("scannerStorage is nil")
	}
	if nil == m {
		return nil, errors.New("metrics instance is nil")
	}
	return &CommonBlockScanner{
		cfg:      cfg,
		logger:   log.Logger.With().Str("module", "commonblockscanner").Logger(),
		wg:       &sync.WaitGroup{},
		stopChan: make(chan struct{}),
		scanChan: make(chan int64, cfg.BlockScanProcessors),
		httpClient: &http.Client{
			Timeout: cfg.HttpRequestTimeout,
		},
		scannerStorage: scannerStorage,
		metrics:        m,
		previousBlock:  cfg.StartBlockHeight,
		errorCounter:   m.GetCounterVec(metrics.CommonBLockScannerError),
	}, nil
}

// GetHttpClient return the http client used internal to ourside world
// right now we need to use this for test
func (b *CommonBlockScanner) GetHttpClient() *http.Client {
	return b.httpClient
}

// GetMessages return the channel
func (b *CommonBlockScanner) GetMessages() <-chan int64 {
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
	if nil != err {
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
	if nil != err {
		b.errorCounter.WithLabelValues("fail_get_scan_pos", "").Inc()
		b.logger.Error().Err(err).Msgf("fail to get current block scan pos,we will start from %d", b.previousBlock)
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
			currentBlock, err := b.getRPCBlock(b.getBlockUrl())
			if nil != err {
				b.errorCounter.WithLabelValues("fail_get_block", "").Inc()
				b.logger.Error().Err(err).Msg("fail to get RPCBlock")
			}
			b.logger.Debug().Int64("current block height", currentBlock).Int64("we are at", b.previousBlock).Msg("get block height")
			// make sure we are one block behind , so when we process the block all the necessary data will be available
			currentBlock = currentBlock - 1
			if b.previousBlock >= currentBlock {
				// back off
				time.Sleep(b.cfg.BlockHeightDiscoverBackoff)
				continue
			}
			if currentBlock > b.previousBlock {
				// scan next block
				for idx := b.previousBlock; idx < currentBlock; idx++ {
					b.previousBlock++
					b.metrics.GetCounter(metrics.TotalBlockScanned).Inc()
					if err := b.scannerStorage.SetBlockScanStatus(b.previousBlock, NotStarted); err != nil {
						b.logger.Error().Err(err).Msg("fail to set block status")
						b.errorCounter.WithLabelValues("fail_set_block_status", strconv.FormatInt(b.previousBlock, 10)).Inc()
						return
					}
					select {
					case <-b.stopChan:
						return // need to bail
					case b.scanChan <- b.previousBlock:
					}
					b.metrics.GetCounter(metrics.CurrentPosition).Inc()
					if err := b.scannerStorage.SetScanPos(b.previousBlock); nil != err {
						b.errorCounter.WithLabelValues("fail_save_block_pos", strconv.FormatInt(b.previousBlock, 10)).Inc()
						b.logger.Error().Err(err).Msg("fail to save block scan pos")
						// alert!!
						return
					}
				}
			}
		}
	}
}

func (b *CommonBlockScanner) GetFromHttpWithRetry(url string) ([]byte, error) {
	backoffCtrl := backoff.NewExponentialBackOff()

	retry := 1
	for {
		res, err := b.getFromHttp(url)
		if nil == err {
			return res, nil
		}
		b.logger.Error().Err(err).Msgf("fail to get from %s try %d", url, retry)
		retry++
		backOffDuration := backoffCtrl.NextBackOff()
		if backOffDuration == backoff.Stop {
			return nil, errors.Wrapf(err, "fail to get from %s after maximum retry", url)
		}
		if retry >= b.cfg.MaxHttpRequestRetry {
			return nil, errors.Errorf("fail to get from %s after maximum retry(%d)", url, b.cfg.MaxHttpRequestRetry)
		}
		t := time.NewTicker(backOffDuration)
		select {
		case <-b.stopChan:
			return nil, err
		case <-t.C:
			t.Stop()
		}
	}
}

func (b *CommonBlockScanner) getFromHttp(url string) ([]byte, error) {
	b.logger.Debug().Str("url", url).Msg("http")
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if nil != err {
		b.errorCounter.WithLabelValues("fail_create_http_request", url).Inc()
		return nil, errors.Wrap(err, "fail to create http request")
	}
	resp, err := b.httpClient.Do(req)
	if nil != err {
		b.errorCounter.WithLabelValues("fail_send_http_request", url).Inc()
		return nil, errors.Wrapf(err, "fail to get from %s ", url)
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			b.logger.Error().Err(err).Msg("fail to close http response body.")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		b.errorCounter.WithLabelValues("unexpected_status_code", resp.Status).Inc()
		return nil, errors.Errorf("unexpected status code:%d from %s", resp.StatusCode, url)
	}
	return ioutil.ReadAll(resp.Body)
}

func (b *CommonBlockScanner) getBlockUrl() string {
	u, err := url.Parse(b.cfg.RPCHost)
	if err != nil {
		fmt.Printf("Failed to parse rpc host: %s\n", b.cfg.RPCHost)
	}
	if u == nil {
		requestUrl := url.URL{
			Scheme: "http",
			Host:   b.cfg.RPCHost,
			Path:   "block",
		}
		return requestUrl.String()
	}
	requestUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "block",
	}
	return requestUrl.String()
}

func (b *CommonBlockScanner) getRPCBlock(requestUrl string) (int64, error) {
	start := time.Now()
	defer func() {
		if err := recover(); nil != err {
			b.logger.Error().Msgf("fail to get RPCBlock:%s", err)
		}
		duration := time.Since(start)
		b.metrics.GetHistograms(metrics.BlockDiscoveryDuration).Observe(duration.Seconds())
	}()
	b.logger.Debug().Str("request_url", requestUrl).Msg("get_block")
	buf, err := b.GetFromHttpWithRetry(requestUrl)
	if nil != err {
		b.errorCounter.WithLabelValues("fail_get_block", requestUrl).Inc()
		return 0, errors.Wrap(err, "fail to get blocks")
	}
	var tx btypes.RPCBlock
	if err := json.Unmarshal(buf, &tx); nil != err {
		b.errorCounter.WithLabelValues("fail_unmarshal_block", requestUrl).Inc()
		return 0, errors.Wrap(err, "fail to unmarshal body to RPCBlock")
	}
	block := tx.Result.Block.Header.Height

	parsedBlock, err := strconv.ParseInt(block, 10, 64)
	if nil != err {
		b.errorCounter.WithLabelValues("fail_parse_block_height", block).Inc()
		return 0, errors.Wrap(err, "fail to convert block height to int")
	}
	return parsedBlock, nil
}

func (b *CommonBlockScanner) Stop() error {
	b.logger.Debug().Msg("receive stop request")
	defer b.logger.Debug().Msg("common block scanner stopped")
	close(b.stopChan)
	b.wg.Wait()
	return nil
}
