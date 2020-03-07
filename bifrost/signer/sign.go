package signer

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tssCommon "gitlab.com/thorchain/tss/go-tss/common"
	tssp "gitlab.com/thorchain/tss/go-tss/tss"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients"
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
	storage               SignerStorage
	m                     *metrics.Metrics
	errCounter            *prometheus.CounterVec
	tssKeygen             *tss.KeyGen
	thorKeys              *thorclient.Keys
	pubkeyMgr             pubkeymanager.PubKeyValidator
}

// NewSigner create a new instance of signer
func NewSigner(cfg config.SignerConfiguration,
	thorchainBridge *thorclient.ThorchainBridge,
	thorKeys *thorclient.Keys,
	pubkeyMgr pubkeymanager.PubKeyValidator,
	tssServer *tssp.TssServer,
	tssCfg config.TSSConfiguration,
	chains map[common.Chain]chainclients.ChainClient,
	m *metrics.Metrics) (*Signer, error) {
	storage, err := NewSignerStore(cfg.SignerDbPath, thorchainBridge.GetConfig().SignerPasswd)
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
	pubkeyMgr.AddNodePubKey(na.PubKeySet.Secp256k1)

	// Create pubkey manager and add our private key (Yggdrasil pubkey)
	thorchainBlockScanner, err := NewThorchainBlockScan(cfg.BlockScanner, storage, thorchainBridge, m, pubkeyMgr)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create thorchain block scan")
	}

	kg, err := tss.NewTssKeyGen(thorKeys, tssServer)
	if err != nil {
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
		storage:               storage,
		errCounter:            m.GetCounterVec(metrics.SignerError),
		pubkeyMgr:             pubkeyMgr,
		thorchainBridge:       thorchainBridge,
		tssKeygen:             kg,
	}, nil
}

func (s *Signer) getChain(chainID common.Chain) (chainclients.ChainClient, error) {
	chain, ok := s.chains[chainID]
	if !ok {
		s.logger.Debug().Str("chain", chainID.String()).Msg("is not supported yet")
		return nil, errors.New("Not supported")
	}
	return chain, nil
}

func (s *Signer) CheckTransaction(key string, chainID common.Chain, metadata interface{}) (TxStatus, error) {
	chain, err := s.getChain(chainID)
	if err != nil {
		return TxUnknown, err
	}

	// if we don't have the transaction yet, say its unavailable
	if !s.storage.Has(key) {
		return TxUnavailable, nil
	}

	tx, err := s.storage.Get(key)
	if err != nil {
		return TxUnknown, err
	}

	// if the tx isn't available, return immediately
	if tx.Status != TxAvailable {
		return tx.Status, nil
	}

	// validate metadata
	if !chain.ValidateMetadata(metadata) {
		return TxUnavailable, nil
	}

	return TxAvailable, nil
}

func (s *Signer) Start() error {
	s.wg.Add(1)
	go s.processTxnOut(s.thorchainBlockScanner.GetTxOutMessages(), 1)

	s.wg.Add(1)
	go s.processKeygen(s.thorchainBlockScanner.GetKeygenMessages())

	s.wg.Add(1)
	go s.signTransactions()

	return s.thorchainBlockScanner.Start()
}

func (s *Signer) shouldSign(tx types.TxOutItem) bool {
	return s.pubkeyMgr.HasPubKey(tx.VaultPubKey)
}

// signTransactions - looks for work to do by getting a list of all unsigned
// transactions stored in the storage
func (s *Signer) signTransactions() {
	s.logger.Info().Msg("start to sign transactions")
	defer s.logger.Info().Msg("stop to sign transactions")
	defer s.wg.Done()
	for {
		select {
		case <-s.stopChan:
			return
		default:
			s.processTransactions()
			time.Sleep(1 * time.Second)
		}
	}
}

func (s *Signer) processTransactions() {
	for _, item := range s.storage.List() {
		select {
		case <-s.stopChan:
			return
		default:
			if item.Status == TxSpent { // don't rebroadcast spent transactions
				continue
			}

			s.logger.Info().Msgf("Signing transaction (Height: %d | Status: %d): %+v", item.Height, item.Status, item.TxOutItem)
			if err := s.signAndBroadcast(item); err != nil {
				s.logger.Error().Err(err).Msg("fail to sign and broadcast tx out store item")
				continue
			}

			// We have a successful broadcast! Remove the item from our store
			item.Status = TxSpent
			if err := s.storage.Set(item); err != nil {
				s.logger.Error().Err(err).Msg("fail to update tx out store item")
			}
		}
	}
}

// processTxnOut processes inbound TxOuts and save them to storage
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
			items := make([]TxOutStoreItem, len(txOut.TxArray))
			for i, tx := range txOut.TxArray {
				items[i] = NewTxOutStoreItem(txOut.Height, tx.TxOutItem())
			}
			if err := s.storage.Batch(items); err != nil {
				s.logger.Error().Err(err).Msg("fail to save tx out items to storage")
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

func (s *Signer) sendKeygenToThorchain(height int64, poolPk common.PubKey, blame tssCommon.Blame, input common.PubKeys, keygenType ttypes.KeygenType) error {
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
func (s *Signer) signAndBroadcast(item TxOutStoreItem) error {
	height := item.Height
	tx := item.TxOutItem
	chain, err := s.getChain(tx.Chain)
	if err != nil {
		s.logger.Error().Err(err).Msgf("not supported %s", tx.Chain.String())
		return err
	}

	if !s.shouldSign(tx) {
		s.logger.Info().Str("signer_address", chain.GetAddress(tx.VaultPubKey)).Msg("different pool address, ignore")
		return fmt.Errorf("not a member of the vault pubkey")
	}

	if len(tx.ToAddress) == 0 {
		s.logger.Info().Msg("To address is empty, THORNode don't know where to send the fund , ignore")
		return nil // return nil and discard item
	}

	// Check if we're sending all funds back (memo "yggdrasil-")
	// In this scenario, we should chose the coins to send ourselves
	if strings.EqualFold(tx.Memo, thorchain.YggdrasilReturnMemo{}.GetType().String()) && tx.Coins.IsEmpty() {
		tx, err = s.handleYggReturn(tx)
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to handle yggdrasil return")
			return err
		}
	}

	start := time.Now()
	defer func() {
		s.m.GetHistograms(metrics.SignAndBroadcastDuration(chain.GetChain())).Observe(time.Since(start).Seconds())
	}()

	if !tx.OutHash.IsEmpty() {
		s.logger.Info().Str("OutHash", tx.OutHash.String()).Msg("tx had been sent out before")
		return nil // return nil and discard item
	}
	signedTx, err := chain.SignTx(tx, height)
	if err != nil {
		s.logger.Error().Err(err).Msg("fail to sign tx")
		return err
	}

	// looks like the transaction is already signed
	if len(signedTx) == 0 {
		return nil
	}

	err = chain.BroadcastTx(signedTx)
	if err != nil {
		s.logger.Error().Err(err).Msg("fail to broadcast tx to chain")
		return err
	}

	return nil
}

func (s *Signer) handleYggReturn(tx types.TxOutItem) (types.TxOutItem, error) {
	chain, err := s.getChain(tx.Chain)
	if err != nil {
		s.logger.Error().Err(err).Msgf("not supported %s", tx.Chain.String())
		return tx, err
	}

	acct, err := chain.GetAccount(chain.GetAddress(tx.VaultPubKey))
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
		amount := sdk.NewUint(coin.Amount)
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
	if err := s.thorchainBlockScanner.Stop(); err != nil {
		s.logger.Error().Err(err).Msg("stop thorchain block scanner")
	}
	return s.storage.Close()
}
