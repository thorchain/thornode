package blockscanner

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	"gitlab.com/thorchain/bepswap/observe/config"
	btypes "gitlab.com/thorchain/bepswap/observe/x/binance/types"
)

// CommonBlockScanner is used to discover block height
// since both binance and statechain use cosmos, so this part logic should be the same
type CommonBlockScanner struct {
	cfg            config.BlockScannerConfiguration
	logger         zerolog.Logger
	wg             *sync.WaitGroup
	scanChan       chan int64
	stopChan       chan struct{}
	httpClient     *fasthttp.Client
	scannerStorage ScannerStorage
	previousBlock  int64
}

// NewCommonBlockScanner create a new instance of CommonBlockScanner
func NewCommonBlockScanner(cfg config.BlockScannerConfiguration, scannerStorage ScannerStorage) (*CommonBlockScanner, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("host is empty")
	}
	if nil == scannerStorage {
		return nil, errors.New("scannerStorage is nil")
	}
	return &CommonBlockScanner{
		cfg:      cfg,
		logger:   log.Logger.With().Str("module", "commonblockscanner").Logger(),
		wg:       &sync.WaitGroup{},
		stopChan: make(chan struct{}),
		scanChan: make(chan int64, cfg.BlockScanProcessors),
		httpClient: &fasthttp.Client{
			ReadTimeout:  cfg.HttpRequestReadTimeout,
			WriteTimeout: cfg.HttpRequestWriteTimeout,
		},
		scannerStorage: scannerStorage,
		previousBlock:  cfg.StartBlockHeight,
	}, nil
}

// GetHttpClient return the http client used internal to ourside world
// right now we need to use this for test
func (b *CommonBlockScanner) GetHttpClient() *fasthttp.Client {
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
		b.logger.Error().Err(err).Msg("fail to get blocks for retry")
	}
	b.logger.Debug().Msgf("find %v blocks need to retry", blocks)
	for _, item := range blocks {
		select {
		case <-b.stopChan:
			return // need to bail
		case b.scanChan <- item:
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
		b.logger.Error().Err(err).Msgf("fail to get current block scan pos,we will start from %d", b.previousBlock)
	} else {
		b.previousBlock = currentPos
	}
	// start up to grab those blocks that we didn't finished
	b.retryBlocks(false)
	for {
		select {
		case <-b.stopChan:
			return
		default:
			currentBlock, err := b.getRPCBlock(b.getBlockUrl())
			if nil != err {
				b.logger.Error().Err(err).Msg("fail to get RPCBlock")
			}
			b.logger.Info().Int64("current block height", currentBlock).Int64("we are at", b.previousBlock).Msg("get block height")
			if b.previousBlock >= currentBlock {
				// back off
				time.Sleep(b.cfg.BlockHeightDiscoverBackoff)
				continue
			}
			if currentBlock > b.previousBlock {
				// scan next block
				for idx := b.previousBlock; idx < currentBlock; idx++ {
					b.previousBlock++
					if err := b.scannerStorage.SetBlockScanStatus(b.previousBlock, NotStarted); err != nil {
						b.logger.Error().Err(err).Msg("fail to set block status")
						// TODO alert here , because we stop scanning blocks
						return
					}
					select {
					case <-b.stopChan:
						return // need to bail
					case b.scanChan <- b.previousBlock:
					}
					if err := b.scannerStorage.SetScanPos(b.previousBlock); nil != err {
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
	req := fasthttp.AcquireRequest()
	req.Reset()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	resp.Reset()
	defer fasthttp.ReleaseResponse(resp)
	if err := b.httpClient.Do(req, resp); nil != err {
		return nil, errors.Wrapf(err, "fail to get from %s ", url)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("unexpected status code:%d from %s", resp.StatusCode(), url)
	}
	return resp.Body(), nil
}

func (b *CommonBlockScanner) getBlockUrl() string {
	requestUrl := url.URL{
		Scheme: b.cfg.Scheme,
		Host:   b.cfg.RPCHost,
		Path:   "block",
	}
	return requestUrl.String()
}
func (b *CommonBlockScanner) getRPCBlock(requestUrl string) (int64, error) {
	defer func() {
		if err := recover(); nil != err {
			b.logger.Error().Msgf("fail to get RPCBlock:%s", err)
		}
	}()
	buf, err := b.GetFromHttpWithRetry(requestUrl)
	if nil != err {
		return 0, errors.Wrap(err, "fail to get blocks")
	}
	var tx btypes.RPCBlock
	if err := json.Unmarshal(buf, &tx); nil != err {
		return 0, errors.Wrap(err, "fail to unmarshal body to RPCBlock")
	}
	block := tx.Result.Block.Header.Height

	parsedBlock, err := strconv.ParseInt(block, 10, 64)
	if nil != err {
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
