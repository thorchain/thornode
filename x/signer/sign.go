package signer

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/bepswap/observe/config"
	"gitlab.com/thorchain/bepswap/observe/x/binance"
	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

// Signer will pull the tx out from statechain and then forward it to binance chain
type Signer struct {
	logger                 zerolog.Logger
	cfg                    config.SignerConfiguration
	wg                     *sync.WaitGroup
	stopChan               chan struct{}
	stateChainBlockScanner *StateChainBlockScan
	Binance                *binance.Binance
	storage                *StateChanBlockScannerStorage
}

// NewSigner create a new instance of signer
func NewSigner(cfg config.SignerConfiguration) (*Signer, error) {
	stateChainScanStorage, err := NewStateChanBlockScannerStorage(cfg.SignerDbPath)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create statechain scan storage")
	}
	stateChainBlockScanner, err := NewStateChainBlockScan(cfg.BlockScannerConfiguration, stateChainScanStorage, cfg.StateChainConfiguration.ChainHost)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create state chain block scan")
	}
	b, err := binance.NewBinance(cfg.Binance)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create binance client")
	}
	return &Signer{
		logger:                 log.With().Str("module", "signer").Logger(),
		cfg:                    cfg,
		wg:                     &sync.WaitGroup{},
		stopChan:               make(chan struct{}),
		stateChainBlockScanner: stateChainBlockScanner,
		Binance:                b,
		storage:                stateChainScanStorage,
	}, nil
}

func (s *Signer) Start() error {
	for idx := s.cfg.MessageProcessor; idx <= s.cfg.MessageProcessor*2; idx++ {
		s.wg.Add(1)
		go s.processTxnOut(s.stateChainBlockScanner.GetMessages(), idx)
	}
	return s.stateChainBlockScanner.Start()
}

func (s *Signer) processTxnOut(ch <-chan types.TxOut, idx int) {
	s.logger.Info().Int("idx", idx).Msg("start to process tx out")
	defer s.logger.Info().Int("idx", idx).Msg("stop to process tx out")
	defer s.wg.Done()
	for {
		select {
		case <-s.stopChan:
			return
		case txOut, more := <-ch:
			if !more {
				return
			}
			log.Info().Msgf("Received a TxOut Array of %v from the StateChain", txOut)
			hexTx, param := s.Binance.SignTx(txOut)
			log.Info().Msgf("Generated a signature for Binance: %s", string(hexTx))

			_, _ = s.Binance.BroadcastTx(hexTx, param)
		}

	}
}

// Stop the signer process
func (s *Signer) Stop() error {
	s.logger.Info().Msg("receive request to stop signer")
	defer s.logger.Info().Msg("signer stopped successfully")
	close(s.stopChan)
	s.wg.Wait()
	return s.storage.Close()
}
