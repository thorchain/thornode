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
	"github.com/tendermint/tendermint/crypto"

	"gitlab.com/thorchain/thornode/bifrost/binance"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain"
	ttypes "gitlab.com/thorchain/thornode/x/thorchain/types"
)

// Binance interface
type Binance interface {
	BroadcastTx(hexTx []byte) error
	GetAccount(addr stypes.AccAddress) (stypes.BaseAccount, error)
	GetAddress(poolPubKey common.PubKey) string
	GetPubKey() crypto.PubKey
	SignTx(tai types.TxOutItem, height int64) ([]byte, map[string]string, error)
}

// Signer will pull the tx out from thorchain and then forward it to binance chain
type Signer struct {
	logger                zerolog.Logger
	cfg                   config.SignerConfiguration
	wg                    *sync.WaitGroup
	thorchainBridge       *thorclient.ThorchainBridge
	stopChan              chan struct{}
	thorchainBlockScanner *ThorchainBlockScan
	Binance               Binance
	storage               *ThorchainBlockScannerStorage
	m                     *metrics.Metrics
	errCounter            *prometheus.CounterVec
	tssKeygen             *tss.KeyGen
	thorKeys              *thorclient.Keys
	pkm                   *PubKeyManager
}

// NewSigner create a new instance of signer
func NewSigner(cfg config.SignerConfiguration, thorchainBridge *thorclient.ThorchainBridge, thorKeys *thorclient.Keys, useTSS bool, tssCfg config.TSSConfiguration, bnb *binance.Binance, m *metrics.Metrics) (*Signer, error) {
	thorchainScanStorage, err := NewThorchainBlockScannerStorage(cfg.SignerDbPath)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create thorchain scan storage")
	}

	pkm := NewPubKeyManager()

	var na ttypes.NodeAccount
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
		pkm.Add(item)
	}
	if na.PubKeySet.Secp256k1.IsEmpty() {
		return nil, fmt.Errorf("Unable to find pubkey for this node account.Exiting...")
	}
	pkm.Add(na.PubKeySet.Secp256k1)

	// Create pubkey manager and add our private key (Yggdrasil pubkey)
	thorchainBlockScanner, err := NewThorchainBlockScan(cfg.BlockScanner, thorchainScanStorage, thorchainBridge, m, pkm)
	if nil != err {
		return nil, errors.Wrap(err, "fail to create thorchain block scan")
	}

	signer := &Signer{
		logger:                log.With().Str("module", "signer").Logger(),
		cfg:                   cfg,
		wg:                    &sync.WaitGroup{},
		stopChan:              make(chan struct{}),
		thorchainBlockScanner: thorchainBlockScanner,
		Binance:               bnb,
		m:                     m,
		storage:               thorchainScanStorage,
		errCounter:            m.GetCounterVec(metrics.SignerError),
		pkm:                   pkm,
		thorchainBridge:       thorchainBridge,
	}

	if useTSS {
		kg, err := tss.NewTssKeyGen(tssCfg, thorKeys)
		if nil != err {
			return nil, fmt.Errorf("fail to create Tss Key gen,err:%w", err)
		}
		signer.tssKeygen = kg
	}
	return signer, nil
}

func (s *Signer) Start() error {
	if err := s.m.Start(); nil != err {
		return errors.Wrap(err, "fail to start metric collector")
	}
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
			s.logger.Info().Msgf("Received a TxOut Array of %v from the Thorchain", txOut)
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

func (s *Signer) processKeygen(ch <-chan types.Keygens, idx int) {
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

				s.pkm.Add(pubKey.Secp256k1)

				if err := s.sendKeygenToThorchain(keygens.Height, pubKey.Secp256k1, keygen); err != nil {
					s.errCounter.WithLabelValues("fail_to_broadcast_keygen", "").Inc()
					s.logger.Error().Err(err).Msg("fail to broadcast keygen")
				}
			}
		}

	}
}

func (s *Signer) sendKeygenToThorchain(height string, poolPk common.PubKey, input common.PubKeys) error {
	stdTx, err := s.thorchainBridge.GetKeygenStdTx(poolPk, input)
	if nil != err {
		s.errCounter.WithLabelValues("fail_to_sign", height).Inc()
		return errors.Wrap(err, "fail to sign the tx")
	}
	txID, err := s.thorchainBridge.Send(*stdTx, types.TxSync)
	if nil != err {
		s.errCounter.WithLabelValues("fail_to_send_to_thorchain", height).Inc()
		return errors.Wrap(err, "fail to send the tx to thorchain")
	}
	s.logger.Info().Str("block", height).Str("thorchain hash", txID.String()).Msg("sign and send to thorchain successfully")
	return nil
}

// signAndSendToBinanceChainWithRetry retry a few times before THORNode move on to he next block
func (s *Signer) signTxOutAndSendToBinanceChain(txOut types.TxOut) error {
	// most case , there should be only one item in txOut.TxArray, but sometimes there might be more than one
	height, err := strconv.ParseInt(txOut.Height, 10, 64)
	if nil != err {
		return errors.Wrapf(err, "fail to parse block height: %s ", txOut.Height)
	}
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
				Str("signer_address", s.Binance.GetAddress(item.VaultPubKey)).
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

		err = s.signAndSendToBinanceChain(out, height)
		if nil != err {
			return fmt.Errorf("fail to broadcast tx to binance chain: %w", err)
		}
		if err := s.storage.SetTxOutItem(item, height); nil != err {
			return fmt.Errorf("fail to mark it off from local db: %w", err)
		}
	}

	return nil
}

func (s *Signer) handleYggReturn(out types.TxOutItem) (types.TxOutItem, error) {
	addr, err := stypes.AccAddressFromHex(s.Binance.GetAddress(out.VaultPubKey))
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to convert to AccAddress")
		return out, err
	}

	acct, err := s.Binance.GetAccount(addr)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get binance account info")
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

func (s *Signer) signAndSendToBinanceChain(tai types.TxOutItem, height int64) error {
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
