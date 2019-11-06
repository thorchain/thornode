package observer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/binance-chain/go-sdk/common/types"
	bmsg "github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/types/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/bepswap/thornode/common"

	btypes "gitlab.com/thorchain/bepswap/thornode/bifrost/binance/types"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/config"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/metrics"
	stypes "gitlab.com/thorchain/bepswap/thornode/bifrost/statechain/types"
)

// BinanceBlockScanner is to scan the blocks
type BinanceBlockScanner struct {
	cfg                config.BlockScannerConfiguration
	logger             zerolog.Logger
	wg                 *sync.WaitGroup
	stopChan           chan struct{}
	txInChan           chan stypes.TxIn
	db                 blockscanner.ScannerStorage
	commonBlockScanner *blockscanner.CommonBlockScanner
	m                  *metrics.Metrics
	errCounter         *prometheus.CounterVec
	addrVal            AddressValidator
}

// NewBinanceBlockScanner create a new instance of BlockScan
func NewBinanceBlockScanner(cfg config.BlockScannerConfiguration, scanStorage blockscanner.ScannerStorage, isTestNet bool, addrVal AddressValidator, m *metrics.Metrics) (*BinanceBlockScanner, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("rpc host is empty")
	}
	if nil == scanStorage {
		return nil, errors.New("scanStorage is nil")
	}
	if addrVal == nil {
		return nil, errors.New("pool address validator is nil")
	}
	if nil == m {
		return nil, errors.New("metrics is nil")
	}
	commonBlockScanner, err := blockscanner.NewCommonBlockScanner(cfg, scanStorage, m)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create common block scanner")
	}
	if isTestNet {
		types.Network = types.TestNetwork
	}
	return &BinanceBlockScanner{
		cfg:                cfg,
		addrVal:            addrVal,
		logger:             log.Logger.With().Str("module", "blockscanner").Logger(),
		wg:                 &sync.WaitGroup{},
		stopChan:           make(chan struct{}),
		txInChan:           make(chan stypes.TxIn),
		db:                 scanStorage,
		commonBlockScanner: commonBlockScanner,
		errCounter:         m.GetCounterVec(metrics.BinanceBlockScanError),
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
	strBlock := strconv.FormatInt(block, 10)
	if err := b.db.SetBlockScanStatus(block, blockscanner.Processing); nil != err {
		b.errCounter.WithLabelValues("fail_set_block_status", strBlock).Inc()
		return errors.Wrapf(err, "fail to set block scan status for block %d", block)
	}
	b.logger.Debug().Str("url", txSearchUrl).Int64("height", block).Msg("start search txs in block")
	buf, err := b.commonBlockScanner.GetFromHttpWithRetry(txSearchUrl)
	if nil != err {
		b.errCounter.WithLabelValues("fail_tx_search", strBlock).Inc()
		return errors.Wrap(err, "fail to send tx search request")
	}
	var query btypes.RPCTxSearch
	if err := json.Unmarshal(buf, &query); nil != err {
		b.errCounter.WithLabelValues("fail_unmarshal_tx_search", strBlock).Inc()
		return errors.Wrap(err, "fail to unmarshal RPCTxSearch")
	}

	b.logger.Info().Int64("block", block).Int("txs", len(query.Result.Txs)).Str("total", query.Result.TotalCount).Msg("txs")
	if len(query.Result.Txs) == 0 {
		b.m.GetCounter(metrics.BlockWithoutTx).Inc()
		b.logger.Debug().Int64("block", block).Msg("there are no txs in this block")
		return nil
	}
	// TODO implement pagination appropriately
	var txIn stypes.TxIn
	for _, txn := range query.Result.Txs {
		txItemIn, err := b.fromTxToTxIn(txn.Hash, txn.Height, txn.Tx) // b.getOneTxFromServer(txn.Hash, b.getSingleTxUrl(txn.Hash))
		if nil != err {
			b.errCounter.WithLabelValues("fail_get_tx", strBlock).Inc()
			b.logger.Error().Err(err).Str("hash", txn.Hash).Msg("fail to get one tx from server")
			// if we fail to get one tx hash from server, then we should bail, because we might miss tx
			// if we bail here, then we should retry later
			return errors.Wrap(err, "fail to get one tx from server")
		}
		if nil != txItemIn {
			txIn.TxArray = append(txIn.TxArray, *txItemIn)
			b.m.GetCounter(metrics.BlockWithTxIn).Inc()
			b.logger.Info().Str("hash", txn.Hash).Msg("we got one tx")
		}
	}
	if len(txIn.TxArray) == 0 {
		b.m.GetCounter(metrics.BlockNoTxIn).Inc()
		b.logger.Debug().Int64("block", block).Msg("no tx need to be processed in this block")
		return nil
	}

	txIn.BlockHeight = strconv.FormatInt(block, 10)
	txIn.Count = strconv.Itoa(len(txIn.TxArray))
	txIn.Chain = common.BNBChain
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
			b.logger.Info().Int64("block", block).Msg("processing block")
			if err := b.searchTxInABlockFromServer(block, b.getTxSearchUrl(block, 1, 100)); nil != err {
				if errStatus := b.db.SetBlockScanStatus(block, blockscanner.Failed); nil != errStatus {
					b.errCounter.WithLabelValues("fail_set_block_status", "").Inc()
					b.logger.Error().Err(err).Int64("height", block).Msg("fail to set block to fail status")
				}
				b.errCounter.WithLabelValues("fail_search_block", "").Inc()
				b.logger.Error().Err(err).Int64("height", block).Msg("fail to search tx in block")
				// we will have a retry go routine to check it.
				continue
			}
			// set a block as success
			if err := b.db.RemoveBlockStatus(block); nil != err {
				b.errCounter.WithLabelValues("fail_remove_block_status", "").Inc()
				b.logger.Error().Err(err).Int64("block", block).Msg("fail to remove block status from data store, thus block will be re processed")
			}
		}
	}
}

// Check if memo is for registering a Yggdrasil pool
func (b *BinanceBlockScanner) isRegisterYggdrasil(addr, memo string) bool {
	return b.isAddrWithMemo(addr, memo, "yggdrasil+")
}

// Check if memo is for de registering a Yggdrasil pool
func (b *BinanceBlockScanner) isDeregisterYggdrasil(addr, memo string) bool {
	return b.isAddrWithMemo(addr, memo, "yggdrasil-")
}

// Check if we have an outbound yggdrasil transaction
func (b *BinanceBlockScanner) isYggdrasil(addr string) bool {
	ok, _ := b.addrVal.IsValidPoolAddress(addr, common.BNBChain)
	return ok
}

func (b *BinanceBlockScanner) isOutboundMsg(addr, memo string) bool {
	return b.isAddrWithMemo(addr, memo, "outbound")
}

func (b *BinanceBlockScanner) isPoolAck(addr, memo string) bool {
	return b.isAddrWithMemo(addr, memo, "ack")
}

func (b *BinanceBlockScanner) isNextPoolMsg(addr, memo string, coins types.Coins) bool {
	if ok := b.isAddrWithMemo(addr, memo, "nextpool"); !ok {
		return false
	}

	nextPoolCoin := types.Coin{
		Denom:  common.BNBTicker.String(),
		Amount: 37501,
	}
	hasCorrectCoin := false
	for _, c := range coins {
		if c.SameDenomAs(nextPoolCoin) && c.Amount == nextPoolCoin.Amount {
			hasCorrectCoin = true
			break
		}
	}
	if hasCorrectCoin {
		return true
	}
	return false
}

func (b *BinanceBlockScanner) isAddrWithMemo(addr, memo, targetMemo string) bool {
	match, _ := b.addrVal.IsValidPoolAddress(addr, common.BNBChain)
	if !match {
		return false
	}
	lowerMemo := strings.ToLower(memo)
	if strings.HasPrefix(lowerMemo, targetMemo) {
		return true
	}
	return false
}

func (b *BinanceBlockScanner) getCoinsForTxIn(coins types.Coins) (common.Coins, error) {
	cc := make(common.Coins, len(coins))
	for i, c := range coins {
		asset, err := common.NewAsset(fmt.Sprintf("BNB.%s", c.Denom))
		if nil != err {
			b.errCounter.WithLabelValues("fail_create_ticker", c.Denom).Inc()
			return nil, errors.Wrapf(err, "fail to create asset, %s is not valid", c.Denom)
		}
		amt := sdk.NewUint(uint64(c.Amount))
		cc[i] = common.NewCoin(asset, amt)
	}
	return cc, nil
}

func (b *BinanceBlockScanner) fromTxToTxIn(hash, height, encodedTx string) (*stypes.TxInItem, error) {
	if len(encodedTx) == 0 {
		return nil, errors.New("tx is empty")
	}
	buf, err := base64.StdEncoding.DecodeString(encodedTx)
	if nil != err {
		b.errCounter.WithLabelValues("fail_decode_tx", hash).Inc()
		return nil, errors.Wrap(err, "fail to decode tx")
	}
	var t tx.StdTx
	if err := tx.Cdc.UnmarshalBinaryLengthPrefixed(buf, &t); nil != err {
		b.errCounter.WithLabelValues("fail_unmarshal_tx", hash).Inc()
		return nil, errors.Wrap(err, "fail to unmarshal tx.StdTx")
	}

	return b.fromStdTx(hash, t)
}

// fromStdTx - process a stdTx
func (b *BinanceBlockScanner) fromStdTx(hash string, stdTx tx.StdTx) (*stypes.TxInItem, error) {
	var err error
	txInItem := stypes.TxInItem{
		Tx: hash,
	}
	// TODO: it is possible to have multiple `SendMsg` in a single stdTx, which
	// we are currently not accounting for. It is also possible to have
	// multiple inputs/outputs within a single stdTx, which we are not yet
	// accounting for.
	for _, msg := range stdTx.Msgs {
		switch sendMsg := msg.(type) {
		case bmsg.SendMsg:
			txInItem.Memo = stdTx.Memo
			sender := sendMsg.Inputs[0]
			receiver := sendMsg.Outputs[0]
			txInItem.Sender = sender.Address.String()
			txInItem.To = receiver.Address.String()
			txInItem.Coins, err = b.getCoinsForTxIn(receiver.Coins)
			if nil != err {
				return nil, errors.Wrap(err, "fail to convert coins")
			}

			// check if the from address is a valid pool
			if ok, cpi := b.addrVal.IsValidPoolAddress(txInItem.Sender, common.BNBChain); ok {
				txInItem.ObservedPoolAddress = cpi.PubKey.String()
			}
			// check if the to address is a valid pool address
			if ok, cpi := b.addrVal.IsValidPoolAddress(txInItem.To, common.BNBChain); ok {
				txInItem.ObservedPoolAddress = cpi.PubKey.String()
			}

			// Check if our pool is registering a new yggdrasil pool. Ie
			// sending the staked assets to the user
			if ok := b.isRegisterYggdrasil(txInItem.Sender, txInItem.Memo); ok {
				b.logger.Debug().Str("memo", txInItem.Memo).Msg("yggdrasil+")

				// **IMPORTANT** If this fails, we won't monitor the address and could lose funds!
				var pk common.PubKey
				if strings.HasPrefix(txInItem.To, "tbnb") {
					pk, _ = common.NewPubKeyFromBech32(txInItem.To, "tbnb")
				} else {
					pk, _ = common.NewPubKeyFromBech32(txInItem.To, "bnb")
				}
				b.addrVal.AddPubKey(pk)
				return &txInItem, nil
			}

			// Check if out pool is de registering a yggdrasil pool. Ie sending
			// the bond back to the user
			if ok := b.isDeregisterYggdrasil(txInItem.Sender, txInItem.Memo); ok {
				b.logger.Debug().Str("memo", txInItem.Memo).Msg("yggdrasil-")

				// **IMPORTANT** If this fails, we may slash a yggdrasil pool inappropriately
				var pk common.PubKey
				if strings.HasPrefix(txInItem.To, "tbnb") {
					pk, _ = common.NewPubKeyFromBech32(txInItem.To, "tbnb")
				} else {
					pk, _ = common.NewPubKeyFromBech32(txInItem.To, "bnb")
				}
				b.addrVal.RemovePubKey(pk)

				return &txInItem, nil
			}

			// Check if we are sending from a yggdrasil address
			if ok := b.isYggdrasil(txInItem.Sender); ok {
				b.logger.Debug().Str("assets sent from yggdrasil pool", txInItem.Memo).Msg("fill order")
				return &txInItem, nil
			}

			// Check if we are sending to a yggdrasil address
			if ok := b.isYggdrasil(txInItem.To); ok {
				b.logger.Debug().Str("assets to yggdrasil pool", txInItem.Memo).Msg("refill")
				return &txInItem, nil
			}

			// outbound message from pool, when it is outbound, it does not matter how much coins we send to customer for now
			if ok := b.isOutboundMsg(txInItem.Sender, txInItem.Memo); ok {
				b.logger.Debug().Str("memo", txInItem.Memo).Msg("outbound")
				return &txInItem, nil
			}

			// Check that if we've gotten an ack from the next pool
			if ok := b.isPoolAck(txInItem.Sender, txInItem.Memo); ok {
				b.logger.Debug().Str("memo", txInItem.Memo).Msg("ack")
				// We've added a new pool, we should refresh our list of pools
				// from thorchain
				b.addrVal.UpdatePoolAddresses()
				return &txInItem, nil
			}

			if ok := b.isNextPoolMsg(txInItem.Sender, txInItem.Memo, sender.Coins); ok {
				b.logger.Debug().Str("memo", txInItem.Memo).Msg("nextpool")
				return &txInItem, nil
			}
		default:
			continue
		}
	}
	return nil, nil
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
