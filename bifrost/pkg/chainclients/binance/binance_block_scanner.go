package binance

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/binance-chain/go-sdk/common/types"
	bmsg "github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/types/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/go-amino"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	bltypes "gitlab.com/thorchain/thornode/bifrost/blockscanner/types"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	btypes "gitlab.com/thorchain/thornode/bifrost/pkg/chainclients/binance/types"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
)

// BinanceBlockScanner is to scan the blocks
type BinanceBlockScanner struct {
	cfg        config.BlockScannerConfiguration
	logger     zerolog.Logger
	db         blockscanner.ScannerStorage
	m          *metrics.Metrics
	errCounter *prometheus.CounterVec
	http       *http.Client
	singleFee  uint64
	multiFee   uint64
}

// NewBinanceBlockScanner create a new instance of BlockScan
func NewBinanceBlockScanner(cfg config.BlockScannerConfiguration, scanStorage blockscanner.ScannerStorage, isTestNet bool, m *metrics.Metrics) (*BinanceBlockScanner, error) {
	if scanStorage == nil {
		return nil, errors.New("scanStorage is nil")
	}
	if m == nil {
		return nil, errors.New("metrics is nil")
	}
	if isTestNet {
		types.Network = types.TestNetwork
	} else {
		types.Network = types.ProdNetwork
	}

	netClient := &http.Client{
		Timeout: cfg.HttpRequestTimeout,
	}

	return &BinanceBlockScanner{
		cfg:        cfg,
		logger:     log.Logger.With().Str("module", "blockscanner").Str("chain", "binance").Logger(),
		db:         scanStorage,
		errCounter: m.GetCounterVec(metrics.BlockScanError(common.BNBChain)),
		http:       netClient,
	}, nil
}

// getTxHash return hex formatted value of tx hash
// raw tx base 64 encoded -> base64 decode -> sha256sum = tx hash
func (b *BinanceBlockScanner) getTxHash(encodedTx string) (string, error) {
	decodedTx, err := base64.StdEncoding.DecodeString(encodedTx)
	if err != nil {
		b.errCounter.WithLabelValues("fail_decode_tx", encodedTx).Inc()
		return "", fmt.Errorf("fail to decode tx: %w", err)
	}
	return fmt.Sprintf("%X", sha256.Sum256(decodedTx)), nil
}

func (b *BinanceBlockScanner) updateFees(height int64) error {
	url := fmt.Sprintf("%s/abci_query?path=\"/param/fees\"&height=%d", b.cfg.RPCHost, height)
	resp, err := b.http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get current gas fees: non 200 error (%d)", resp.StatusCode)
	}

	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var result btypes.QueryResult
	if err := json.Unmarshal(bz, &result); err != nil {
		return err
	}

	val, err := base64.StdEncoding.DecodeString(result.Result.Response.Value)
	if err != nil {
		return err
	}

	var fees []types.FeeParam
	cdc := amino.NewCodec()
	types.RegisterWire(cdc)
	err = cdc.UnmarshalBinaryLengthPrefixed(val, &fees)
	if err != nil {
		return err
	}

	for _, fee := range fees {
		if fee.GetParamType() == types.TransferFeeType {
			if err := fee.Check(); err != nil {
				return err
			}

			transferFee := fee.(*types.TransferFeeParam)
			if transferFee.FixedFeeParams.Fee > 0 {
				b.singleFee = uint64(transferFee.FixedFeeParams.Fee)
			}
			if transferFee.MultiTransferFee > 0 {
				b.multiFee = uint64(transferFee.MultiTransferFee)
			}
		}
	}

	return nil
}

func (b *BinanceBlockScanner) processBlock(block blockscanner.Block) (stypes.TxIn, error) {
	var txIn stypes.TxIn
	strBlock := strconv.FormatInt(block.Height, 10)
	if err := b.db.SetBlockScanStatus(block, blockscanner.Processing); err != nil {
		b.errCounter.WithLabelValues("fail_set_block_status", strBlock).Inc()
		return txIn, fmt.Errorf("fail to set block scan status for block %d: %w", block.Height, err)
	}

	b.logger.Debug().Int64("block", block.Height).Int("txs", len(block.Txs)).Msg("txs")
	if len(block.Txs) == 0 {
		b.m.GetCounter(metrics.BlockWithoutTx("BNB")).Inc()
		b.logger.Debug().Int64("block", block.Height).Msg("there are no txs in this block")
		return txIn, nil
	}

	// update our gas fees from binance RPC node
	if err := b.updateFees(block.Height); err != nil {
		b.logger.Error().Err(err).Msg("fail to update Binance gas fees")
	}

	// TODO implement pagination appropriately
	for _, txn := range block.Txs {
		hash, err := b.getTxHash(txn)
		if err != nil {
			b.errCounter.WithLabelValues("fail_get_tx_hash", strBlock).Inc()
			b.logger.Error().Err(err).Str("tx", txn).Msg("fail to get tx hash from raw data")
			return txIn, fmt.Errorf("fail to get tx hash from tx raw data: %w", err)
		}

		txItemIns, err := b.fromTxToTxIn(hash, txn)
		if err != nil {
			b.errCounter.WithLabelValues("fail_get_tx", strBlock).Inc()
			b.logger.Error().Err(err).Str("hash", hash).Msg("fail to get one tx from server")
			// if THORNode fail to get one tx hash from server, then THORNode should bail, because THORNode might miss tx
			// if THORNode bail here, then THORNode should retry later
			return txIn, fmt.Errorf("fail to get one tx from server: %w", err)
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
		return txIn, nil
	}

	txIn.BlockHeight = strconv.FormatInt(block.Height, 10)
	txIn.Count = strconv.Itoa(len(txIn.TxArray))
	txIn.Chain = common.BNBChain
	return txIn, nil
}

func (b *BinanceBlockScanner) getFromHttp(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		b.errCounter.WithLabelValues("fail_create_http_request", url).Inc()
		return nil, fmt.Errorf("fail to create http request: %w", err)
	}
	resp, err := b.http.Do(req)
	if err != nil {
		b.errCounter.WithLabelValues("fail_send_http_request", url).Inc()
		return nil, fmt.Errorf("fail to get from %s: %w", url, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			b.logger.Error().Err(err).Msg("fail to close http response body.")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		b.errCounter.WithLabelValues("unexpected_status_code", resp.Status).Inc()
		return nil, fmt.Errorf("unexpected status code:%d from %s", resp.StatusCode, url)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// test if our response body is an error block json format
	errorBlock := struct {
		Error struct {
			Code    int64  `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		} `json:"error"`
	}{}

	_ = json.Unmarshal(buf, &errorBlock) // ignore error
	if errorBlock.Error.Code != 0 {
		return nil, fmt.Errorf(
			"%s (%d): %s",
			errorBlock.Error.Message,
			errorBlock.Error.Code,
			errorBlock.Error.Data,
		)
	}

	return buf, nil
}

func (b *BinanceBlockScanner) getRPCBlock(height int64) ([]string, error) {
	start := time.Now()
	defer func() {
		if err := recover(); err != nil {
			b.logger.Error().Msgf("fail to get RPCBlock:%s", err)
		}
		duration := time.Since(start)
		b.m.GetHistograms(metrics.BlockDiscoveryDuration).Observe(duration.Seconds())
	}()
	url := b.BlockRequest(height)
	buf, err := b.getFromHttp(url)
	if err != nil {
		b.errCounter.WithLabelValues("fail_get_block", url).Inc()
		time.Sleep(b.cfg.BlockHeightDiscoverBackoff)
		if strings.Contains(err.Error(), "Height must be less than or equal to the current blockchain height") {
			return nil, bltypes.UnavailableBlock
		}
		return nil, err
	}

	rawTxns, err := b.UnmarshalBlock(buf)
	if err != nil {
		b.errCounter.WithLabelValues("fail_unmarshal_block", url).Inc()
	}
	return rawTxns, err
}

func (b *BinanceBlockScanner) BlockRequest(height int64) string {
	u, _ := url.Parse(b.cfg.RPCHost)
	u.Path = "block"
	if height > 0 {
		u.RawQuery = fmt.Sprintf("height=%d", height)
	}
	return u.String()
}

func (b *BinanceBlockScanner) UnmarshalBlock(buf []byte) ([]string, error) {
	// check if the block is null. This can happen when binance gets the block,
	// but not the data within it. In which case, we'll never have the data and
	// we should just move onto the next block.
	// { "jsonrpc": "2.0", "id": "", "result": { "block_meta": null, "block": null } }
	if bytes.Contains(buf, []byte(`"block": null`)) {
		return nil, nil
	}

	var block btypes.BlockResult
	err := json.Unmarshal(buf, &block)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal body to rpcBlock: %w", err)
	}

	return block.Result.Block.Data.Txs, nil
}

func (b *BinanceBlockScanner) FetchTxs(height int64) (stypes.TxIn, error) {
	rawTxs, err := b.getRPCBlock(height)
	if err != nil {
		return stypes.TxIn{}, err
	}

	block := blockscanner.Block{Height: height, Txs: rawTxs}
	b.logger.Debug().Int64("block", block.Height).Msg("processing block")
	txIn, err := b.processBlock(block)
	if err != nil {
		if errStatus := b.db.SetBlockScanStatus(block, blockscanner.Failed); errStatus != nil {
			b.errCounter.WithLabelValues("fail_set_block_status", "").Inc()
			b.logger.Error().Err(err).Int64("height", block.Height).Msg("fail to set block to fail status")
		}
		b.errCounter.WithLabelValues("fail_search_block", "").Inc()
		b.logger.Error().Err(err).Int64("height", block.Height).Msg("fail to search tx in block")
		// THORNode will have a retry go routine to check it.
		return txIn, err
	}
	// set a block as success
	if err := b.db.RemoveBlockStatus(block.Height); err != nil {
		b.errCounter.WithLabelValues("fail_remove_block_status", "").Inc()
		b.logger.Error().Err(err).Int64("block", block.Height).Msg("fail to remove block status from data store, thus block will be re processed")
	}
	return txIn, nil
}

func (b *BinanceBlockScanner) getCoinsForTxIn(outputs []bmsg.Output) (common.Coins, error) {
	cc := common.Coins{}
	for _, output := range outputs {
		for _, c := range output.Coins {
			asset, err := common.NewAsset(fmt.Sprintf("BNB.%s", c.Denom))
			if err != nil {
				b.errCounter.WithLabelValues("fail_create_ticker", c.Denom).Inc()
				return nil, fmt.Errorf("fail to create asset, %s is not valid: %w", c.Denom, err)
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
		return nil, fmt.Errorf("fail to decode tx: %w", err)
	}
	var t tx.StdTx
	if err := tx.Cdc.UnmarshalBinaryLengthPrefixed(buf, &t); err != nil {
		b.errCounter.WithLabelValues("fail_unmarshal_tx", hash).Inc()
		return nil, fmt.Errorf("fail to unmarshal tx.StdTx: %w", err)
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
				return nil, fmt.Errorf("fail to convert coins: %w", err)
			}

			// Calculate gas for this tx
			txInItem.Gas = common.CalcGasPrice(common.Tx{Coins: txInItem.Coins}, common.BNBAsset, []sdk.Uint{sdk.NewUint(b.singleFee), sdk.NewUint(b.multiFee)})

			txs = append(txs, txInItem)
		default:
			continue
		}
	}
	return txs, nil
}
