package ethereum

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/types"

	"gitlab.com/thorchain/thornode/common"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	etypes "gitlab.com/thorchain/thornode/bifrost/pkg/chainclients/ethereum/types"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

const DefaultObserverLevelDBFolder = `observer_data`

// BlockScanner is to scan the blocks
type BlockScanner struct {
	cfg        config.BlockScannerConfiguration
	logger     zerolog.Logger
	httpClient *http.Client
	db         blockscanner.ScannerStorage
	m          *metrics.Metrics
	errCounter *prometheus.CounterVec
	client     *ethclient.Client
}

// NewBlockScanner create a new instance of BlockScan
func NewBlockScanner(cfg config.BlockScannerConfiguration, scanStorage blockscanner.ScannerStorage, isTestNet bool, client *ethclient.Client, m *metrics.Metrics) (*BlockScanner, error) {
	if scanStorage == nil {
		return nil, errors.New("scanStorage is nil")
	}
	if m == nil {
		return nil, errors.New("metrics is nil")
	}

	return &BlockScanner{
		cfg:        cfg,
		logger:     log.Logger.With().Str("module", "blockscanner").Str("chain", "ethereum").Logger(),
		db:         scanStorage,
		errCounter: m.GetCounterVec(metrics.BlockScanError(common.ETHChain)),
		client:     client,
		httpClient: &http.Client{
			Timeout: cfg.HttpRequestTimeout,
		},
	}, nil
}

// processBlock extracts transactions from block
func (e *BlockScanner) processBlock(block blockscanner.Block) (stypes.TxIn, error) {
	noTx := stypes.TxIn{}

	strBlock := strconv.FormatInt(block.Height, 10)
	if err := e.db.SetBlockScanStatus(block, blockscanner.Processing); err != nil {
		e.errCounter.WithLabelValues("fail_set_block_status", strBlock).Inc()
		return noTx, errors.Wrapf(err, "fail to set block scan status for block %d", block.Height)
	}

	e.logger.Debug().Int64("block", block.Height).Int("txs", len(block.Txs)).Msg("txs")
	if len(block.Txs) == 0 {
		e.m.GetCounter(metrics.BlockWithoutTx("ETH")).Inc()
		e.logger.Debug().Int64("block", block.Height).Msg("there are no txs in this block")
		return noTx, nil
	}

	// TODO: add block txs logic here

	return noTx, nil
}

func (e *BlockScanner) FetchTxs(height int64) (stypes.TxIn, error) {
	rawTxs, err := e.getRPCBlock(height)
	if err != nil {
		return stypes.TxIn{}, err
	}

	block := blockscanner.Block{Height: height, Txs: rawTxs}
	e.logger.Debug().Int64("block", block.Height).Msg("processing block")
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

func (e *BlockScanner) getFromHttp(url, body string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, strings.NewReader(body))
	if err != nil {
		e.errCounter.WithLabelValues("fail_create_http_request", url).Inc()
		return nil, errors.Wrap(err, "fail to create http request")
	}
	if len(body) > 0 {
		req.Header.Add("Content-Type", "application/json")
	}
	resp, err := e.httpClient.Do(req)
	if err != nil {
		e.errCounter.WithLabelValues("fail_send_http_request", url).Inc()
		return nil, errors.Wrapf(err, "fail to get from %s ", url)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			e.logger.Error().Err(err).Msg("fail to close http response body.")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		e.errCounter.WithLabelValues("unexpected_status_code", resp.Status).Inc()
		return nil, errors.Errorf("unexpected status code:%d from %s", resp.StatusCode, url)
	}
	return ioutil.ReadAll(resp.Body)
}

func (e *BlockScanner) getRPCBlock(height int64) ([]string, error) {
	start := time.Now()
	defer func() {
		if err := recover(); err != nil {
			e.logger.Error().Msgf("fail to get RPCBlock:%s", err)
		}
		duration := time.Since(start)
		e.m.GetHistograms(metrics.BlockDiscoveryDuration).Observe(duration.Seconds())
	}()
	body := e.BlockRequest(height)
	buf, err := e.getFromHttp(e.cfg.RPCHost, body)
	if err != nil {
		e.errCounter.WithLabelValues("fail_get_block", e.cfg.RPCHost).Inc()
		time.Sleep(300 * time.Millisecond)
		return nil, err
	}

	rawTxns, err := e.UnmarshalBlock(buf)
	if err != nil {
		e.errCounter.WithLabelValues("fail_unmarshal_block", e.cfg.RPCHost).Inc()
	}
	return rawTxns, err
}

func (e *BlockScanner) BlockRequest(height int64) string {
	return `{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x` + fmt.Sprintf("%x", height) + `", true],"id":1}`
}

func (e *BlockScanner) UnmarshalBlock(buf []byte) ([]string, error) {
	var head *types.Header
	var body etypes.RPCBlock
	if err := json.Unmarshal(buf, &head); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(buf, &body); err != nil {
		return nil, err
	}
	txs := make([]string, 0)
	for _, tx := range body.Transactions {
		bytes, err := tx.Transaction.MarshalJSON()
		if err != nil {
			return nil, errors.Wrap(err, "fail to unmarshal tx from block")
		}
		txs = append(txs, string(bytes))
	}
	return txs, nil
}
