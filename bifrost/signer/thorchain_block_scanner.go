package signer

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	btypes "gitlab.com/thorchain/thornode/bifrost/blockscanner/types"
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
	if err := b.processTxOutBlock(height); err != nil {
		time.Sleep(b.cfg.BlockHeightDiscoverBackoff)
		return stypes.TxIn{}, err
	}
	if err := b.processKeygenBlock(height); err != nil {
		time.Sleep(b.cfg.BlockHeightDiscoverBackoff)
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
	if keygen.Height == 0 {
		return btypes.UnavailableBlock
	}

	if len(keygen.Keygens) > 0 {
		b.keygenChan <- keygen
	}
	return nil
}

func (b *ThorchainBlockScan) processTxOutBlock(blockHeight int64) error {
	for _, pk := range b.pubkeyMgr.GetSignPubKeys() {
		if len(pk.String()) == 0 {
			continue
		}
		tx, err := b.thorchain.GetKeysign(blockHeight, pk.String())
		if err != nil {
			if errors.Is(err, btypes.UnavailableBlock) {
				// custom error (to be dropped and not logged) because the block is
				// available yet
				return btypes.UnavailableBlock
			}
			return fmt.Errorf("fail to get keysign from block scanner: %w", err)
		}

		for c, out := range tx.Chains {
			b.logger.Debug().Str("chain", c.String()).Msg("chain")
			if len(out.TxArray) == 0 {
				b.logger.Debug().Int64("block", blockHeight).Msg("nothing to process")
				b.m.GetCounter(metrics.BlockNoTxOut(c)).Inc()
				continue
			}
			b.txOutChan <- out
		}
	}
	return nil
}
