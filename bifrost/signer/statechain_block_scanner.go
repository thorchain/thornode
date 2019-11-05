package signer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/thorchain/bepswap/thornode/common"

	"gitlab.com/thorchain/bepswap/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/config"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/metrics"
	stypes "gitlab.com/thorchain/bepswap/thornode/bifrost/statechain/types"
)

type StateChainBlockScan struct {
	logger             zerolog.Logger
	wg                 *sync.WaitGroup
	stopChan           chan struct{}
	txOutChan          chan stypes.TxOut
	cfg                config.BlockScannerConfiguration
	scannerStorage     blockscanner.ScannerStorage
	commonBlockScanner *blockscanner.CommonBlockScanner
	chainHost          string
	m                  *metrics.Metrics
	errCounter         *prometheus.CounterVec
}

// NewStateChainBlockScan create a new instance of statechain block scanner
func NewStateChainBlockScan(cfg config.BlockScannerConfiguration, scanStorage blockscanner.ScannerStorage, chainHost string, m *metrics.Metrics) (*StateChainBlockScan, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("rpc host is empty")
	}
	if nil == scanStorage {
		return nil, errors.New("scanStorage is nil")
	}
	if nil == m {
		return nil, errors.New("metric is nil")
	}
	commonBlockScanner, err := blockscanner.NewCommonBlockScanner(cfg, scanStorage, m)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create common block scanner")
	}
	return &StateChainBlockScan{
		logger:             log.With().Str("module", "statechainblockscanner").Logger(),
		wg:                 &sync.WaitGroup{},
		stopChan:           make(chan struct{}),
		txOutChan:          make(chan stypes.TxOut),
		cfg:                cfg,
		scannerStorage:     scanStorage,
		commonBlockScanner: commonBlockScanner,
		chainHost:          chainHost,
		errCounter:         m.GetCounterVec(metrics.StateChainBlockScanError),
	}, nil
}

// GetMessages return the channel
func (b *StateChainBlockScan) GetMessages() <-chan stypes.TxOut {
	return b.txOutChan
}

// Start to scan blocks
func (b *StateChainBlockScan) Start() error {
	b.wg.Add(1)
	go b.processBlocks(1)
	b.commonBlockScanner.Start()
	return nil
}

func (b *StateChainBlockScan) processABlock(blockHeight int64) error {
	uri := url.URL{
		Scheme: "http",
		Host:   b.chainHost,
		Path:   fmt.Sprintf("/thorchain/txoutarray/%v", blockHeight),
	}
	strBlockHeight := strconv.FormatInt(blockHeight, 10)
	buf, err := b.commonBlockScanner.GetFromHttpWithRetry(uri.String())
	if nil != err {
		b.errCounter.WithLabelValues("fail_get_tx_out", strBlockHeight)
		return errors.Wrap(err, "fail to get tx out from a block")
	}

	type txOut struct {
		Chains map[common.Chain]stypes.TxOut `json:"chains"`
	}

	var tx txOut
	if err := json.Unmarshal(buf, &tx); err != nil {
		b.errCounter.WithLabelValues("fail_unmarshal_tx_out", strBlockHeight)
		return errors.Wrap(err, "fail to unmarshal TxOut")
	}
	for c, out := range tx.Chains {
		b.logger.Debug().Str("chain", c.String()).Msg("chain")
		if len(out.TxArray) == 0 {
			b.logger.Debug().Int64("block", blockHeight).Msg("nothing to process")
			b.m.GetCounter(metrics.BlockNoTxOut).Inc()
			return nil
		}
		// TODO here we will need to dispatch to different chain processor
		b.txOutChan <- out
	}
	return nil
}

func (b *StateChainBlockScan) processBlocks(idx int) {
	b.logger.Debug().Int("idx", idx).Msg("start searching tx out in a block")
	defer b.logger.Debug().Int("idx", idx).Msg("stop searching tx out in a block")
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
			if err := b.processABlock(block); nil != err {
				if errStatus := b.scannerStorage.SetBlockScanStatus(block, blockscanner.Failed); nil != errStatus {
					b.errCounter.WithLabelValues("fail_set_block_Status", strconv.FormatInt(block, 10))
					b.logger.Error().Err(err).Int64("height", block).Msg("fail to set block to fail status")
				}
				b.errCounter.WithLabelValues("fail_search_tx", strconv.FormatInt(block, 10))
				b.logger.Error().Err(err).Int64("height", block).Msg("fail to search tx in block")
				// we will have a retry go routine to check it.
				continue
			}
			// set a block as success
			if err := b.scannerStorage.RemoveBlockStatus(block); nil != err {
				b.errCounter.WithLabelValues("fail_remove_block_Status", strconv.FormatInt(block, 10))
				b.logger.Error().Err(err).Int64("block", block).Msg("fail to remove block status from data store, thus block will be re processed")
			}
		}
	}
}

// Stop the scanner
func (b *StateChainBlockScan) Stop() error {
	b.logger.Info().Msg("received request to stop state chain block scanner")
	defer b.logger.Info().Msg("statechain block scanner stopped successfully")
	close(b.stopChan)
	b.wg.Wait()
	return nil
}
