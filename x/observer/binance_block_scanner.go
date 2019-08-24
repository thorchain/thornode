package observer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/thorchain/bepswap/common"

	"gitlab.com/thorchain/bepswap/observe/config"
	btypes "gitlab.com/thorchain/bepswap/observe/x/binance/types"
	"gitlab.com/thorchain/bepswap/observe/x/blockscanner"
	stypes "gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

// BinanceBlockScanner is to scan the blocks
type BinanceBlockScanner struct {
	cfg                config.BlockScannerConfiguration
	dexHost            string
	poolAddress        common.BnbAddress
	logger             zerolog.Logger
	wg                 *sync.WaitGroup
	stopChan           chan struct{}
	txInChan           chan stypes.TxIn
	db                 blockscanner.ScannerStorage
	commonBlockScanner *blockscanner.CommonBlockScanner
}

// NewBinanceBlockScanner create a new instance of BlockScan
func NewBinanceBlockScanner(cfg config.BlockScannerConfiguration, scanStorage blockscanner.ScannerStorage, DEXHost string, poolAddress common.BnbAddress) (*BinanceBlockScanner, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("rpc host is empty")
	}
	if nil == scanStorage {
		return nil, errors.New("scanStorage is nil")
	}
	if len(DEXHost) == 0 {
		return nil, errors.New("DEXHost is empty")
	}
	if poolAddress.IsEmpty() {
		return nil, errors.New("pool address is empty")
	}
	commonBlockScanner, err := blockscanner.NewCommonBlockScanner(cfg, scanStorage)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create common block scanner")
	}
	return &BinanceBlockScanner{
		cfg:                cfg,
		dexHost:            DEXHost,
		poolAddress:        poolAddress,
		logger:             log.Logger.With().Str("module", "blockscanner").Logger(),
		wg:                 &sync.WaitGroup{},
		stopChan:           make(chan struct{}),
		txInChan:           make(chan stypes.TxIn),
		db:                 scanStorage,
		commonBlockScanner: commonBlockScanner,
	}, nil
}

// GetMessages return the channel
func (b *BinanceBlockScanner) GetMessages() <-chan stypes.TxIn {
	return b.txInChan
}

// Start block scanner
func (b *BinanceBlockScanner) Start() {
	for idx := 1; idx <= b.cfg.BlockScanProcessors; idx++ {
		b.wg.Add(1)
		go b.searchTxInABlock(idx)
	}
	b.commonBlockScanner.Start()
}

// need to process multiple pages
func (b *BinanceBlockScanner) getTxSearchUrl(block int64, currentPage, numberPerPage int64) string {
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

func (b *BinanceBlockScanner) searchTxInABlockFromServer(block int64, txSearchUrl string) error {
	if err := b.db.SetBlockScanStatus(block, blockscanner.Processing); nil != err {
		return errors.Wrapf(err, "fail to set block scan status for block %d", block)
	}
	b.logger.Debug().Str("url", txSearchUrl).Int64("height", block).Msg("start search txs in block")
	buf, err := b.commonBlockScanner.GetFromHttpWithRetry(txSearchUrl)
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
			b.logger.Info().Str("hash", txn.Hash).Msg("we got one tx")
		}
	}
	if len(txIn.TxArray) == 0 {
		b.logger.Debug().Int64("block", block).Msg("no tx need to be processed in this block")
		return nil
	}

	txIn.BlockHeight = strconv.FormatInt(block, 10)
	txIn.Count = strconv.Itoa(len(txIn.TxArray))
	b.txInChan <- txIn
	return nil
}

func (b *BinanceBlockScanner) searchTxInABlock(idx int) {
	b.logger.Debug().Int("idx", idx).Msg("start searching tx in a block")
	defer b.logger.Debug().Int("idx", idx).Msg("stop searching tx in a block")
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
			if err := b.searchTxInABlockFromServer(block, b.getTxSearchUrl(block, 1, 100)); nil != err {
				if errStatus := b.db.SetBlockScanStatus(block, blockscanner.Failed); nil != errStatus {
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
func (b *BinanceBlockScanner) getSingleTxUrl(txHash string) string {
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
func (b *BinanceBlockScanner) getOneTxFromServer(txhash string, requestUrl string) (*stypes.TxInItem, error) {
	b.logger.Debug().Str("txhash", txhash).Str("requesturi", requestUrl).Msg("get one tx from server")
	buf, err := b.commonBlockScanner.GetFromHttpWithRetry(requestUrl)
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
func (b *BinanceBlockScanner) fromApiTxToTxInItem(txInput btypes.ApiTx) (*stypes.TxInItem, error) {
	var existTx bool
	txInItem := stypes.TxInItem{
		Tx: txInput.Hash,
	}
	for _, msg := range txInput.Tx.Value.Msg {
		if len(msg.Value.Inputs) == 0 {
			continue
		}
		sender := msg.Value.Inputs[0]
		for _, output := range msg.Value.Outputs {
			if !strings.EqualFold(output.Address, b.poolAddress.String()) {
				continue
			}
			existTx = true
			txInItem.Memo = txInput.Tx.Value.Memo
			txInItem.Sender = sender.Address
			for _, coin := range output.Coins {
				ticker, err := common.NewTicker(coin.Denom)
				if nil != err {
					return nil, errors.Wrapf(err, "fail to create ticker, %s is not valid", coin.Denom)
				}
				amt, err := common.NewAmount(coin.Amount)
				if nil != err {
					return nil, errors.Wrapf(err, "fail to parse coin amount,%s is not valid", coin.Amount)
				}
				txInItem.Coins = append(txInItem.Coins, common.NewCoin(ticker, common.NewAmountFromFloat(amt.Float64()/100000000)))
			}

		}
	}
	if !existTx {
		b.logger.Debug().Str("hash", txInput.Hash).Str("height", txInput.Height).Msg("didn't find any tx that we should process")
		return nil, nil
	}
	return &txInItem, nil
}

func (b *BinanceBlockScanner) Stop() error {
	b.logger.Debug().Msg("receive stop request")
	defer b.logger.Debug().Msg("block scanner stopped")
	if err := b.commonBlockScanner.Stop(); nil != err {
		b.logger.Error().Err(err).Msg("fail to stop common block scanner")
	}
	close(b.stopChan)
	b.wg.Wait()

	return nil
}
