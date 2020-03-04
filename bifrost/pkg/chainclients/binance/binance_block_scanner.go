package binance

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
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

	"gitlab.com/thorchain/thornode/common"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

// BinanceBlockScanner is to scan the blocks
type BinanceBlockScanner struct {
	cfg                config.BlockScannerConfiguration
	logger             zerolog.Logger
	wg                 *sync.WaitGroup
	stopChan           chan struct{}
	db                 blockscanner.ScannerStorage
	commonBlockScanner *blockscanner.CommonBlockScanner
	m                  *metrics.Metrics
	errCounter         *prometheus.CounterVec
	pubkeyMgr          pubkeymanager.PubKeyValidator
	globalTxsQueue     chan stypes.TxIn
	rpcHost            string
}

// NewBinanceBlockScanner create a new instance of BlockScan
func NewBinanceBlockScanner(cfg config.BlockScannerConfiguration, startBlockHeight int64, scanStorage blockscanner.ScannerStorage, isTestNet bool, pkmgr pubkeymanager.PubKeyValidator, m *metrics.Metrics) (*BinanceBlockScanner, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("rpc host is empty")
	}

	rpcHost := cfg.RPCHost
	if !strings.HasPrefix(rpcHost, "http") {
		rpcHost = fmt.Sprintf("http://%s", rpcHost)
	}

	if scanStorage == nil {
		return nil, errors.New("scanStorage is nil")
	}
	if pkmgr == nil {
		return nil, errors.New("pubkey validator is nil")
	}
	if m == nil {
		return nil, errors.New("metrics is nil")
	}
	commonBlockScanner, err := blockscanner.NewCommonBlockScanner(cfg, startBlockHeight, scanStorage, m)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create common block scanner")
	}
	if isTestNet {
		types.Network = types.TestNetwork
	} else {
		types.Network = types.ProdNetwork
	}
	return &BinanceBlockScanner{
		cfg:                cfg,
		pubkeyMgr:          pkmgr,
		logger:             log.Logger.With().Str("module", "blockscanner").Logger(),
		wg:                 &sync.WaitGroup{},
		stopChan:           make(chan struct{}),
		db:                 scanStorage,
		commonBlockScanner: commonBlockScanner,
		errCounter:         m.GetCounterVec(metrics.BlockScanError(common.BNBChain)),
		rpcHost:            rpcHost,
	}, nil
}

// Start block scanner
func (b *BinanceBlockScanner) Start(globalTxsQueue chan stypes.TxIn) {
	b.globalTxsQueue = globalTxsQueue
	for idx := 1; idx <= b.cfg.BlockScanProcessors; idx++ {
		b.wg.Add(1)
		go b.searchTxInABlock(idx)
	}
	b.commonBlockScanner.Start()
}

func (b *BinanceBlockScanner) processBlock(block blockscanner.Block) error {
	strBlock := strconv.FormatInt(block.Height, 10)
	if err := b.db.SetBlockScanStatus(block, blockscanner.Processing); err != nil {
		b.errCounter.WithLabelValues("fail_set_block_status", strBlock).Inc()
		return errors.Wrapf(err, "fail to set block scan status for block %d", block.Height)
	}

	b.logger.Debug().Int64("block", block.Height).Int("txs", len(block.Txs)).Msg("txs")
	if len(block.Txs) == 0 {
		b.m.GetCounter(metrics.BlockWithoutTx("BNB")).Inc()
		b.logger.Debug().Int64("block", block.Height).Msg("there are no txs in this block")
		return nil
	}

	// TODO implement pagination appropriately
	var txIn stypes.TxIn
	for _, txn := range block.Txs {
		hash := fmt.Sprintf("%X", sha256.Sum256([]byte(txn)))
		txItemIns, err := b.fromTxToTxIn(hash, txn)
		if err != nil {
			b.errCounter.WithLabelValues("fail_get_tx", strBlock).Inc()
			b.logger.Error().Err(err).Str("hash", hash).Msg("fail to get one tx from server")
			// if THORNode fail to get one tx hash from server, then THORNode should bail, because THORNode might miss tx
			// if THORNode bail here, then THORNode should retry later
			return errors.Wrap(err, "fail to get one tx from server")
		}
		if len(txItemIns) > 0 {
			txIn.TxArray = append(txIn.TxArray, txItemIns...)
			b.m.GetCounter(metrics.BlockWithTxIn("BNB")).Inc()
			b.logger.Info().Str("hash", hash).Msg("THORNode got one tx")
		}
	}
	if len(txIn.TxArray) == 0 {
		b.m.GetCounter(metrics.BlockNoTxIn("BNB")).Inc()
		b.logger.Debug().Int64("block", block.Height).Msg("no tx need to be processed in this block")
		return nil
	}

	txIn.BlockHeight = strconv.FormatInt(block.Height, 10)
	txIn.Count = strconv.Itoa(len(txIn.TxArray))
	txIn.Chain = common.BNBChain
	b.globalTxsQueue <- txIn
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
			b.logger.Debug().Int64("block", block.Height).Msg("processing block")
			if err := b.processBlock(block); err != nil {
				if errStatus := b.db.SetBlockScanStatus(block, blockscanner.Failed); errStatus != nil {
					b.errCounter.WithLabelValues("fail_set_block_status", "").Inc()
					b.logger.Error().Err(err).Int64("height", block.Height).Msg("fail to set block to fail status")
				}
				b.errCounter.WithLabelValues("fail_search_block", "").Inc()
				b.logger.Error().Err(err).Int64("height", block.Height).Msg("fail to search tx in block")
				// THORNode will have a retry go routine to check it.
				continue
			}
			// set a block as success
			if err := b.db.RemoveBlockStatus(block.Height); err != nil {
				b.errCounter.WithLabelValues("fail_remove_block_status", "").Inc()
				b.logger.Error().Err(err).Int64("block", block.Height).Msg("fail to remove block status from data store, thus block will be re processed")
			}
		}
	}
}

func (b BinanceBlockScanner) MatchedAddress(txInItem stypes.TxInItem) bool {
	// Check if we are migrating our funds...
	if ok := b.isMigration(txInItem.Sender, txInItem.Memo); ok {
		b.logger.Debug().Str("memo", txInItem.Memo).Msg("migrate")
		return true
	}

	// Check if our pool is registering a new yggdrasil pool. Ie
	// sending the staked assets to the user
	if ok := b.isRegisterYggdrasil(txInItem.Sender, txInItem.Memo); ok {
		b.logger.Debug().Str("memo", txInItem.Memo).Msg("yggdrasil+")
		return true
	}

	// Check if out pool is de registering a yggdrasil pool. Ie sending
	// the bond back to the user
	if ok := b.isDeregisterYggdrasil(txInItem.Sender, txInItem.Memo); ok {
		b.logger.Debug().Str("memo", txInItem.Memo).Msg("yggdrasil-")
		return true
	}

	// Check if THORNode are sending from a yggdrasil address
	if ok := b.isYggdrasil(txInItem.Sender); ok {
		b.logger.Debug().Str("assets sent from yggdrasil pool", txInItem.Memo).Msg("fill order")
		return true
	}

	// Check if THORNode are sending to a yggdrasil address
	if ok := b.isYggdrasil(txInItem.To); ok {
		b.logger.Debug().Str("assets to yggdrasil pool", txInItem.Memo).Msg("refill")
		return true
	}

	// outbound message from pool, when it is outbound, it does not matter how much coins THORNode send to customer for now
	if ok := b.isOutboundMsg(txInItem.Sender, txInItem.Memo); ok {
		b.logger.Debug().Str("memo", txInItem.Memo).Msg("outbound")
		return true
	}

	return false
}

// Check if memo is for registering an Asgard vault
func (b *BinanceBlockScanner) isMigration(addr, memo string) bool {
	return b.isAddrWithMemo(addr, memo, "migrate")
}

// Check if memo is for registering a Yggdrasil vault
func (b *BinanceBlockScanner) isRegisterYggdrasil(addr, memo string) bool {
	return b.isAddrWithMemo(addr, memo, "yggdrasil+")
}

// Check if memo is for de registering a Yggdrasil vault
func (b *BinanceBlockScanner) isDeregisterYggdrasil(addr, memo string) bool {
	return b.isAddrWithMemo(addr, memo, "yggdrasil-")
}

// Check if THORNode have an outbound yggdrasil transaction
func (b *BinanceBlockScanner) isYggdrasil(addr string) bool {
	ok, _ := b.pubkeyMgr.IsValidPoolAddress(addr, common.BNBChain)
	return ok
}

func (b *BinanceBlockScanner) isOutboundMsg(addr, memo string) bool {
	return b.isAddrWithMemo(addr, memo, "outbound")
}

func (b *BinanceBlockScanner) isAddrWithMemo(addr, memo, targetMemo string) bool {
	match, _ := b.pubkeyMgr.IsValidPoolAddress(addr, common.BNBChain)
	if !match {
		return false
	}
	lowerMemo := strings.ToLower(memo)
	if strings.HasPrefix(lowerMemo, targetMemo) {
		return true
	}
	return false
}

func (b *BinanceBlockScanner) getCoinsForTxIn(outputs []bmsg.Output) (common.Coins, error) {
	cc := common.Coins{}
	for _, output := range outputs {
		for _, c := range output.Coins {
			asset, err := common.NewAsset(fmt.Sprintf("BNB.%s", c.Denom))
			if err != nil {
				b.errCounter.WithLabelValues("fail_create_ticker", c.Denom).Inc()
				return nil, errors.Wrapf(err, "fail to create asset, %s is not valid", c.Denom)
			}
			amt := sdk.NewUint(uint64(c.Amount))
			cc = append(cc, common.NewCoin(asset, amt))
		}
	}
	return cc, nil
}

func (b *BinanceBlockScanner) fromTxToTxIn(hash, encodedTx string) ([]stypes.TxInItem, error) {
	if len(encodedTx) == 0 {
		return nil, errors.New("tx is empty")
	}
	buf, err := base64.StdEncoding.DecodeString(encodedTx)
	if err != nil {
		b.errCounter.WithLabelValues("fail_decode_tx", hash).Inc()
		return nil, errors.Wrap(err, "fail to decode tx")
	}
	var t tx.StdTx
	if err := tx.Cdc.UnmarshalBinaryLengthPrefixed(buf, &t); err != nil {
		b.errCounter.WithLabelValues("fail_unmarshal_tx", hash).Inc()
		return nil, errors.Wrap(err, "fail to unmarshal tx.StdTx")
	}

	return b.fromStdTx(hash, t)
}

// fromStdTx - process a stdTx
func (b *BinanceBlockScanner) fromStdTx(hash string, stdTx tx.StdTx) ([]stypes.TxInItem, error) {
	var err error
	var txs []stypes.TxInItem

	// TODO: It is also possible to have multiple inputs/outputs within a
	// single stdTx, which THORNode are not yet accounting for.
	for _, msg := range stdTx.Msgs {
		switch sendMsg := msg.(type) {
		case bmsg.SendMsg:
			txInItem := stypes.TxInItem{
				Tx: hash,
			}
			txInItem.Memo = stdTx.Memo
			// THORNode take the first Input as sender, first Output as receiver
			// so if THORNode send to multiple different receiver within one tx, this won't be able to process it.
			sender := sendMsg.Inputs[0]
			receiver := sendMsg.Outputs[0]
			txInItem.Sender = sender.Address.String()
			txInItem.To = receiver.Address.String()
			txInItem.Coins, err = b.getCoinsForTxIn(sendMsg.Outputs)
			if err != nil {
				return nil, errors.Wrap(err, "fail to convert coins")
			}

			// TODO: We should not assume what the gas fees are going to be in
			// the future (although they are largely static for binance). We
			// should modulus the binance block height and get the latest fee
			// prices every 1,000 or so blocks. This would ensure that all
			// observers will always report the same gas prices as they update
			// their price fees at the same time.

			// Calculate gas for this tx
			if len(txInItem.Coins) > 1 {
				// Multisend gas fees
				txInItem.Gas = common.GetBNBGasFeeMulti(uint64(len(txInItem.Coins)))
			} else {
				// Single transaction gas fees
				txInItem.Gas = common.BNBGasFeeSingleton
			}

			if ok := b.MatchedAddress(txInItem); !ok {
				continue
			}

			// NOTE: the following could result in the same tx being added
			// twice, which is expected. We want to make sure we generate both
			// a inbound and outbound txn, if we both apply.

			// check if the from address is a valid pool
			if ok, cpi := b.pubkeyMgr.IsValidPoolAddress(txInItem.Sender, common.BNBChain); ok {
				txInItem.ObservedPoolAddress = cpi.PubKey.String()
				txs = append(txs, txInItem)
			}
			// check if the to address is a valid pool address
			if ok, cpi := b.pubkeyMgr.IsValidPoolAddress(txInItem.To, common.BNBChain); ok {
				txInItem.ObservedPoolAddress = cpi.PubKey.String()
				txs = append(txs, txInItem)
			} else {
				// Apparently we don't recognize where we are sending funds to.
				// Lets check if we should because its an internal transaction
				// moving funds between vaults (for example). If it is, lets
				// manually trigger an update of pubkeys, then check again...
				switch strings.ToLower(txInItem.Memo) {
				case "migrate", "yggdrasil-", "yggdrasil+":
					b.pubkeyMgr.FetchPubKeys()
					if ok, cpi := b.pubkeyMgr.IsValidPoolAddress(txInItem.To, common.BNBChain); ok {
						txInItem.ObservedPoolAddress = cpi.PubKey.String()
						txs = append(txs, txInItem)
					}
				}
			}

		default:
			continue
		}
	}
	return txs, nil
}

func (b *BinanceBlockScanner) Stop() error {
	b.logger.Debug().Msg("receive stop request")
	defer b.logger.Debug().Msg("block scanner stopped")
	if err := b.commonBlockScanner.Stop(); err != nil {
		b.logger.Error().Err(err).Msg("fail to stop common block scanner")
	}
	close(b.stopChan)
	b.wg.Wait()

	return nil
}
