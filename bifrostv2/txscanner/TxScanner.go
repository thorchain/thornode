package txscanner

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/addressmanager"
	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/thorclient"
	"gitlab.com/thorchain/thornode/bifrostv2/txscanner/types"
)

type TxScanner struct {
	cfg            config.TxScannerConfigurations
	logger         zerolog.Logger
	stopChan       chan struct{}
	addressManager *addressmanager.AddressManager
	thorClient     *thorclient.Client
	chains         []BlockChainClients
	wg             sync.WaitGroup
	closeOnce      sync.Once
}

type BlockChainClients interface {
	Start(txInChan chan<- types.TxIn, startHeight types.FnLastScannedBlockHeight) error
	Stop() error
}

func NewTxScanner(cfg config.TxScannerConfigurations, addressManager *addressmanager.AddressManager, thorClient *thorclient.Client) *TxScanner {
	return &TxScanner{
		logger:         log.Logger.With().Str("module", "txScanner").Logger(),
		cfg:            cfg,
		stopChan:       make(chan struct{}),
		addressManager: addressManager,
		thorClient:     thorClient,
		wg:             sync.WaitGroup{},
		chains:         loadChains(cfg),
	}
}

func (s *TxScanner) Start() error {
	txInChan := make(chan types.TxIn)

	for _, chain := range s.chains {
		err := chain.Start(txInChan, s.thorClient.GetLastObservedInHeight)
		if err != nil {
			s.logger.Err(err).Msg("failed to start chain")
			continue
		}
		s.wg.Add(1)
		go s.processTxIns(txInChan)
	}
	return nil
}

func (s *TxScanner) Stop() error {
	for _, chain := range s.chains {
		if err := chain.Stop(); err != nil {
			s.logger.Err(err).Msg("failed to stop chain")
		}
	}
	s.closeOnce.Do(func() {
		close(s.stopChan)
	})
	s.wg.Wait()
	s.logger.Info().Msg("stopped TxScanner")
	return nil
}

func (s *TxScanner) processTxIns(ch <-chan types.TxIn) {
	s.logger.Info().Msg("started processTxIns")
	defer s.logger.Info().Msg("stopped processTxIns")
	defer s.wg.Done()
	for {
		select {
		case <-s.stopChan:
			return
		case txIn, more := <-ch:
			fmt.Printf("%v: %v: %v\n", txIn.Chain.String(), txIn.BlockHeight, txIn.BlockHash)
			if !more {
				// channel closed
				return
			}
			// if len(txIn.TxArray) == 0 {
			// 	s.logger.Debug().Msg("nothing to be forward to thorchain")
			// 	continue
			// }
			s.processTxIn(txIn)
		}
	}
}

func (s *TxScanner) processTxIn(txIn types.TxIn) {

}
