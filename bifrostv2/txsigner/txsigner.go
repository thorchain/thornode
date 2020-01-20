package txsigner

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/keys"
	"gitlab.com/thorchain/thornode/bifrostv2/metrics"
	"gitlab.com/thorchain/thornode/bifrostv2/pkg/blockclients"
	"gitlab.com/thorchain/thornode/bifrostv2/thorchain"
	"gitlab.com/thorchain/thornode/bifrostv2/tss"
	"gitlab.com/thorchain/thornode/bifrostv2/types"
	"gitlab.com/thorchain/thornode/bifrostv2/vaultmanager"
	stypes "gitlab.com/thorchain/thornode/x/thorchain/types"
)

// TxSigner represents a transaction signer
type TxSigner struct {
	cfg                   config.TxSignerConfigurations
	logger                zerolog.Logger
	chains                []blockclients.BlockChainClient
	blockInChan           chan types.Block
	stopChan              chan struct{}
	wg                    sync.WaitGroup
	vaultMgr              *vaultmanager.VaultManager
	thorchainClient       *thorchain.Client
	thorchainBlockScanner *thorchain.BlockScanner
	storage               *thorchain.BlockScannerStorage
	metrics               *metrics.Metrics
	thorchainKeys         *keys.Keys
	tssKeygen             *tss.KeyGen
	errCounter            *prometheus.CounterVec
}

// NewTxSigner instantiates TxSigner
func NewTxSigner(cfg config.TxSignerConfigurations, tssCfg config.TSSConfiguration, vaultMgr *vaultmanager.VaultManager, thorchainClient *thorchain.Client, m *metrics.Metrics) (*TxSigner, error) {
	thorchainScanStorage, err := thorchain.NewBlockScannerStorage(cfg.SignerDbPath)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create thorchain scan storage")
	}

	thorchainBlockScanner, err := thorchain.NewBlockScanner(cfg.BlockScanner, thorchainScanStorage, thorchainClient, vaultMgr, m)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create thorchain block scan")
	}

	// Retrieve thorchain keys including private
	thorchainKeys := thorchainClient.Keys()

	kg, err := tss.NewTssKeyGen(tssCfg, thorchainKeys)
	if nil != err {
		return nil, fmt.Errorf("fail to create Tss Key gen,err:%w", err)
	}

	return &TxSigner{
		logger:                log.Logger.With().Str("module", "txTxSigner").Logger(),
		cfg:                   cfg,
		thorchainClient:       thorchainClient,
		thorchainBlockScanner: thorchainBlockScanner,
		storage:               thorchainScanStorage,
		chains:                blockclients.LoadChains(cfg.BlockChains),
		blockInChan:           make(chan types.Block),
		wg:                    sync.WaitGroup{},
		vaultMgr:              vaultMgr,
		thorchainKeys:         thorchainKeys,
		metrics:               m,
		stopChan:              make(chan struct{}),
		tssKeygen:             kg,
		errCounter:            m.GetCounterVec(metrics.SignerError),
	}, nil
}

// Start starts TxSigner, listening new block on ThorChain
func (s *TxSigner) Start() error {
	s.wg.Add(1)
	go s.processTxnOut(s.thorchainBlockScanner.GetTxOutMessages(), 1)
	if err := s.retryAll(); nil != err {
		return errors.Wrap(err, "fail to retry txouts")
	}
	s.wg.Add(1)
	go s.retryFailedTxOutProcessor()

	s.wg.Add(1)
	go s.processKeygen(s.thorchainBlockScanner.GetKeygenMessages(), 1)

	return s.thorchainBlockScanner.Start()
}

func (s *TxSigner) retryAll() error {
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

func (s *TxSigner) retryTxOut(txOuts []stypes.TxOut) error {
	if len(txOuts) == 0 {
		return nil
	}
	for _, item := range txOuts {
		select {
		case <-s.stopChan:
			return nil
		default:
			if err := s.signTxOutAndBroadcast(item); nil != err {
				s.errCounter.WithLabelValues("fail_sign_and_broadcast", strconv.FormatInt(item.Height, 10)).Inc()
				s.logger.Error().Err(err).Int64("height", item.Height).Msg("fail to sign and broadcast")
				continue
			}
			if err := s.storage.RemoveTxOut(item); err != nil {
				s.errCounter.WithLabelValues("fail_remove_txout_from_local", strconv.FormatInt(item.Height, 10)).Inc()
				s.logger.Error().Err(err).Msg("fail to remove txout from local storage")
				return errors.Wrap(err, "fail to remove txout from local storage")
			}
		}
	}
	return nil
}

func (s *TxSigner) shouldSign(toi *stypes.TxOutItem) bool {
	return s.vaultMgr.HasKey(toi.VaultPubKey)
}

func (s *TxSigner) retryFailedTxOutProcessor() {
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

func (s *TxSigner) processTxnOut(ch <-chan stypes.TxOut, idx int) {
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
			strHeight := strconv.FormatInt(txOut.Height, 10)
			s.logger.Info().Msgf("Received a TxOut Array of %v from the Thorchain", txOut)
			if err := s.storage.SetTxOutStatus(txOut, thorchain.Processing); nil != err {
				s.errCounter.WithLabelValues("fail_update_txout_local", strHeight).Inc()
				s.logger.Error().Err(err).Msg("fail to update txout local storage")
				// raise alert
				return
			}

			if err := s.signTxOutAndBroadcast(txOut); nil != err {
				s.errCounter.WithLabelValues("fail_sign_and_broadcast", strHeight).Inc()
				s.logger.Error().Err(err).Msg("fail to broadcast txout to tx chain, will retry later")
				if err := s.storage.SetTxOutStatus(txOut, thorchain.Failed); nil != err {
					s.errCounter.WithLabelValues("fail_update_txout_local", strHeight).Inc()
					s.logger.Error().Err(err).Msg("fail to update txout local storage")
					// raise alert
					return
				}
			}
			if err := s.storage.RemoveTxOut(txOut); nil != err {
				s.errCounter.WithLabelValues("fail_remove_txout_local", strHeight).Inc()
				s.logger.Error().Err(err).Msg("fail to remove txout from local store")
			}
		}

	}
}

func (s *TxSigner) processKeygen(ch <-chan stypes.Keygens, idx int) {
	s.logger.Info().Int("idx", idx).Msg("start to process keygen")
	defer s.logger.Info().Int("idx", idx).Msg("stop to process keygen")
	defer s.wg.Done()
	for {
		select {
		case <-s.stopChan:
			return
		case keygens, more := <-ch:
			if !more {
				return
			}
			for _, keygen := range keygens.Keygens {
				s.logger.Info().Msgf("Received a keygen of %+v from the Thorchain", keygens)
				pubKey, err := s.tssKeygen.GenerateNewKey(keygen)
				if err != nil {
					s.errCounter.WithLabelValues("fail_to_keygen_pubkey", "").Inc()
					s.logger.Error().Err(err).Msg("fail to generate new pubkey")
					continue
				}

				if pubKey.IsEmpty() {
					continue
				}

				if err := s.thorchainClient.BroadcastKeygen(keygens.Height, pubKey.Secp256k1, keygen); err != nil {
					s.errCounter.WithLabelValues("fail_to_broadcast_keygen", "").Inc()
					s.logger.Error().Err(err).Msg("fail to broadcast keygen")
				}
			}
		}

	}
}

// signTxOutAndBroadcast retry a few times before THORNode move on to the next block
// it will try to sign and broadcast to the chain specific to the tx item
func (s *TxSigner) signTxOutAndBroadcast(txOut stypes.TxOut) error {
	// most case , there should be only one item in txOut.TxArray, but sometimes there might be more than one
	for _, tx := range txOut.TxArray {
		processed, err := s.storage.HasTxOutItem(tx, txOut.Height)
		if nil != err {
			return fmt.Errorf("fail to check against local level db: %w", err)
		}
		if processed {
			s.logger.Debug().Msgf("%+v processed already", tx)
			continue
		}

		if !s.shouldSign(tx) {
			s.logger.Info().Str("pubkey", tx.VaultPubKey.String()).Msg("tx pubkey don't match any keys from the node")
			continue
		}

		if len(tx.ToAddress) == 0 {
			s.logger.Info().Msg("To address is empty, THORNode don't know where to send the fund , ignore")
			continue
		}

		// Retrieve chain client matching tx out chain
		chain, err := s.getChainFromTxOutItem(tx)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve chain from tx out item")
		}

		// Sign against tx item chain
		signedTx, err := chain.SignTx(tx, txOut.Height)
		if err != nil {
			return errors.Wrap(err, "failed to sign tx")
		}

		// Broadcast to item chain
		err = chain.BroadcastTx(signedTx)
		if err != nil {
			return errors.Wrap(err, "failed to broadcast signed tx")
		}

		// Successfully processed tx out -> mark off from local db
		if err := s.storage.SetTxOutItem(tx, txOut.Height); nil != err {
			return fmt.Errorf("fail to mark it off from local db: %w", err)
		}
	}

	return nil
}

func (s *TxSigner) getChainFromTxOutItem(txOutItem *stypes.TxOutItem) (blockclients.BlockChainClient, error) {
	for _, chain := range s.chains {
		if chain.EqualsChain(txOutItem.Chain) {
			return chain, nil
		}
	}
	return nil, errors.New("no chain matching tx out item chain")
}

// Stop the signer process
func (s *TxSigner) Stop() error {
	s.logger.Info().Msg("receive request to stop signer")
	defer s.logger.Info().Msg("signer stopped successfully")
	close(s.stopChan)
	s.wg.Wait()
	if err := s.metrics.Stop(); nil != err {
		s.logger.Error().Err(err).Msg("fail to stop metric server")
	}
	return s.storage.Close()
}
