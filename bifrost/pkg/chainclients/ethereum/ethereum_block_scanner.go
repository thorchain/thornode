package ethereum

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	btypes "gitlab.com/thorchain/thornode/bifrost/blockscanner/types"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients/ethereum/types"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
)

const (
	DefaultObserverLevelDBFolder = `observer_data`
	GasPriceUpdateInterval       = 100
	DefaultGasPrice              = 1
	ETHTransferGas               = uint64(21000)
)

// BlockScanner is to scan the blocks
type BlockScanner struct {
	cfg        config.BlockScannerConfiguration
	logger     zerolog.Logger
	httpClient *http.Client
	db         blockscanner.ScannerStorage
	m          *metrics.Metrics
	errCounter *prometheus.CounterVec
	gasPrice   *big.Int
	client     *ethclient.Client
}

// NewBlockScanner create a new instance of BlockScan
func NewBlockScanner(cfg config.BlockScannerConfiguration, scanStorage blockscanner.ScannerStorage, chainID types.ChainID, client *ethclient.Client, m *metrics.Metrics) (*BlockScanner, error) {
	if scanStorage == nil {
		return nil, errors.New("scanStorage is nil")
	}
	if m == nil {
		return nil, errors.New("metrics is nil")
	}
	eipSigner = etypes.NewEIP155Signer(big.NewInt(int64(chainID)))
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	return &BlockScanner{
		cfg:        cfg,
		logger:     log.Logger.With().Str("module", "blockscanner").Str("chain", common.ETHChain.String()).Logger(),
		db:         scanStorage,
		errCounter: m.GetCounterVec(metrics.BlockScanError(common.ETHChain)),
		client:     client,
		gasPrice:   gasPrice,
		httpClient: &http.Client{
			Timeout: cfg.HttpRequestTimeout,
		},
	}, nil
}

// GetTxHash return hex formatted value of tx hash
func GetTxHash(encodedTx string) (string, error) {
	var tx *etypes.Transaction = &etypes.Transaction{}
	if err := tx.UnmarshalJSON([]byte(encodedTx)); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", tx.Hash().Hex()), nil
}

// GetGasPrice returns current gas price
func (e *BlockScanner) GetGasPrice() *big.Int {
	return e.gasPrice
}

// processBlock extracts transactions from block
func (e *BlockScanner) processBlock(block blockscanner.Block) (stypes.TxIn, error) {
	noTx := stypes.TxIn{}
	var err error

	strBlock := strconv.FormatInt(block.Height, 10)
	if err = e.db.SetBlockScanStatus(block, blockscanner.Processing); err != nil {
		e.errCounter.WithLabelValues("fail_set_block_status", strBlock).Inc()
		return noTx, fmt.Errorf("fail to set block scan status for block %d: %w", block.Height, err)
	}

	if len(block.Txs) == 0 {
		e.m.GetCounter(metrics.BlockWithoutTx("ETH")).Inc()
		return noTx, nil
	}
	// Update gas price once per 100 blocks
	if block.Height%GasPriceUpdateInterval == 0 {
		gasPrice, err := e.client.SuggestGasPrice(context.Background())
		if err != nil {
			return noTx, nil
		}
		e.gasPrice = gasPrice
	}

	var txIn stypes.TxIn
	for _, txn := range block.Txs {
		hash, err := GetTxHash(txn)
		if err != nil {
			e.errCounter.WithLabelValues("fail_get_tx_hash", strBlock).Inc()
			e.logger.Error().Err(err).Str("tx", txn).Msg("fail to get tx hash from raw data")
			return noTx, fmt.Errorf("fail to get tx hash from tx raw data: %w", err)
		}

		txInItem, err := e.fromTxToTxIn(txn)
		if err != nil {
			e.errCounter.WithLabelValues("fail_get_tx", strBlock).Inc()
			e.logger.Error().Err(err).Str("hash", hash).Msg("fail to get one tx from server")
			// if THORNode fail to get one tx hash from server, then THORNode should bail, because THORNode might miss tx
			// if THORNode bail here, then THORNode should retry later
			return noTx, fmt.Errorf("fail to get one tx from server: %w", err)
		}
		if txInItem != nil {
			txIn.TxArray = append(txIn.TxArray, *txInItem)
			e.m.GetCounter(metrics.BlockWithTxIn("ETH")).Inc()
			e.logger.Info().Str("hash", hash).Msg("THORNode got one tx")
		}
	}
	if len(txIn.TxArray) == 0 {
		e.m.GetCounter(metrics.BlockNoTxIn("ETH")).Inc()
		e.logger.Debug().Int64("block", block.Height).Msg("no tx need to be processed in this block")
		return noTx, nil
	}

	txIn.BlockHeight = strconv.FormatInt(block.Height, 10)
	txIn.Count = strconv.Itoa(len(txIn.TxArray))
	txIn.Chain = common.ETHChain
	return txIn, nil
}

func (e *BlockScanner) FetchTxs(height int64) (stypes.TxIn, error) {
	rawTxs, err := e.getRPCBlock(height)
	if err != nil {
		return stypes.TxIn{}, err
	}

	block := blockscanner.Block{Height: height, Txs: rawTxs}
	txIn, err := e.processBlock(block)
	if err != nil {
		if errStatus := e.db.SetBlockScanStatus(block, blockscanner.Failed); errStatus != nil {
			e.errCounter.WithLabelValues("fail_set_block_status", "").Inc()
			e.logger.Error().Err(err).Int64("height", block.Height).Msg("fail to set block to fail status")
		}
		e.errCounter.WithLabelValues("fail_search_block", "").Inc()
		e.logger.Error().Err(err).Int64("height", block.Height).Msg("fail to search tx in block")
		// THORNode will have a retry go routine to check it.
		return txIn, err
	}
	// set a block as success
	if err := e.db.RemoveBlockStatus(block.Height); err != nil {
		e.errCounter.WithLabelValues("fail_remove_block_status", "").Inc()
		e.logger.Error().Err(err).Int64("block", block.Height).Msg("fail to remove block status from data store, thus block will be re processed")
	}
	return txIn, nil
}

func (e *BlockScanner) getRPCBlock(height int64) ([]string, error) {
	block, err := e.client.BlockByNumber(context.Background(), big.NewInt(height))
	if err == ethereum.NotFound {
		return nil, btypes.UnavailableBlock
	}
	if err != nil {
		e.logger.Error().Err(err).Int64("block", height).Msg("fail to fetch block")
		return nil, err
	}
	rawTxs, err := e.getTransactionsFromBlock(block)
	if err != nil {
		e.errCounter.WithLabelValues("fail_to_get_txs", e.cfg.RPCHost).Inc()
	}
	return rawTxs, err
}

func (e *BlockScanner) getTransactionsFromBlock(block *etypes.Block) ([]string, error) {
	txs := make([]string, 0)
	for _, tx := range block.Transactions() {
		bytes, err := tx.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("fail to marshal tx from block: %w", err)
		}
		txs = append(txs, string(bytes))
	}
	return txs, nil
}

func (e *BlockScanner) fromTxToTxIn(encodedTx string) (*stypes.TxInItem, error) {
	if len(encodedTx) == 0 {
		return nil, errors.New("tx is empty")
	}
	var tx *etypes.Transaction = &etypes.Transaction{}
	if err := tx.UnmarshalJSON([]byte(encodedTx)); err != nil {
		return nil, err
	}

	txInItem := &stypes.TxInItem{
		Tx: tx.Hash().Hex()[2:],
	}
	// tx data field bytes should be hex encoded byres string as outboud or yggradsil- or migrate or yggdrasil+, etc
	txInItem.Memo = string(tx.Data())

	sender, err := eipSigner.Sender(tx)
	if err != nil {
		return nil, err
	}
	txInItem.Sender = strings.ToLower(sender.String())
	if tx.To() == nil {
		return nil, err
	}
	txInItem.To = strings.ToLower(tx.To().String())

	asset, err := common.NewAsset("ETH.ETH")
	if err != nil {
		e.errCounter.WithLabelValues("fail_create_ticker", "ETH").Inc()
		return nil, fmt.Errorf("fail to create asset, ETH is not valid: %w", err)
	}

	txInItem.Coins = append(txInItem.Coins, common.NewCoin(asset, sdk.NewUintFromBigInt(tx.Value())))
	txInItem.Gas = common.GetETHGasFee(e.gasPrice, uint64(len(txInItem.Memo)))

	return txInItem, nil
}
