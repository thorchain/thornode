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
	Chain                 chainclients.ChainClient
	storage               *ThorchainBlockScannerStorage
	m                     *metrics.Metrics
	errCounter            *prometheus.CounterVec
	tssKeygen             *tss.KeyGen
	thorKeys              *thorclient.Keys
	pubkeyMgr             pubkeymanager.PubKeyValidator
}

// NewSigner create a new instance of signer
func NewSigner(cfg config.SignerConfiguration, thorchainBridge *thorclient.ThorchainBridge, thorKeys *thorclient.Keys, pubkeyMgr pubkeymanager.PubKeyValidator, tssCfg config.TSSConfiguration, chain chainclients.ChainClient, m *metrics.Metrics) (*Signer, error) {
	thorchainScanStorage, err := NewThorchainBlockScannerStorage(cfg.SignerDbPath)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create thorchain scan storage")
	}

	var na *ttypes.NodeAccount
	for i := 0; i < 300; i++ { // wait for 5 min before timing out
		var err error
		na, err = thorchainBridge.GetNodeAccount(thorKeys.GetSignerInfo().GetAddress().String())
		if nil != err {
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
	if nil != err {
		return nil, errors.Wrap(err, "fail to create thorchain block scan")
	}

	signer := &Signer{
		logger:                log.With().Str("module", "signer").Logger(),
		cfg:                   cfg,
		wg:                    &sync.WaitGroup{},
		stopChan:              make(chan struct{}),
		thorchainBlockScanner: thorchainBlockScanner,
		Chain:                 chain,
		m:                     m,
		storage:               thorchainScanStorage,
		errCounter:            m.GetCounterVec(metrics.SignerError),
		pubkeyMgr:             pubkeyMgr,
		thorchainBridge:       thorchainBridge,
	}

	kg, err := tss.NewTssKeyGen(tssCfg, thorKeys)
	if nil != err {
		return nil, fmt.Errorf("fail to create Tss Key gen,err:%w", err)
	}
	signer.tssKeygen = kg
	return signer, nil
}

func (s *Signer) Start() error {
	s.wg.Add(1)
	go s.processTxnOut(s.thorchainBlockScanner.GetTxOutMessages(), 1)
	if err := s.retryAll(); nil != err {
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
					Msg("not chain , THORNode don't sign it")
				continue
			}

			if err := s.signTxOutAndSendToChain(item); nil != err {
				s.errCounter.WithLabelValues("fail_sign_send_to_chain", strconv.FormatInt(item.Height, 10)).Inc()
				s.logger.Error().Err(err).Str("height", strconv.FormatInt(item.Height, 10)).Msg("fail to sign and send it to chain")
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

func (s *Signer) shouldSign(tai types.TxArrayItem) bool {
	return s.pubkeyMgr.HasPubKey(tai.VaultPubKey)
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
			strHeight := strconv.FormatInt(txOut.Height, 10)
			if !more {
				return
			}
			s.logger.Info().Msgf("Received a TxOut Array of %v from the Thorchain", txOut)
			if err := s.storage.SetTxOutStatus(txOut, Processing); nil != err {
				s.errCounter.WithLabelValues("fail_update_txout_local", strHeight).Inc()
				s.logger.Error().Err(err).Msg("fail to update txout local storage")
				// raise alert
				return
			}

			if err := s.signTxOutAndSendToChain(txOut); nil != err {
				s.errCounter.WithLabelValues("fail_sign_send_to_chain", strHeight).Inc()
				s.logger.Error().Err(err).Msg("fail to send txout to chain, will retry later")
				if err := s.storage.SetTxOutStatus(txOut, Failed); nil != err {
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
	if nil != err {
		s.errCounter.WithLabelValues("fail_to_sign", strHeight).Inc()
		return errors.Wrap(err, "fail to sign the tx")
	}
	txID, err := s.thorchainBridge.Broadcast(*stdTx, types.TxSync)
	if nil != err {
		s.errCounter.WithLabelValues("fail_to_send_to_thorchain", strHeight).Inc()
		return errors.Wrap(err, "fail to send the tx to thorchain")
	}
	s.logger.Info().Int64("block", height).Str("thorchain hash", txID.String()).Msg("sign and send to thorchain successfully")
	return nil
}

// signAndSendToChainWithRetry retry a few times before THORNode move on to he next block
func (s *Signer) signTxOutAndSendToChain(txOut types.TxOut) error {
	// most case , there should be only one item in txOut.TxArray, but sometimes there might be more than one
	height := txOut.Height
	for _, item := range txOut.TxArray {
		processed, err := s.storage.HasTxOutItem(item, height)
		if nil != err {
			return fmt.Errorf("fail to check against local level db: %w", err)
		}
		if processed {
			s.logger.Debug().Msgf("%+v processed already", item)
			continue
		}

		if !s.shouldSign(item) {
			s.logger.Info().
				Str("signer_address", s.Chain.GetAddress(item.VaultPubKey)).
				Msg("different pool address, ignore")
			continue
		}

		if len(item.ToAddress) == 0 {
			s.logger.Info().Msg("To address is empty, THORNode don't know where to send the fund , ignore")
			continue
		}

		// Check if we're sending all funds back (memo "yggdrasil-")
		// In this scenario, we should chose the coins to send ourselves
		out := item.TxOutItem()
		if strings.EqualFold(out.Memo, thorchain.YggdrasilReturnMemo{}.GetType().String()) && item.Coin.IsEmpty() {
			out, err = s.handleYggReturn(out)
			if err != nil {
				continue
			}
		}

		err = s.signAndSendToChain(out, height)
		if nil != err {
			return fmt.Errorf("fail to broadcast tx to chain: %w", err)
		}
		if err := s.storage.SetTxOutItem(item, height); nil != err {
			return fmt.Errorf("fail to mark it off from local db: %w", err)
		}
	}

	return nil
}

func (s *Signer) handleYggReturn(out types.TxOutItem) (types.TxOutItem, error) {
	addr, err := stypes.AccAddressFromHex(s.Chain.GetAddress(out.VaultPubKey))
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to convert to AccAddress")
		return out, err
	}

	acct, err := s.Chain.GetAccount(addr)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get chain account info")
		return out, err
	}
	out.Coins = make(common.Coins, 0)
	gas := common.GetBNBGasFee(uint64(len(acct.Coins)))
	for _, coin := range acct.Coins {
		asset, err := common.NewAsset(coin.Denom)
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to parse asset")
			return out, err
		}
		amount := sdk.NewUint(uint64(coin.Amount))
		if asset.IsBNB() {
			amount = common.SafeSub(amount, gas[0].Amount)
		}
		out.Coins = append(out.Coins, common.NewCoin(asset, amount))
	}

	return out, nil
}

func (s *Signer) signAndSendToChain(tai types.TxOutItem, height int64) error {
	start := time.Now()
	defer func() {
		s.m.GetHistograms(metrics.SignAndBroadcastToChainDuration(s.Chain.GetChain())).Observe(time.Since(start).Seconds())
	}()
	if !tai.OutHash.IsEmpty() {
		s.logger.Info().Str("OutHash", tai.OutHash.String()).Msg("tx had been sent out before")
		return nil
	}
	if err := s.Chain.SignAndBroadcastToChain(tai, height); nil != err {
		s.logger.Error().Err(err).Msg("fail to broadcast a tx to chain")
		return err
	}

	s.logger.Debug().
		Msg("signed and send to chain successfully")
	s.m.GetCounter(metrics.TxToChainSignedBroadcast(s.Chain.GetChain())).Inc()

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
