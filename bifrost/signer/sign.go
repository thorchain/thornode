package signer

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/binance-chain/go-sdk/keys"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/bepswap/thornode/bifrost/binance"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/config"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/metrics"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/statechain/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
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
	m                      *metrics.Metrics
	errCounter             *prometheus.CounterVec
}

// NewSigner create a new instance of signer
func NewSigner(cfg config.SignerConfiguration) (*Signer, error) {
	stateChainScanStorage, err := NewStateChanBlockScannerStorage(cfg.SignerDbPath)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create statechain scan storage")
	}
	m, err := metrics.NewMetrics(cfg.Metric)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create metric instance")
	}

	// Create pubkey manager and add our private key (Yggdrasil pubkey)
	pkm := NewPubKeyManager()
	km, err := keys.NewPrivateKeyManager(cfg.Binance.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create private key manager")
	}
	pkm.Add(common.PubKey(km.GetPrivKey().PubKey().Bytes()))

	stateChainBlockScanner, err := NewStateChainBlockScan(cfg.BlockScanner, stateChainScanStorage, cfg.StateChain.ChainHost, m, pkm)
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
		m:                      m,
		storage:                stateChainScanStorage,
		errCounter:             m.GetCounterVec(metrics.SignerError),
	}, nil
}

func (s *Signer) Start() error {
	s.wg.Add(1)
	go s.processTxnOut(s.stateChainBlockScanner.GetMessages(), 1)
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
		s.errCounter.WithLabelValues("fail_get_txout_for_retry", "").Inc()
		return errors.Wrap(err, "fail to get txout for retry")
	}
	s.logger.Info().Msgf("we find (%d) txOut need to be retry, retrying now", len(txOuts))
	if err := s.retryTxOut(txOuts); nil != err {
		s.errCounter.WithLabelValues("fail_retry_txout", "").Inc()
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
			if !item.Chain.Equals(common.BNBChain) {
				s.logger.Debug().Str("chain", item.Chain.String()).
					Msg("not binance chain , we don't sign it")
				continue
			}

			if err := s.signTxOutAndSendToBinanceChain(item); nil != err {
				s.errCounter.WithLabelValues("fail_sign_send_to_binance", item.Height).Inc()
				s.logger.Error().Err(err).Str("height", item.Height).Msg("fail to sign and send it to binance chain")
				continue
			}
			if err := s.storage.RemoveTxOut(item); err != nil {
				s.errCounter.WithLabelValues("fail_remove_txout_from_local", item.Height).Inc()
				s.logger.Error().Err(err).Msg("fail to remove txout from local storage")
				return errors.Wrap(err, "fail to remove txout from local storage")
			}
		}
	}
	return nil
}

func (s *Signer) shouldSign(tai types.TxArrayItem) bool {
	binanceAddr := s.Binance.GetAddress()
	s.logger.Info().Str("address", binanceAddr).Msg("current signer address")
	pubKey, err := common.NewPubKeyFromHexString(tai.PoolAddress.String())
	if nil != err {
		s.logger.Error().Err(err).Msg("fail to parse pool address")
		return false
	}
	address, err := pubKey.GetAddress(common.BNBChain)
	if nil != err {
		s.logger.Error().Err(err).Msg("fail to get address")
		return false
	}
	s.logger.Info().Str("address", address.String()).Msg("")
	if strings.EqualFold(binanceAddr, address.String()) {
		return true
	}

	return false
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
				s.errCounter.WithLabelValues("fail_get_txout_for_retry", "").Inc()
				s.logger.Error().Err(err).Msg("fail to get txout for retry")
				continue
			}
			if err := s.retryTxOut(txOuts); nil != err {
				s.errCounter.WithLabelValues("fail_retry_txout", "").Inc()
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
				s.errCounter.WithLabelValues("fail_update_txout_local", txOut.Height).Inc()
				s.logger.Error().Err(err).Msg("fail to update txout local storage")
				// raise alert
				return
			}

			if err := s.signTxOutAndSendToBinanceChain(txOut); nil != err {
				s.errCounter.WithLabelValues("fail_sign_send_to_binance", txOut.Height).Inc()
				s.logger.Error().Err(err).Msg("fail to send txout to binance chain, will retry later")
				if err := s.storage.SetTxOutStatus(txOut, Failed); nil != err {
					s.errCounter.WithLabelValues("fail_update_txout_local", txOut.Height).Inc()
					s.logger.Error().Err(err).Msg("fail to update txout local storage")
					// raise alert
					return
				}
			}
			if err := s.storage.RemoveTxOut(txOut); nil != err {
				s.errCounter.WithLabelValues("fail_remove_txout_local", txOut.Height).Inc()
				s.logger.Error().Err(err).Msg("fail to remove txout from local store")
			}
		}

	}
}

// signAndSendToBinanceChainWithRetry retry a few times before we move on to he next block
func (s *Signer) signTxOutAndSendToBinanceChain(txOut types.TxOut) error {
	// most case , there should be only one item in txOut.TxArray, but sometimes there might be more than one
	// especially when we get populate , more and more transactions
	for _, item := range txOut.TxArray {
		if !s.shouldSign(item) {
			s.logger.Debug().
				Str("signer_address", s.Binance.GetAddress()).
				Msg("different pool address, ignore")
			continue
		}
		height, err := strconv.ParseInt(txOut.Height, 10, 64)
		if nil != err {
			return errors.Wrapf(err, "fail to parse block height: %s ", txOut.Height)
		}
		err = s.signAndSendToBinanceChain(item, height)
		if nil == err {
			return nil
		}
		s.logger.Error().Err(err).Int("try", 1).Msg("fail to send to binance chain")
		return fmt.Errorf("fail to send to binance chain,err:%w", err)
	}
	return nil
}

func (s *Signer) signAndSendToBinanceChain(tai types.TxArrayItem, height int64) error {
	start := time.Now()
	defer func() {
		s.m.GetHistograms(metrics.SignAndBroadcastToBinanceDuration).Observe(time.Since(start).Seconds())
	}()
	strHeight := strconv.FormatInt(height, 10)
	hexTx, param, err := s.Binance.SignTx(tai, height)
	if nil != err {
		s.errCounter.WithLabelValues("fail_sign_txout", strHeight).Inc()
		s.logger.Error().Err(err).Msg("fail to sign txOut")
	}
	if nil == hexTx {
		// nothing need to be send
		return nil
	}
	s.m.GetCounter(metrics.TxToBinanceSigned).Inc()
	log.Info().Msgf("Generated a signature for Binance: %s", string(hexTx))
	commitResult, err := s.Binance.BroadcastTx(hexTx, param)
	if nil != err {
		s.errCounter.WithLabelValues("fail_broadcast_txout", strHeight).Inc()
		s.logger.Error().Err(err).Msg("fail to broadcast a tx to binance chain")
		return errors.Wrap(err, "fail to broadcast a tx to binance chain")
	}
	s.logger.Debug().
		Str("hash", commitResult.Hash).
		Msg("signed and send to binance chain successfully")
	s.m.GetCounter(metrics.TxToBinanceSignedBroadcast).Inc()
	return nil
}

// Stop the signer process
func (s *Signer) Stop() error {
	s.logger.Info().Msg("receive request to stop signer")
	defer s.logger.Info().Msg("signer stopped successfully")
	close(s.stopChan)
	s.wg.Wait()
	if err := s.m.Stop(); nil != err {
		s.logger.Error().Err(err).Msg("fail to stop metric server")
	}
	return s.storage.Close()
}
