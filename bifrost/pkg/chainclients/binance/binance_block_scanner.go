package binance

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/binance-chain/go-sdk/common/types"
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
	http               *http.Client
	singleFee          uint64
	multiFee           uint64
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
	commonBlockScanner, err := blockscanner.NewCommonBlockScanner(cfg, startBlockHeight, scanStorage, m, blockscanner.CosmosSupplemental{})
	if err != nil {
		return nil, errors.Wrap(err, "fail to create common block scanner")
	}
	if isTestNet {
		types.Network = types.TestNetwork
	} else {
		types.Network = types.ProdNetwork
	}

	netClient := &http.Client{
		Timeout: time.Second * 10,
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
		http:               netClient,
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
