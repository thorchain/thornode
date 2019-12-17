package signer

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrost/binance"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
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
	tssKeygen              *tss.KeyGen
	thorKeys               *thorclient.Keys
	pkm                    *PubKeyManager
}

// NewSigner create a new instance of signer
func NewSigner(cfg config.SignerConfiguration) (*Signer, error) {
	stateChainScanStorage, err := NewStateChanBlockScannerStorage(cfg.SignerDbPath)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create thorchain scan storage")
	}
	m, err := metrics.NewMetrics(cfg.Metric)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create metric instance")
	}
	pkm := NewPubKeyManager()
	thorKeys, err := thorclient.NewKeys(cfg.StateChain.ChainHomeFolder, cfg.StateChain.SignerName, cfg.StateChain.SignerPasswd)
	if nil != err {
		return nil, fmt.Errorf("fail to load keys,err:%w", err)
	}
	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}
	na, err := thorclient.GetNodeAccount(httpClient, cfg.StateChain.ChainHost, thorKeys.GetSignerInfo().GetAddress().String())
	if nil != err {
		return nil, fmt.Errorf("fail to get node account from thorchain,err:%w", err)
	}
	if na.IsEmpty() || na.NodePubKey.Secp256k1.IsEmpty() {
		// TODO: uncommenting to test if this is break smoke tests
		// return nil, fmt.Errorf("node account: %s not ready yet", thorKeys.GetSignerInfo().GetAddress().String())
	}

	for _, item := range na.SignerMembership {
		pkm.Add(item)
	}
	pkm.Add(na.NodePubKey.Secp256k1)

	// Create pubkey manager and add our private key (Yggdrasil pubkey)
	stateChainBlockScanner, err := NewStateChainBlockScan(cfg.BlockScanner, stateChainScanStorage, cfg.StateChain.ChainHost, m, pkm)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create state chain block scan")
	}
	b, err := binance.NewBinance(cfg.Binance, cfg.UseTSS, cfg.TSS)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create binance client")
	}

	signer := &Signer{
		logger:                 log.With().Str("module", "signer").Logger(),
		cfg:                    cfg,
		wg:                     &sync.WaitGroup{},
		stopChan:               make(chan struct{}),
		stateChainBlockScanner: stateChainBlockScanner,
		Binance:                b,
		m:                      m,
		storage:                stateChainScanStorage,
		errCounter:             m.GetCounterVec(metrics.SignerError),
		pkm:                    pkm,
	}

	if cfg.UseTSS {
		kg, err := tss.NewTssKeyGen(cfg.TSS, cfg.StateChain, thorKeys)
		if nil != err {
			return nil, fmt.Errorf("fail to create Tss Key gen,err:%w", err)
		}
		signer.tssKeygen = kg
	}
	return signer, nil
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
	s.logger.Info().Msgf("THORNode find (%d) txOut need to be retry, retrying now", len(txOuts))
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
					Msg("not binance chain , THORNode don't sign it")
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
	return s.pkm.HasKey(tai.VaultPubKey)
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

const nextPoolPrefix = `nextpool`

func (s *Signer) processTssKeyGenCeremony(tai types.TxArrayItem) (types.TxArrayItem, error) {
	if !strings.HasPrefix(tai.Memo, nextPoolPrefix) {
		return tai, nil
	}
	pubKey, err := s.tssKeygen.GenerateNewKey()
	if nil != err {
		return tai, fmt.Errorf("fail to generate new pool pub key,err:%w", err)
	}
	if pubKey.IsEmpty() {
		return tai, fmt.Errorf("fail to generate new pool pub key")
	}
	s.pkm.Add(pubKey.Secp256k1)
	tai.Memo = fmt.Sprintf("%s:%s", nextPoolPrefix, pubKey.Secp256k1.String())
	addr, err := pubKey.GetAddress(common.BNBChain)
	if nil != err {
		return tai, fmt.Errorf("fail to get address,err:%w", err)
	}
	tai.To = addr.String()
	return tai, nil
}

// signAndSendToBinanceChainWithRetry retry a few times before THORNode move on to he next block
func (s *Signer) signTxOutAndSendToBinanceChain(txOut types.TxOut) error {
	// most case , there should be only one item in txOut.TxArray, but sometimes there might be more than one
	// especially when THORNode get populate , more and more transactions
	for _, item := range txOut.TxArray {
		height, err := strconv.ParseInt(txOut.Height, 10, 64)
		if nil != err {
			return errors.Wrapf(err, "fail to parse block height: %s ", txOut.Height)
		}
		if s.tssKeygen != nil && strings.HasPrefix(item.Memo, nextPoolPrefix) {
			tai, err := s.processTssKeyGenCeremony(item)
			if nil != err {
				return fmt.Errorf("fail to get get next pool address,err:%w", err)
			}
			item = tai
		}

		if !s.shouldSign(item) {
			s.logger.Info().
				Str("signer_address", s.Binance.GetAddress(item.VaultPubKey)).
				Msg("different pool address, ignore")
			continue
		}
		if len(item.To) == 0 {
			s.logger.Info().Msg("To address is empty, THORNode don't know where to send the fund , ignore")
			continue
		}
		err = s.signAndSendToBinanceChain(item, height)
		if nil != err {
			s.logger.Error().Err(err).Int("try", 1).Msg("fail to send to binance chain")
			// This might happen when THORNode signed it successfully however somehow fail to broadcast to binance chain
			// given THORNode run a node locally , this should be rare let's log it and move on for now.
		}
	}
	return nil
}

func (s *Signer) signAndSendToBinanceChain(tai types.TxArrayItem, height int64) error {
	start := time.Now()
	defer func() {
		s.m.GetHistograms(metrics.SignAndBroadcastToBinanceDuration).Observe(time.Since(start).Seconds())
	}()
	if !tai.OutHash.IsEmpty() {
		s.logger.Info().Str("OutHash", tai.OutHash.String()).Msg("tx had been sent out before")
		return nil
	}
	strHeight := strconv.FormatInt(height, 10)
	hexTx, _, err := s.Binance.SignTx(tai, height)
	if nil != err {
		s.errCounter.WithLabelValues("fail_sign_txout", strHeight).Inc()
		s.logger.Error().Err(err).Msg("fail to sign txOut")
	}
	if nil == hexTx {
		s.logger.Error().Msg("nothing need to be send")
		// nothing need to be send
		return nil
	}
	s.m.GetCounter(metrics.TxToBinanceSigned).Inc()
	log.Info().Msgf("Generated a signature for Binance: %s", string(hexTx))
	err = s.Binance.BroadcastTx(hexTx)
	if nil != err {
		s.errCounter.WithLabelValues("fail_broadcast_txout", strHeight).Inc()
		s.logger.Error().Err(err).Msg("fail to broadcast a tx to binance chain")
		return errors.Wrap(err, "fail to broadcast a tx to binance chain")
	}
	s.logger.Debug().
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
