package ethereum

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

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

	// TODO: add block txs logic here

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
