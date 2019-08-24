package signer

import (
	"sync"
	"time"

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
	stateChainBlockScanner, err := NewStateChainBlockScan(cfg.BlockScanner, stateChainScanStorage, cfg.StateChain.ChainHost)
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
	for idx := 1; idx <= s.cfg.MessageProcessor; idx++ {
		s.wg.Add(1)
		go s.processTxnOut(s.stateChainBlockScanner.GetMessages(), idx)
	}
	if err := s.retryAll(); nil != err {
		return errors.Wrap(err, "fail to retry txouts")
	}
	s.wg.Add(1)
	go s.retryFailedTxOutProcessor()
	return s.stateChainBlockScanner.Start()
}

func (s *Signer) retryAll() error {
	txOuts, err := s.storage.GetTxOutsForRetry(false)
	if nil != err {
		return errors.Wrap(err, "fail to get txout for retry")
	}
	s.logger.Info().Msgf("we find (%d) txOut need to be retry, retrying now", len(txOuts))
	if err := s.retryTxOut(txOuts); nil != err {
		return errors.Wrap(err, "fail to retry txouts")
	}
	return nil
}
func (s *Signer) retryTxOut(txOuts []types.TxOut) error {
	if len(txOuts) == 0 {
		return nil
	}
	for _, item := range txOuts {
		select {
		case <-s.stopChan:
			return nil
		default:
			if err := s.signAndSendToBinanceChain(item); nil != err {
				s.logger.Error().Err(err).Str("height", item.Height).Msg("fail to sign and send it to binance chain")
				continue
			}
			if err := s.storage.RemoveTxOut(item); err != nil {
				s.logger.Error().Err(err).Msg("fail to remove txout from local storage")
				return errors.Wrap(err, "fail to remove txout from local storage")
			}
		}
	}
	return nil
}
func (s *Signer) retryFailedTxOutProcessor() {
	s.logger.Info().Msg("start retry process")
	defer s.logger.Info().Msg("stop retry process")
	defer s.wg.Done()
	// retry all
	t := time.NewTicker(s.cfg.RetryInterval)
	defer t.Stop()
	for {
		select {
		case <-s.stopChan:
			return
		case <-t.C:
			txOuts, err := s.storage.GetTxOutsForRetry(true)
			if nil != err {
				s.logger.Error().Err(err).Msg("fail to get txout for retry")
				continue
			}
			if err := s.retryTxOut(txOuts); nil != err {
				s.logger.Error().Err(err).Msg("fail to retry Txouts")
			}
		}
	}
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
			s.logger.Info().Msgf("Received a TxOut Array of %v from the StateChain", txOut)
			if err := s.storage.SetTxOutStatus(txOut, Processing); nil != err {
				s.logger.Error().Err(err).Msg("fail to update txout local storage")
				// raise alert
				return
			}
			if err := s.signAndSendToBinanceChain(txOut); nil != err {
				s.logger.Error().Err(err).Msg("fail to send txout to binance chain, will retry later")
				if err := s.storage.SetTxOutStatus(txOut, Failed); nil != err {
					s.logger.Error().Err(err).Msg("fail to update txout local storage")
					// raise alert
					return
				}
			}
			if err := s.storage.RemoveTxOut(txOut); nil != err {
				s.logger.Error().Err(err).Msg("fail to remove txout from local store")
			}
		}

	}
}
func (s *Signer) signAndSendToBinanceChain(txOut types.TxOut) error {
	hexTx, param, err := s.Binance.SignTx(txOut)
	if nil != err {
		s.logger.Error().Err(err).Msg("fail to sign txOut")
	}
	if nil == hexTx {
		// nothing need to be send
		return nil
	}

	log.Info().Msgf("Generated a signature for Binance: %s", string(hexTx))
	commitResult, err := s.Binance.BroadcastTx(hexTx, param)
	if nil != err {
		s.logger.Error().Err(err).Msg("fail to broadcast a tx to binance chain")
		return errors.Wrap(err, "fail to broadcast a tx to binance chain")
	}
	s.logger.Debug().
		Str("hash", commitResult.Hash).
		Msg("signed and send to binance chain successfully")
	return nil
}

// Stop the signer process
func (s *Signer) Stop() error {
	s.logger.Info().Msg("receive request to stop signer")
	defer s.logger.Info().Msg("signer stopped successfully")
	close(s.stopChan)
	s.wg.Wait()
	return s.storage.Close()
}
