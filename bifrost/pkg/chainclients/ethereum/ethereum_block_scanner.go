package ethereum

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/syndtr/goleveldb/leveldb"

	"gitlab.com/thorchain/thornode/common"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients/ethereum/types"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

const DefaultObserverLevelDBFolder = `observer_data`

var eipSigner = etypes.NewEIP155Signer(big.NewInt(1))

// BlockScanner is to scan the blocks
type BlockScanner struct {
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
	client             *ethclient.Client
	rpcHost            string
}

// NewBlockScanner create a new instance of BlockScan
func NewBlockScanner(cfg config.BlockScannerConfiguration, startBlockHeight int64, isTestNet bool, client *ethclient.Client, pkmgr pubkeymanager.PubKeyValidator, m *metrics.Metrics) (*BlockScanner, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("rpc host is empty")
	}

	rpcHost := cfg.RPCHost
	if !strings.HasPrefix(rpcHost, "http") {
		rpcHost = fmt.Sprintf("http://%s", rpcHost)
	}
	scanStorage, err := NewStorage(cfg.DBPath)
	if err != nil {
		return nil, err
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
	commonBlockScanner, err := blockscanner.NewCommonBlockScanner(cfg, startBlockHeight, scanStorage, m, types.EthereumSupplemental{})
	if err != nil {
		return nil, errors.Wrap(err, "fail to create common block scanner")
	}

	return &BlockScanner{
		cfg:                cfg,
		pubkeyMgr:          pkmgr,
		logger:             log.Logger.With().Str("module", "blockscanner").Str("chain", "ethereum").Logger(),
		wg:                 &sync.WaitGroup{},
		stopChan:           make(chan struct{}),
		db:                 scanStorage,
		commonBlockScanner: commonBlockScanner,
		errCounter:         m.GetCounterVec(metrics.BlockScanError(common.ETHChain)),
		rpcHost:            rpcHost,
		client:             client,
	}, nil
}

func NewStorage(levelDbFolder string) (*blockscanner.LevelDBScannerStorage, error) {
	if len(levelDbFolder) == 0 {
		levelDbFolder = DefaultObserverLevelDBFolder
	}
	db, err := leveldb.OpenFile(levelDbFolder, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to open level db %s", levelDbFolder)
	}
	levelDbStorage, err := blockscanner.NewLevelDBScannerStorage(db)
	if err != nil {
		return nil, errors.New("fail to create leven db")
	}
	return levelDbStorage, nil
}

// Start starts block scanner
func (e *BlockScanner) Start(globalTxsQueue chan stypes.TxIn) {
	e.globalTxsQueue = globalTxsQueue
	for idx := 1; idx <= e.cfg.BlockScanProcessors; idx++ {
		e.wg.Add(1)
		go e.processBlocks(idx)
	}
	e.commonBlockScanner.Start()
}

// GetTxHash return hex formatted value of tx hash
func GetTxHash(encodedTx string) (string, error) {
	var tx etypes.Transaction
	if err := json.Unmarshal([]byte(encodedTx), &tx); err != nil {
		return "", err
	}
	return fmt.Sprintf("0x%x", tx.Hash().String()), nil
}

// processBlock extracts transactions from block
func (e *BlockScanner) processBlock(block blockscanner.Block) error {
	strBlock := strconv.FormatInt(block.Height, 10)
	if err := e.db.SetBlockScanStatus(block, blockscanner.Processing); err != nil {
		e.errCounter.WithLabelValues("fail_set_block_status", strBlock).Inc()
		return errors.Wrapf(err, "fail to set block scan status for block %d", block.Height)
	}

	e.logger.Debug().Int64("block", block.Height).Int("txs", len(block.Txs)).Msg("txs")
	if len(block.Txs) == 0 {
		e.m.GetCounter(metrics.BlockWithoutTx("ETH")).Inc()
		e.logger.Debug().Int64("block", block.Height).Msg("there are no txs in this block")
		return nil
	}

	var txIn stypes.TxIn
	for _, txn := range block.Txs {
		hash, err := GetTxHash(txn)
		if err != nil {
			e.errCounter.WithLabelValues("fail_get_tx_hash", strBlock).Inc()
			e.logger.Error().Err(err).Str("tx", txn).Msg("fail to get tx hash from raw data")
			return errors.Wrap(err, "fail to get tx hash from tx raw data")
		}

		txItemIn, err := e.fromTxToTxIn(txn)
		if err != nil {
			e.errCounter.WithLabelValues("fail_get_tx", strBlock).Inc()
			e.logger.Error().Err(err).Str("hash", hash).Msg("fail to get one tx from server")
			// if THORNode fail to get one tx hash from server, then THORNode should bail, because THORNode might miss tx
			// if THORNode bail here, then THORNode should retry later
			return errors.Wrap(err, "fail to get one tx from server")
		}
		if txItemIn != nil {
			txIn.TxArray = append(txIn.TxArray, *txItemIn)
			e.m.GetCounter(metrics.BlockWithTxIn("ETH")).Inc()
			e.logger.Info().Str("hash", hash).Msg("THORNode got one tx")
		}
	}
	if len(txIn.TxArray) == 0 {
		e.m.GetCounter(metrics.BlockNoTxIn("ETH")).Inc()
		e.logger.Debug().Int64("block", block.Height).Msg("no tx need to be processed in this block")
		return nil
	}

	// TODO implement postponing for transactions if total transfer values per block for address to secure against attacks
	txIn.BlockHeight = strconv.FormatInt(block.Height, 10)
	txIn.Count = strconv.Itoa(len(txIn.TxArray))
	txIn.Chain = common.ETHChain
	e.globalTxsQueue <- txIn

	return nil
}

// processBlocks processes blocks and gets transactions
func (e *BlockScanner) processBlocks(idx int) {
	e.logger.Debug().Int("idx", idx).Msg("start searching tx in a block")
	defer e.logger.Debug().Int("idx", idx).Msg("stop searching tx in a block")
	defer e.wg.Done()

	for {
		select {
		case <-e.stopChan: // time to get out
			return
		case block, more := <-e.commonBlockScanner.GetMessages():
			if !more {
				return
			}
			e.logger.Debug().Int64("block", block.Height).Msg("processing block")
			if err := e.processBlock(block); err != nil {
				if errStatus := e.db.SetBlockScanStatus(block, blockscanner.Failed); errStatus != nil {
					e.errCounter.WithLabelValues("fail_set_block_status", "").Inc()
					e.logger.Error().Err(err).Int64("height", block.Height).Msg("fail to set block to fail status")
				}
				e.errCounter.WithLabelValues("fail_search_block", "").Inc()
				e.logger.Error().Err(err).Int64("height", block.Height).Msg("fail to search tx in block")
				// THORNode will have a retry go routine to check it.
				continue
			}
			// set a block as success
			if err := e.db.RemoveBlockStatus(block.Height); err != nil {
				e.errCounter.WithLabelValues("fail_remove_block_status", "").Inc()
				e.logger.Error().Err(err).Int64("block", block.Height).Msg("fail to remove block status from data store, thus block will be re processed")
			}
		}
	}
}

func (e BlockScanner) MatchedAddress(txInItem *stypes.TxInItem) bool {
	// Check if we are migrating our funds...
	if ok := e.isMigration(txInItem.Sender, txInItem.Memo); ok {
		e.logger.Debug().Str("memo", txInItem.Memo).Msg("migrate")
		return true
	}

	// Check if our pool is registering a new yggdrasil pool. Ie
	// sending the staked assets to the user
	if ok := e.isRegisterYggdrasil(txInItem.Sender, txInItem.Memo); ok {
		e.logger.Debug().Str("memo", txInItem.Memo).Msg("yggdrasil+")
		return true
	}

	// Check if out pool is de registering a yggdrasil pool. Ie sending
	// the bond back to the user
	if ok := e.isDeregisterYggdrasil(txInItem.Sender, txInItem.Memo); ok {
		e.logger.Debug().Str("memo", txInItem.Memo).Msg("yggdrasil-")
		return true
	}

	// Check if THORNode are sending from a yggdrasil address
	if ok := e.isYggdrasil(txInItem.Sender); ok {
		e.logger.Debug().Str("assets sent from yggdrasil pool", txInItem.Memo).Msg("fill order")
		return true
	}

	// Check if THORNode are sending to a yggdrasil address
	if ok := e.isYggdrasil(txInItem.To); ok {
		e.logger.Debug().Str("assets to yggdrasil pool", txInItem.Memo).Msg("refill")
		return true
	}

	// outbound message from pool, when it is outbound, it does not matter how much coins THORNode send to customer for now
	if ok := e.isOutboundMsg(txInItem.Sender, txInItem.Memo); ok {
		e.logger.Debug().Str("memo", txInItem.Memo).Msg("outbound")
		return true
	}

	return false
}

// Check if memo is for registering an Asgard vault
func (e *BlockScanner) isMigration(addr, memo string) bool {
	return e.isAddrWithMemo(addr, memo, "migrate")
}

// Check if memo is for registering a Yggdrasil vault
func (e *BlockScanner) isRegisterYggdrasil(addr, memo string) bool {
	return e.isAddrWithMemo(addr, memo, "yggdrasil+")
}

// Check if memo is for de registering a Yggdrasil vault
func (e *BlockScanner) isDeregisterYggdrasil(addr, memo string) bool {
	return e.isAddrWithMemo(addr, memo, "yggdrasil-")
}

// Check if THORNode have an outbound yggdrasil transaction
func (e *BlockScanner) isYggdrasil(addr string) bool {
	ok, _ := e.pubkeyMgr.IsValidPoolAddress(addr, common.ETHChain)
	return ok
}

func (e *BlockScanner) isOutboundMsg(addr, memo string) bool {
	return e.isAddrWithMemo(addr, memo, "outbound")
}

func (e *BlockScanner) isAddrWithMemo(addr, memo, targetMemo string) bool {
	match, _ := e.pubkeyMgr.IsValidPoolAddress(addr, common.ETHChain)
	if !match {
		return false
	}
	lowerMemo := strings.ToLower(memo)
	if strings.HasPrefix(lowerMemo, targetMemo) {
		return true
	}
	return false
}

func (e *BlockScanner) getCoinsForTxIn(tx *etypes.Transaction) (common.Coins, error) {
	asset, err := common.NewAsset("ETH.ETH")
	if err != nil {
		e.errCounter.WithLabelValues("fail_create_ticker", "ETH").Inc()
		return nil, errors.Wrap(err, "fail to create asset, ETH is not valid")
	}
	amt := sdk.NewUint(tx.Value().Uint64())
	return common.Coins{common.NewCoin(asset, amt)}, nil
}

func (e *BlockScanner) fromTxToTxIn(encodedTx string) (*stypes.TxInItem, error) {
	if len(encodedTx) == 0 {
		return nil, errors.New("tx is empty")
	}
	var tx *etypes.Transaction = &etypes.Transaction{}
	if err := json.Unmarshal([]byte(encodedTx), tx); err != nil {
		return nil, err
	}
	return e.fromStdTx(tx)
}

// fromStdTx - process a stdTx
func (e *BlockScanner) fromStdTx(tx *etypes.Transaction) (*stypes.TxInItem, error) {
	txInItem := &stypes.TxInItem{
		Tx: tx.Hash().String(),
	}
	// tx data field bytes should be exactly as outboud or yggradsil- or migrate or yggdrasil+, etc
	txInItem.Memo = string(tx.Data())

	sender, err := eipSigner.Sender(tx)
	if err != nil {
		return nil, nil
	}
	txInItem.Sender = sender.String()
	if tx.To() == nil {
		return nil, errors.New("missing receiver")
	}
	txInItem.To = tx.To().String()
	txInItem.Coins, err = e.getCoinsForTxIn(tx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to convert coins")
	}

	// TODO: We should not assume what the gas fees are going to be in
	// the future (although they are largely static for Ethereum). We
	// should modulus the Ethereum block height and get the latest fee
	// prices every 1,000 or so blocks. This would ensure that all
	// observers will always report the same gas prices as they update
	// their price fees at the same time.

	txInItem.Gas = common.GetETHGasFee()

	if ok := e.MatchedAddress(txInItem); !ok {
		return nil, errors.New("address is not matched")
	}

	// NOTE: the following could result in the same tx being added
	// twice, which is expected. We want to make sure we generate both
	// a inbound and outbound txn, if we both apply.

	// check if the from address is a valid pool
	if ok, cpi := e.pubkeyMgr.IsValidPoolAddress(txInItem.Sender, common.ETHChain); ok {
		txInItem.ObservedPoolAddress = cpi.PubKey.String()
	}
	// check if the to address is a valid pool address
	if ok, cpi := e.pubkeyMgr.IsValidPoolAddress(txInItem.To, common.ETHChain); ok {
		txInItem.ObservedPoolAddress = cpi.PubKey.String()
	} else {
		// Apparently we don't recognize where we are sending funds to.
		// Lets check if we should because its an internal transaction
		// moving funds between vaults (for example). If it is, lets
		// manually trigger an update of pubkeys, then check again...
		switch strings.ToLower(txInItem.Memo) {
		case "migrate", "yggdrasil-", "yggdrasil+":
			e.pubkeyMgr.FetchPubKeys()
			if ok, cpi := e.pubkeyMgr.IsValidPoolAddress(txInItem.To, common.ETHChain); ok {
				txInItem.ObservedPoolAddress = cpi.PubKey.String()
			}
		}
	}
	return txInItem, nil
}

// Stop stops block scanner
func (e *BlockScanner) Stop() error {
	e.logger.Debug().Msg("receive stop request")
	defer e.logger.Debug().Msg("block scanner stopped")
	if err := e.commonBlockScanner.Stop(); err != nil {
		e.logger.Error().Err(err).Msg("fail to stop common block scanner")
	}
	close(e.stopChan)
	e.wg.Wait()
	return nil
}
