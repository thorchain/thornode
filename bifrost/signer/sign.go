package signer

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	stypes "github.com/binance-chain/go-sdk/common/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrost/chainclients"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain"
	ttypes "gitlab.com/thorchain/thornode/x/thorchain/types"
)

// Signer will pull the tx out from thorchain and then forward it to chain
type Signer struct {
	logger                zerolog.Logger
	cfg                   config.SignerConfiguration
	wg                    *sync.WaitGroup
	thorchainBridge       *thorclient.ThorchainBridge
	stopChan              chan struct{}
	thorchainBlockScanner *ThorchainBlockScan
	chains                map[common.Chain]chainclients.ChainClient
	storage               *ThorchainBlockScannerStorage
	m                     *metrics.Metrics
	errCounter            *prometheus.CounterVec
	tssKeygen             *tss.KeyGen
	thorKeys              *thorclient.Keys
	pubkeyMgr             pubkeymanager.PubKeyValidator
}

// NewSigner create a new instance of signer
func NewSigner(cfg config.SignerConfiguration, thorchainBridge *thorclient.ThorchainBridge, thorKeys *thorclient.Keys, pubkeyMgr pubkeymanager.PubKeyValidator, tssCfg config.TSSConfiguration, chains map[common.Chain]chainclients.ChainClient, m *metrics.Metrics) (*Signer, error) {
	thorchainScanStorage, err := NewThorchainBlockScannerStorage(cfg.SignerDbPath)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create thorchain scan storage")
	}

	var na *ttypes.NodeAccount
	for i := 0; i < 300; i++ { // wait for 5 min before timing out
		var err error
		na, err = thorchainBridge.GetNodeAccount(thorKeys.GetSignerInfo().GetAddress().String())
		if err != nil {
			return nil, fmt.Errorf("fail to get node account from thorchain,err:%w", err)
		}

		if !na.PubKeySet.Secp256k1.IsEmpty() {
			break
		}
		time.Sleep(5 * time.Second)
		fmt.Println("Waiting for node account to be registered...")
	}
	for _, item := range na.SignerMembership {
		pubkeyMgr.AddPubKey(item, true)
	}
	if na.PubKeySet.Secp256k1.IsEmpty() {
		return nil, fmt.Errorf("unable to find pubkey for this node account. Exiting...")
	}
	pubkeyMgr.AddPubKey(na.PubKeySet.Secp256k1, true)

	// Create pubkey manager and add our private key (Yggdrasil pubkey)
	thorchainBlockScanner, err := NewThorchainBlockScan(cfg.BlockScanner, thorchainScanStorage, thorchainBridge, m, pubkeyMgr)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create thorchain block scan")
	}

	kg, err := tss.NewTssKeyGen(tssCfg, thorKeys)
	if nil != err {
		return nil, fmt.Errorf("fail to create Tss Key gen,err:%w", err)
	}

	return &Signer{
		logger:                log.With().Str("module", "signer").Logger(),
		cfg:                   cfg,
		wg:                    &sync.WaitGroup{},
		stopChan:              make(chan struct{}),
		thorchainBlockScanner: thorchainBlockScanner,
		chains:                chains,
		m:                     m,
		storage:               thorchainScanStorage,
		errCounter:            m.GetCounterVec(metrics.SignerError),
		pubkeyMgr:             pubkeyMgr,
		thorchainBridge:       thorchainBridge,
		tssKeygen:             kg,

	}, nil
}

func (s *Signer) getChain(chainName common.Chain) (chainclients.ChainClient, error) {
	chain := s.chains[chainName]
	if chain == nil {
		s.logger.Debug().Str("chain", chainName.String()).Msg("is not supported yet")
		return nil, chainclients.NotSupported
	}
	return chain, nil
}

func (s *Signer) Start() error {
	s.wg.Add(1)
	go s.processTxnOut(s.thorchainBlockScanner.GetTxOutMessages(), 1)
	if err := s.retryAll(); err != nil {
		return errors.Wrap(err, "fail to retry txouts")
	}
	s.wg.Add(1)
	go s.retryFailedTxOutProcessor()

	s.wg.Add(1)
	go s.processKeygen(s.thorchainBlockScanner.GetKeygenMessages())

	return s.thorchainBlockScanner.Start()
}

func (s *Signer) retryAll() error {
	txOuts, err := s.storage.GetTxOutsForRetry(false)
	if err != nil {
		s.errCounter.WithLabelValues("fail_get_txout_for_retry", "").Inc()
		return errors.Wrap(err, "fail to get txout for retry")
	}
	s.logger.Info().Msgf("THORNode find (%d) txOut need to be retry, retrying now", len(txOuts))
	if err := s.retryTxOut(txOuts); err != nil {
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
			_, err := s.getChain(item.Chain)
			if err != nil {
				s.logger.Error().Err(err).Msgf("not supported %s", item.Chain.String())
				continue
			}

			if err := s.signAndBroadcast(item); err != nil {
				s.errCounter.WithLabelValues("fail_sign_and_broadcast", strconv.FormatInt(item.Height, 10)).Inc()
				s.logger.Error().Err(err).Str("height", strconv.FormatInt(item.Height, 10)).Msg("fail to sign and broadcast")
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

func (s *Signer) shouldSign(tx types.TxArrayItem) bool {
	return s.pubkeyMgr.HasPubKey(tx.VaultPubKey)
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
			if err != nil {
				s.errCounter.WithLabelValues("fail_get_txout_for_retry", "").Inc()
				s.logger.Error().Err(err).Msg("fail to get txout for retry")
				continue
			}
			if err := s.retryTxOut(txOuts); err != nil {
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

			strHeight := strconv.FormatInt(txOut.Height, 10)
			s.logger.Info().Msgf("Received a TxOut Array of %v from the Thorchain", txOut)
			if err := s.storage.SetTxOutStatus(txOut, Processing); err != nil {
				s.errCounter.WithLabelValues("fail_update_txout_local", strHeight).Inc()
				s.logger.Error().Err(err).Msg("fail to update txout local storage")
				// raise alert
				return
			}

			if err := s.signAndBroadcast(txOut); err != nil {
				s.errCounter.WithLabelValues("fail_sign_and_broadcast", strHeight).Inc()
				s.logger.Error().Err(err).Msg("fail to sign txout and broadcast, will retry later")
				if err := s.storage.SetTxOutStatus(txOut, Failed); err != nil {
					s.errCounter.WithLabelValues("fail_update_txout_local", strHeight).Inc()
					s.logger.Error().Err(err).Msg("fail to update txout local storage")
					// raise alert
					return
				}
			}
			if err := s.storage.RemoveTxOut(txOut); err != nil {
				s.errCounter.WithLabelValues("fail_remove_txout_local", strHeight).Inc()
				s.logger.Error().Err(err).Msg("fail to remove txout from local store")
			}
		}
	}
}

func (s *Signer) processKeygen(ch <-chan ttypes.KeygenBlock) {
	s.logger.Info().Msg("start to process keygen")
	defer s.logger.Info().Msg("stop to process keygen")
	defer s.wg.Done()
	for {
		select {
		case <-s.stopChan:
			return
		case keygenBlock, more := <-ch:
			if !more {
				return
			}
			s.logger.Info().Msgf("Received a keygen block %+v from the Thorchain", keygenBlock)
			for _, keygenReq := range keygenBlock.Keygens {
				// Add pubkeys to pubkey manager for monitoring...
				// each member might become a yggdrasil pool
				for _, pk := range keygenReq.Members {
					s.pubkeyMgr.AddPubKey(pk, false)
				}

				pubKey, blame, err := s.tssKeygen.GenerateNewKey(keygenReq.Members)
				if !blame.IsEmpty() {
					err := fmt.Errorf("reason: %s, nodes %+v", blame.FailReason, blame.BlameNodes)
					s.logger.Error().Err(err).Msg("Blame")
				}

				if err != nil {
					s.errCounter.WithLabelValues("fail_to_keygen_pubkey", "").Inc()
					s.logger.Error().Err(err).Msg("fail to generate new pubkey")
				}
				if !pubKey.Secp256k1.IsEmpty() {
					s.pubkeyMgr.AddPubKey(pubKey.Secp256k1, true)
				}

				if err := s.sendKeygenToThorchain(keygenBlock.Height, pubKey.Secp256k1, blame, keygenReq.Members, keygenReq.Type); err != nil {
					s.errCounter.WithLabelValues("fail_to_broadcast_keygen", "").Inc()
					s.logger.Error().Err(err).Msg("fail to broadcast keygen")
				}
			}
		}
	}
}

func (s *Signer) sendKeygenToThorchain(height int64, poolPk common.PubKey, blame common.Blame, input common.PubKeys, keygenType ttypes.KeygenType) error {
	stdTx, err := s.thorchainBridge.GetKeygenStdTx(poolPk, blame, input, keygenType, height)
	strHeight := strconv.FormatInt(height, 10)
	if err != nil {
		s.errCounter.WithLabelValues("fail_to_sign", strHeight).Inc()
		return errors.Wrap(err, "fail to sign the tx")
	}
	txID, err := s.thorchainBridge.Broadcast(*stdTx, types.TxSync)
	if err != nil {
		s.errCounter.WithLabelValues("fail_to_send_to_thorchain", strHeight).Inc()
		return errors.Wrap(err, "fail to send the tx to thorchain")
	}
	s.logger.Info().Int64("block", height).Str("thorchain hash", txID.String()).Msg("sign and send to thorchain successfully")
	return nil
}

// signAndBroadcast retry a few times before THORNode move on to he next block
func (s *Signer) signAndBroadcast(txOut types.TxOut) error {
	// most case , there should be only one item in txOut.TxArray, but sometimes there might be more than one
	height := txOut.Height
	for _, item := range txOut.TxArray {
		key := item.GetKey(height)
		processed, err := s.storage.HasTxOutItem(key)
		if err != nil {
			s.logger.Error().Err(err).Msg("fail to check against local level db")
			continue
		}
		if processed {
			s.logger.Debug().Msgf("%+v processed already", item)
			continue
		}

		chain, err := s.getChain(item.Chain)
		if err != nil {
			s.logger.Error().Err(err).Msgf("not supported %s", item.Chain.String())
			continue
		}

		if !s.shouldSign(item) {
			s.logger.Info().Str("signer_address", chain.GetAddress(item.VaultPubKey)).Msg("different pool address, ignore")
			if err := s.storage.ClearTxOutItem(key); err != nil {
				s.logger.Error().Err(err).Msg("fail to mark it off from local db")
			}
			continue
		}

		if len(item.ToAddress) == 0 {
			s.logger.Info().Msg("To address is empty, THORNode don't know where to send the fund , ignore")
			if err := s.storage.ClearTxOutItem(key); err != nil {
				s.logger.Error().Err(err).Msg("fail to mark it off from local db")
			}
			continue
		}

		// Check if we're sending all funds back (memo "yggdrasil-")
		// In this scenario, we should chose the coins to send ourselves
		tx := item.TxOutItem()
		if strings.EqualFold(tx.Memo, thorchain.YggdrasilReturnMemo{}.GetType().String()) && item.Coin.IsEmpty() {
			tx, err = s.handleYggReturn(tx)
			if err != nil {
				s.logger.Error().Err(err).Msg("failed to handle yggdrasil return")
				if err := s.storage.ClearTxOutItem(key); err != nil {
					s.logger.Error().Err(err).Msg("fail to mark it off from local db")
				}
				continue
			}
		}

		start := time.Now()
		defer func() {
			s.m.GetHistograms(metrics.SignAndBroadcastDuration(chain.GetChain().String())).Observe(time.Since(start).Seconds())
		}()

		if !tx.OutHash.IsEmpty() {
			s.logger.Info().Str("OutHash", tx.OutHash.String()).Msg("tx had been sent out before")
			return nil
		}

		signedTx, err := chain.SignTx(tx, height)
		if err != nil {
			s.logger.Error().Err(err).Msg("fail to sign tx")
			continue
		}

		err = chain.BroadcastTx(signedTx)
		if err != nil {
			// since we failed the txn, we'll clear the local db of this record
			// for retry later
			if err := s.storage.ClearTxOutItem(key); err != nil {
				s.logger.Error().Err(err).Msg("fail to mark it off from local db")
			}
			s.logger.Error().Err(err).Msg("fail to broadcast tx to chain")
			continue
		}
		if err := s.storage.SuccessTxOutItem(key); err != nil {
			s.logger.Error().Err(err).Msg("fail to mark it off from local db")
			continue
		}
	}

	return nil
}

func (s *Signer) handleYggReturn(tx types.TxOutItem) (types.TxOutItem, error) {
	chain, err := s.getChain(tx.Chain)
	if err != nil {
		s.logger.Error().Err(err).Msgf("not supported %s", tx.Chain.String())
		return tx, err
	}
	addr, err := stypes.AccAddressFromHex(chain.GetAddress(tx.VaultPubKey))
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to convert to AccAddress")
		return tx, err
	}

	acct, err := chain.GetAccount(addr)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get chain account info")
		return tx, err
	}
	tx.Coins = make(common.Coins, 0)
	gas := chain.GetGasFee(uint64(len(acct.Coins)))
	for _, coin := range acct.Coins {
		asset, err := common.NewAsset(coin.Denom)
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to parse asset")
			return tx, err
		}
		amount := sdk.NewUint(uint64(coin.Amount))
		if asset.Chain == tx.Chain {
			amount = common.SafeSub(amount, gas[0].Amount)
		}
		tx.Coins = append(tx.Coins, common.NewCoin(asset, amount))
	}

	return tx, nil
}

// Stop the signer process
func (s *Signer) Stop() error {
	s.logger.Info().Msg("receive request to stop signer")
	defer s.logger.Info().Msg("signer stopped successfully")
	close(s.stopChan)
	s.wg.Wait()
	if err := s.m.Stop(); err != nil {
		s.logger.Error().Err(err).Msg("fail to stop metric server")
	}
	return s.storage.Close()
}
