package ethereum

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/types/tx"
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

// EthereumBlockScanner is to scan the blocks
type EthereumBlockScanner struct {
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
	signer             etypes.EIP155Signer
	rpcHost            string
}

// NewEthereumBlockScanner create a new instance of BlockScan
func NewEthereumBlockScanner(cfg config.BlockScannerConfiguration, startBlockHeight int64, signer etypes.EIP155Signer, isTestNet bool, client *ethclient.Client, pkmgr pubkeymanager.PubKeyValidator, m *metrics.Metrics) (*EthereumBlockScanner, error) {
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
	if level {
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

	return &EthereumBlockScanner{
		cfg:                cfg,
		pubkeyMgr:          pkmgr,
		logger:             log.Logger.With().Str("module", "blockscanner").Logger(),
		wg:                 &sync.WaitGroup{},
		stopChan:           make(chan struct{}),
		db:                 scanStorage,
		commonBlockScanner: commonBlockScanner,
		errCounter:         m.GetCounterVec(metrics.BlockScanError(common.ETHChain)),
		rpcHost:            rpcHost,
		client:             client,
		signer:             signer,
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

// Start block scanner
func (e *EthereumBlockScanner) Start(globalTxsQueue chan stypes.TxIn) {
	e.globalTxsQueue = globalTxsQueue
	for idx := 1; idx <= e.cfg.BlockScanProcessors; idx++ {
		e.wg.Add(1)
		go e.searchTxInABlock(idx)
	}
	e.commonBlockScanner.Start()
}

// getTxHash return hex formatted value of tx hash
func (e *EthereumBlockScanner) getTxHash(encodedTx string) (string, error) {
	return fmt.Sprintf("%X", string(crypto.Keccak256([]byte(encodedTx)))), nil
}

func (e *EthereumBlockScanner) processBlock(block blockscanner.Block) error {
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

	// TODO implement pagination appropriately
	var txIn stypes.TxIn
	for _, txn := range block.Txs {
		hash, err := e.getTxHash(txn)
		if err != nil {
			e.errCounter.WithLabelValues("fail_get_tx_hash", strBlock).Inc()
			e.logger.Error().Err(err).Str("tx", txn).Msg("fail to get tx hash from raw data")
			return errors.Wrap(err, "fail to get tx hash from tx raw data")
		}

		txItemIns, err := e.fromTxToTxIn(hash, txn)
		if err != nil {
			e.errCounter.WithLabelValues("fail_get_tx", strBlock).Inc()
			e.logger.Error().Err(err).Str("hash", hash).Msg("fail to get one tx from server")
			// if THORNode fail to get one tx hash from server, then THORNode should bail, because THORNode might miss tx
			// if THORNode bail here, then THORNode should retry later
			return errors.Wrap(err, "fail to get one tx from server")
		}
		if len(txItemIns) > 0 {
			txIn.TxArray = append(txIn.TxArray, txItemIns...)
			e.m.GetCounter(metrics.BlockWithTxIn("ETH")).Inc()
			e.logger.Info().Str("hash", hash).Msg("THORNode got one tx")
		}
	}
	if len(txIn.TxArray) == 0 {
		e.m.GetCounter(metrics.BlockNoTxIn("ETH")).Inc()
		e.logger.Debug().Int64("block", block.Height).Msg("no tx need to be processed in this block")
		return nil
	}

	txIn.BlockHeight = strconv.FormatInt(block.Height, 10)
	txIn.Count = strconv.Itoa(len(txIn.TxArray))
	txIn.Chain = common.ETHChain
	e.globalTxsQueue <- txIn
	return nil
}

func (e *EthereumBlockScanner) searchTxInABlock(idx int) {
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

func (e EthereumBlockScanner) MatchedAddress(txInItem stypes.TxInItem) bool {
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
func (e *EthereumBlockScanner) isMigration(addr, memo string) bool {
	return e.isAddrWithMemo(addr, memo, "migrate")
}

// Check if memo is for registering a Yggdrasil vault
func (e *EthereumBlockScanner) isRegisterYggdrasil(addr, memo string) bool {
	return e.isAddrWithMemo(addr, memo, "yggdrasil+")
}

// Check if memo is for de registering a Yggdrasil vault
func (e *EthereumBlockScanner) isDeregisterYggdrasil(addr, memo string) bool {
	return e.isAddrWithMemo(addr, memo, "yggdrasil-")
}

// Check if THORNode have an outbound yggdrasil transaction
func (e *EthereumBlockScanner) isYggdrasil(addr string) bool {
	ok, _ := e.pubkeyMgr.IsValidPoolAddress(addr, common.ETHChain)
	return ok
}

func (e *EthereumBlockScanner) isOutboundMsg(addr, memo string) bool {
	return e.isAddrWithMemo(addr, memo, "outbound")
}

func (e *EthereumBlockScanner) isAddrWithMemo(addr, memo, targetMemo string) bool {
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

func (e *EthereumBlockScanner) getCoinsForTxIn(tx etypes.Transaction) (common.Coins, error) {
	asset, err := common.NewAsset("ETH.ETH")
	if err != nil {
		e.errCounter.WithLabelValues("fail_create_ticker", "ETH").Inc()
		return nil, errors.Wrap(err, "fail to create asset, ETH is not valid")
	}
	amt := sdk.NewUint(c.Value().NumberU64())
	return common.Coins{common.NewCoin(asset, amt)}, nil
}

func (e *EthereumBlockScanner) fromTxToTxIn(encodedTx string) ([]stypes.TxInItem, error) {
	if len(encodedTx) == 0 {
		return nil, errors.New("tx is empty")
	}
	var tx etypes.Transaction
	if err := json.Unmarshal([]byte(encodedTx), &tx); err != nil {
		return err
	}
	return e.fromStdTx(tx)
}

// fromStdTx - process a stdTx
func (e *EthereumBlockScanner) fromStdTx(tx etypes.Transaction) ([]stypes.TxInItem, error) {
	var err error
	var txs []stypes.TxInItem

	txInItem := stypes.TxInItem{
		Tx: tx.Hash().String(),
	}
	txInItem.Memo = string(tx.Data())
	// THORNode take the first Input as sender, first Output as receiver
	// so if THORNode send to multiple different receiver within one tx, this won't be able to process it.
	sender, err := e.signer.Sender(tx)
	if err != nil {	
		return make([]stypes.TxInItem, 0), nil
	}
	txInItem.Sender = sender.String()
	if tx.To() == nil {
		return make([]stypes.TxInItem, 0), errors.New("missing receiver")
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
		continue
	}

	// NOTE: the following could result in the same tx being added
	// twice, which is expected. We want to make sure we generate both
	// a inbound and outbound txn, if we both apply.

	// check if the from address is a valid pool
	if ok, cpi := e.pubkeyMgr.IsValidPoolAddress(txInItem.Sender, common.ETHChain); ok {
		txInItem.ObservedPoolAddress = cpi.PubKey.String()
		txs = append(txs, txInItem)
	}
	// check if the to address is a valid pool address
	if ok, cpi := e.pubkeyMgr.IsValidPoolAddress(txInItem.To, common.ETHChain); ok {
		txInItem.ObservedPoolAddress = cpi.PubKey.String()
		txs = append(txs, txInItem)
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
				txs = append(txs, txInItem)
			}
		}
	}
	return txs, nil
}

func (e *EthereumBlockScanner) Stop() error {
	e.logger.Debug().Msg("receive stop request")
	defer e.logger.Debug().Msg("block scanner stopped")
	if err := e.commonBlockScanner.Stop(); err != nil {
		e.logger.Error().Err(err).Msg("fail to stop common block scanner")
	}
	close(e.stopChan)
	e.wg.Wait()

	return nil
}
