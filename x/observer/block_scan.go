package observer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"gitlab.com/thorchain/bepswap/common"

	"gitlab.com/thorchain/bepswap/observe/config"
	btypes "gitlab.com/thorchain/bepswap/observe/x/binance/types"
	stypes "gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

// BlockScanStatus
type BlockScanStatus byte

const (
	Processing BlockScanStatus = iota
	Failed
	Finished
	NotStarted
)

// BlockScanner is to scan the blocks
type BlockScanner struct {
	cfg           config.BlockScannerConfiguration
	dexHost       string
	poolAddress   common.BnbAddress
	logger        zerolog.Logger
	wg            *sync.WaitGroup
	stopChan      chan struct{}
	txInChan      chan stypes.TxIn
	scanChan      chan int64
	httpClient    *fasthttp.Client
	db            ScannerStorage
	previousBlock int64
}

// NewBlockScanner create a new instance of BlockScan
func NewBlockScanner(cfg config.BlockScannerConfiguration, scanStorage ScannerStorage, DEXHost string, poolAddress common.BnbAddress) (*BlockScanner, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("rpc host is empty")
	}
	if nil == scanStorage {
		return nil, errors.New("scanStorage is nill")
	}
	if len(DEXHost) == 0 {
		return nil, errors.New("DEXHost is empty")
	}
	if poolAddress.IsEmpty() {
		return nil, errors.New("pool address is empty")
	}

	return &BlockScanner{
		cfg:         cfg,
		dexHost:     DEXHost,
		poolAddress: poolAddress,
		logger:      log.Logger.With().Str("module", "blockscanner").Logger(),
		wg:          &sync.WaitGroup{},
		stopChan:    make(chan struct{}),
		txInChan:    make(chan stypes.TxIn),
		scanChan:    make(chan int64, cfg.BlockScanProcessors),
		httpClient: &fasthttp.Client{
			ReadTimeout:  cfg.HttpRequestReadTimeout,
			WriteTimeout: cfg.HttpRequestWriteTimeout,
		},
		db:            scanStorage,
		previousBlock: cfg.StartBlockHeight,
	}, nil
}

// GetMessages return the channel
func (b *BlockScanner) GetMessages() <-chan stypes.TxIn {
	return b.txInChan
}

// Start block scanner
func (b *BlockScanner) Start() {
	for idx := 1; idx <= b.cfg.BlockScanProcessors; idx++ {
		b.wg.Add(1)
		go b.searchTxInABlock(idx)
	}

	b.wg.Add(1)
	go b.scanBlocks()
	go b.retryFailedBlocks()
}

// retryFailedBlocks , if somehow we failed to process a block , it will be retried
func (b *BlockScanner) retryFailedBlocks() {
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
func (b *BlockScanner) retryBlocks(failedonly bool) {
	// start up to grab those blocks that we didn't finished
	blocks, err := b.db.GetBlocksForRetry(failedonly)
	if nil != err {
		b.logger.Error().Err(err).Msg("fail to get blocks for retry")
	}
	b.logger.Debug().Msgf("find %q blocks need to retry", blocks)
	for _, item := range blocks {
		select {
		case <-b.stopChan:
			return // need to bail
		case b.scanChan <- item:
		}
	}
}

// scanBlocks
func (b *BlockScanner) scanBlocks() {
	b.logger.Debug().Msg("start to scan blocks")
	defer b.logger.Debug().Msg("stop scan blocks")
	defer b.wg.Done()
	currentPos, err := b.db.GetScanPos()
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
			b.logger.Debug().Int64("current block height", currentBlock).Msg("get block height")
			if b.previousBlock == currentBlock {
				// back off
				time.Sleep(b.cfg.BlockHeightDiscoverBackoff)
				continue
			}
			if currentBlock > b.previousBlock {
				// scan next block
				for idx := b.previousBlock; idx <= currentBlock; idx++ {
					if err := b.db.SetBlockScanStatus(b.previousBlock, NotStarted); err != nil {
						b.logger.Error().Err(err).Msg("fail to set block status")
						// alert!!
						// TODO what should we do if that happen
						return
					}
					select {
					case <-b.stopChan:
						return // need to bail
					case b.scanChan <- b.previousBlock:
					}
					b.previousBlock++
					if err := b.db.SetScanPos(b.previousBlock); nil != err {
						b.logger.Error().Err(err).Msg("fail to save block scan pos")
						// alert!!
						return
					}
				}
			}
		}
	}
}

func (b *BlockScanner) getBlockUrl() string {
	requestUrl := url.URL{
		Scheme: "https",
		Host:   b.cfg.RPCHost,
		Path:   "block",
	}
	return requestUrl.String()
}
func (b *BlockScanner) getRPCBlock(requestUrl string) (int64, error) {
	buf, err := b.getFromHttpWithRetry(requestUrl)
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

// need to process multiple pages
func (b *BlockScanner) getTxSearchUrl(block int64, currentPage, numberPerPage int64) string {
	uri := url.URL{
		Scheme: "https",
		Host:   b.cfg.RPCHost,
		Path:   "tx_search",
	}
	q := uri.Query()
	q.Set("query", fmt.Sprintf("\"tx.height=%d\"", block))
	q.Set("prove", "true")
	q.Set("page", strconv.FormatInt(currentPage, 10))
	q.Set("per_page", strconv.FormatInt(numberPerPage, 10))
	uri.RawQuery = q.Encode()
	return uri.String()
}

func (b *BlockScanner) getFromHttpWithRetry(url string) ([]byte, error) {
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

func (b *BlockScanner) getFromHttp(url string) ([]byte, error) {
	b.logger.Debug().Str("url", url).Msg("http")
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	if err := b.httpClient.Do(req, resp); nil != err {
		return nil, errors.Wrapf(err, "fail to get from %s ", url)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("unexpected status code:%d from %s", resp.StatusCode(), url)
	}
	return resp.Body(), nil
}

func (b *BlockScanner) searchTxInABlockFromServer(block int64, txSearchUrl string) error {
	if err := b.db.SetBlockScanStatus(block, Processing); nil != err {
		return errors.Wrapf(err, "fail to set block scan status for block %d", block)
	}
	b.logger.Debug().Str("url", txSearchUrl).Int64("height", block).Msg("start search txs in block")
	buf, err := b.getFromHttpWithRetry(txSearchUrl)
	if nil != err {
		return errors.Wrap(err, "fail to send tx search request")
	}
	var query btypes.RPCTxSearch
	if err := json.Unmarshal(buf, &query); nil != err {
		return errors.Wrap(err, "fail to unmarshal RPCTxSearch")
	}
	b.logger.Debug().Int("txs", len(query.Result.Txs)).Str("total", query.Result.TotalCount).Msg("txs")
	if len(query.Result.Txs) == 0 {
		b.logger.Debug().Int64("block", block).Msg("there are no txs in this block")
		return nil
	}
	// TODO implement pagination appropriately
	var txIn stypes.TxIn
	for _, txn := range query.Result.Txs {
		txItemIn, err := b.getOneTxFromServer(txn.Hash, b.getSingleTxUrl(txn.Hash))
		if nil != err {
			b.logger.Error().Err(err).Str("hash", txn.Hash).Msg("fail to get one tx from server")
			// if we fail to get one tx hash from server, then we should bail, because we might miss tx
			// if we bail here, then we should retry later
			return errors.Wrap(err, "fail to get one tx from server")
		}
		if nil != txItemIn {
			txIn.TxArray = append(txIn.TxArray, *txItemIn)
		}
	}

	txIn.BlockHeight = strconv.FormatInt(block, 10)
	txIn.Count = strconv.Itoa(len(txIn.TxArray))
	b.txInChan <- txIn
	return nil
}

func (b *BlockScanner) searchTxInABlock(idx int) {
	b.logger.Debug().Int("idx", idx).Msg("start searching tx in a block")
	defer b.logger.Debug().Int("idx", idx).Msg("stop searching tx in a block")
	defer b.wg.Done()

	for {
		select {
		case <-b.stopChan: // time to get out
			return
		case block, more := <-b.scanChan:
			if !more {
				return
			}
			b.logger.Debug().Int64("block", block).Msg("processing block")
			if err := b.searchTxInABlockFromServer(block, b.getTxSearchUrl(block, 1, 100)); nil != err {
				if errStatus := b.db.SetBlockScanStatus(block, Failed); nil != errStatus {
					b.logger.Error().Err(err).Int64("height", block).Msg("fail to set block to fail status")
				}
				b.logger.Error().Err(err).Int64("height", block).Msg("fail to search tx in block")
				// we will have a retry go routine to check it.
				continue
			}
			// set a block as success
			if err := b.db.RemoveBlockStatus(block); nil != err {
				b.logger.Error().Err(err).Int64("block", block).Msg("fail to remove block status from data store, thus block will be re processed")
			}
		}
	}
}
func (b *BlockScanner) getSingleTxUrl(txHash string) string {
	uri := url.URL{
		Scheme: "https",
		Host:   b.dexHost,
		Path:   fmt.Sprintf("api/v1/tx/%s", txHash),
	}
	query := uri.Query()
	query.Set("format", "json")
	uri.RawQuery = query.Encode()
	return uri.String()
}

// getOneTxFromServer
func (b *BlockScanner) getOneTxFromServer(txhash string, requestUrl string) (*stypes.TxInItem, error) {
	b.logger.Debug().Str("txhash", txhash).Str("requesturi", requestUrl).Msg("get one tx from server")
	buf, err := b.getFromHttpWithRetry(requestUrl)
	if nil != err {
		return nil, errors.Wrap(err, "fail to get query tx detail")
	}
	var tx btypes.ApiTx
	if err := json.Unmarshal(buf, &tx); nil != err {
		return nil, errors.Wrap(err, "fail to unmarshal tx detail")
	}
	return b.fromApiTxToTxInItem(tx)
}

// fromApiTxToTxInItem convert ApiTx to txinitem
func (b *BlockScanner) fromApiTxToTxInItem(txInput btypes.ApiTx) (*stypes.TxInItem, error) {
	var existTx bool
	txInItem := stypes.TxInItem{
		Tx: txInput.Hash,
	}

	for _, msg := range txInput.Tx.Value.Msg {
		// TODO check msg type , we probably only care about transfer
		for j, output := range msg.Value.Outputs {
			if !strings.EqualFold(output.Address, b.poolAddress.String()) {
				continue
			}
			existTx = true
			sender := msg.Value.Inputs[j]
			txInItem.Memo = txInput.Tx.Value.Memo
			txInItem.Sender = sender.Address
			// TODO shouldn be output
			for _, coin := range sender.Coins {
				ticker, err := common.NewTicker(coin.Denom)
				if nil != err {
					return nil, errors.Wrapf(err, "fail to create ticker, %s is not valid", coin.Denom)
				}
				amt, err := common.NewAmount(coin.Amount)
				if nil != err {
					return nil, errors.Wrapf(err, "fail to parse coin amount,%s is not valid", coin.Amount)
				}
				txInItem.Coins = append(txInItem.Coins, common.NewCoin(ticker, amt))
			}

		}
	}
	if !existTx {
		b.logger.Debug().Str("hash", txInput.Hash).Str("height", txInput.Height).Msg("didn't find any tx that we should process")
		return nil, nil
	}
	return &txInItem, nil
}

func (b *BlockScanner) Stop() error {
	b.logger.Debug().Msg("receive stop request")
	defer b.logger.Debug().Msg("block scanner stopped")
	close(b.stopChan)
	b.wg.Wait()
	if err := b.db.Close(); nil != err {
		return errors.Wrap(err, "fail to stop level db")
	}

	return nil
}
