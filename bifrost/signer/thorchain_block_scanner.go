package signer

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	ttypes "gitlab.com/thorchain/thornode/x/thorchain/types"
)

type ThorchainBlockScan struct {
	logger         zerolog.Logger
	wg             *sync.WaitGroup
	stopChan       chan struct{}
	txOutChan      chan types.TxOut
	keygenChan     chan ttypes.KeygenBlock
	cfg            config.BlockScannerConfiguration
	scannerStorage blockscanner.ScannerStorage
	thorchain      *thorclient.ThorchainBridge
	m              *metrics.Metrics
	errCounter     *prometheus.CounterVec
	pubkeyMgr      pubkeymanager.PubKeyValidator
	cdc            *codec.Codec
}

// NewThorchainBlockScan create a new instance of thorchain block scanner
func NewThorchainBlockScan(cfg config.BlockScannerConfiguration, scanStorage blockscanner.ScannerStorage, thorchain *thorclient.ThorchainBridge, m *metrics.Metrics, pubkeyMgr pubkeymanager.PubKeyValidator) (*ThorchainBlockScan, error) {
	if scanStorage == nil {
		return nil, errors.New("scanStorage is nil")
	}
	if m == nil {
		return nil, errors.New("metric is nil")
	}
	return &ThorchainBlockScan{
		logger:         log.With().Str("module", "thorchainblockscanner").Logger(),
		wg:             &sync.WaitGroup{},
		stopChan:       make(chan struct{}),
		txOutChan:      make(chan types.TxOut),
		keygenChan:     make(chan ttypes.KeygenBlock),
		cfg:            cfg,
		scannerStorage: scanStorage,
		thorchain:      thorchain,
		errCounter:     m.GetCounterVec(metrics.ThorchainBlockScannerError),
		pubkeyMgr:      pubkeyMgr,
		cdc:            codec.New(),
	}, nil
}

// GetMessages return the channel
func (b *ThorchainBlockScan) GetTxOutMessages() <-chan types.TxOut {
	return b.txOutChan
}

func (b *ThorchainBlockScan) GetKeygenMessages() <-chan ttypes.KeygenBlock {
	return b.keygenChan
}

func (b *ThorchainBlockScan) FetchTxs(height int64) (stypes.TxIn, error) {
	fmt.Printf("<<<< Block: %d\n", height)
	if err := b.processTxOutBlock(height); err != nil {
		fmt.Println("ERR1")
		return stypes.TxIn{}, err
	}
	if err := b.processKeygenBlock(height); err != nil {
		fmt.Println("ERR2")
		return stypes.TxIn{}, err
	}
	return stypes.TxIn{}, nil
}

func (b *ThorchainBlockScan) processKeygenBlock(blockHeight int64) error {
	pk := b.pubkeyMgr.GetNodePubKey()
	keygen, err := b.thorchain.GetKeygenBlock(blockHeight, pk.String())
	if err != nil {
		return fmt.Errorf("fail to get keygen from thorchain: %w", err)
	}

	// custom error (to be dropped and not logged) because the block is
	// available yet
	if keygen == nil {
		return errors.New("")
	}

	b.keygenChan <- *keygen
	return nil
}

func (b *ThorchainBlockScan) processTxOutBlock(blockHeight int64) error {
	for _, pk := range b.pubkeyMgr.GetSignPubKeys() {
		if len(pk.String()) == 0 {
			continue
		}
		tx, err := b.thorchain.GetKeysign(blockHeight, pk.String())
		if err != nil {
			if errors.Is(err, thorclient.ErrNotFound) {
				// custom error (to be dropped and not logged) because the block is
				// available yet
				return errors.New("")
			}
			return fmt.Errorf("fail to get keysign from block scanner: %w", err)
		}

		for c, out := range tx.Chains {
			b.logger.Debug().Str("chain", c.String()).Msg("chain")
			if len(out.TxArray) == 0 {
				b.logger.Debug().Int64("block", blockHeight).Msg("nothing to process")
				b.m.GetCounter(metrics.BlockNoTxOut(c)).Inc()
				return nil
			}
			b.txOutChan <- out
		}
	}
	return nil
}

/*
func (b *ThorchainBlockScan) processBlocks(idx int) {
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
			b.logger.Debug().Int64("block", block.Height).Msg("processing block")
			if err := b.processTxOutBlock(block.Height); err != nil {
				if errStatus := b.scannerStorage.SetBlockScanStatus(block, blockscanner.Failed); errStatus != nil {
					b.errCounter.WithLabelValues("fail_set_block_Status", strconv.FormatInt(block.Height, 10))
					b.logger.Error().Err(err).Int64("height", block.Height).Msg("fail to set block to fail status")
				}
				// the error is blank, which means its an error we skip logging
				if err.Error() == "" {
					continue
				}
				b.errCounter.WithLabelValues("fail_search_tx", strconv.FormatInt(block.Height, 10))
				b.logger.Error().Err(err).Int64("height", block.Height).Msg("fail to search tx in block")
				// THORNode will have a retry go routine to check it.
				continue
			}

			// set a block as success
			if err := b.scannerStorage.RemoveBlockStatus(block.Height); err != nil {
				b.errCounter.WithLabelValues("fail_remove_block_Status", strconv.FormatInt(block.Height, 10))
				b.logger.Error().Err(err).Int64("block", block.Height).Msg("fail to remove block status from data store, thus block will be re processed")
			}

			// Intentionally not covering this before the block is marked as
			// success. This is because we don't care if keygen is successful
			// or not.
			b.logger.Debug().Int64("block", block.Height).Msg("processing keygen block")
			if err := b.processKeygenBlock(block.Height); err != nil {
				b.errCounter.WithLabelValues("fail_process_keygen", strconv.FormatInt(block.Height, 10))
				b.logger.Error().Err(err).Int64("height", block.Height).Msg("fail to process keygen")
			}
		}
	}
}
*/
