package signer

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/thorchain/tss/go-tss/blame"
	tssp "gitlab.com/thorchain/tss/go-tss/tss"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients"
	"gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
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
	blockScanner          *blockscanner.BlockScanner
	thorchainBlockScanner *ThorchainBlockScan
	chains                map[common.Chain]chainclients.ChainClient
	storage               SignerStorage
	m                     *metrics.Metrics
	errCounter            *prometheus.CounterVec
	tssKeygen             *tss.KeyGen
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
		return nil, fmt.Errorf("fail to create thorchain scan storage: %w", err)
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

	cfg.BlockScanner.ChainID = common.THORChain // hard code to thorchain

	// Create pubkey manager and add our private key (Yggdrasil pubkey)
	thorchainBlockScanner, err := NewThorchainBlockScan(cfg.BlockScanner, storage, thorchainBridge, m, pubkeyMgr)
	if err != nil {
		return nil, fmt.Errorf("fail to create thorchain block scan: %w", err)
	}

	blockScanner, err := blockscanner.NewBlockScanner(cfg.BlockScanner, storage, m, thorchainBridge, thorchainBlockScanner)
	if err != nil {
		return nil, fmt.Errorf("fail to create block scanner: %w", err)
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
		blockScanner:          blockScanner,
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

func (s *Signer) Start() error {
	s.wg.Add(1)
	go s.processTxnOut(s.thorchainBlockScanner.GetTxOutMessages(), 1)

	s.wg.Add(1)
	go s.processKeygen(s.thorchainBlockScanner.GetKeygenMessages())

	s.wg.Add(1)
	go s.signTransactions()

	s.blockScanner.Start(nil)
	return nil
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
	wg := &sync.WaitGroup{}
	for _, items := range s.storage.OrderedLists() {
		wg.Add(1)
		go func(items []TxOutStoreItem) {
			defer wg.Done()
			for i, item := range items {
				select {
				case <-s.stopChan:
					return
				default:
					if item.Status == TxSpent { // don't rebroadcast spent transactions
						continue
					}

					s.logger.Info().Msgf("Signing transaction (Num: %d | Height: %d | Status: %d): %+v", i, item.Height, item.Status, item.TxOutItem)
					if err := s.signAndBroadcast(item); err != nil {
						s.logger.Error().Err(err).Msg("fail to sign and broadcast tx out store item")
						return
					}

					// We have a successful broadcast! Remove the item from our store
					item.Status = TxSpent
					if err := s.storage.Set(item); err != nil {
						s.logger.Error().Err(err).Msg("fail to update tx out store item")
					}
				}
			}
		}(items)
	}
	wg.Wait()
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

func (s *Signer) sendKeygenToThorchain(height int64, poolPk common.PubKey, blame blame.Blame, input common.PubKeys, keygenType ttypes.KeygenType) error {
	// collect supported chains in the configuration
	chains := make(common.Chains, 0)
	for name, chain := range s.chains {
		if !chain.GetConfig().OptToRetire {
			chains = append(chains, name)
		}
	}

	stdTx, err := s.thorchainBridge.GetKeygenStdTx(poolPk, blame, input, keygenType, chains, height)
	strHeight := strconv.FormatInt(height, 10)
	if err != nil {
		s.errCounter.WithLabelValues("fail_to_sign", strHeight).Inc()
		return fmt.Errorf("fail to sign the tx: %w", err)
	}
	txID, err := s.thorchainBridge.Broadcast(*stdTx, types.TxSync)
	if err != nil {
		s.errCounter.WithLabelValues("fail_to_send_to_thorchain", strHeight).Inc()
		return fmt.Errorf("fail to send the tx to thorchain: %w", err)
	}
	s.logger.Info().Int64("block", height).Str("thorchain hash", txID.String()).Msg("sign and send to thorchain successfully")
	return nil
}

// signAndBroadcast retry a few times before THORNode move on to he next block
func (s *Signer) signAndBroadcast(item TxOutStoreItem) error {
	height := item.Height
	tx := item.TxOutItem
	blockHeight, err := s.thorchainBridge.GetBlockHeight()
	if err != nil {
		s.logger.Error().Err(err).Msgf("fail to get block height")
		return err
	}
	// TODO hardcode it as 0.1.0 for now, will need to get it appropriately later
	cv := constants.GetConstantValues(semver.MustParse("0.1.0"))
	if blockHeight-height > cv.GetInt64Value(constants.SigningTransactionPeriod) {
		s.logger.Error().Msgf("tx was created at block height(%d), now it is (%d), it is older than (%d) blocks , skip it ", height, blockHeight, cv.GetInt64Value(constants.SigningTransactionPeriod))
		return nil
	}
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

	// Check if we're sending all funds back , given we don't have memo in txoutitem anymore, so it rely on the coins field to be empty
	// In this scenario, we should chose the coins to send ourselves
	if tx.Coins.IsEmpty() {
		tx, err = s.handleYggReturn(height, tx)
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

	// We get the keysign object from thorchain again to ensure it hasn't
	// been signed already, and we can skip. This helps us not get stuck on
	// a task that we'll never sign, because 2/3rds already has and will
	// never be available to sign again.
	txOut, err := s.thorchainBridge.GetKeysign(height, tx.VaultPubKey.String())
	if err != nil {
		s.logger.Error().Err(err).Msg("fail to get keysign items")
		return err
	}
	for _, out := range txOut.Chains {
		for _, txArray := range out.TxArray {
			if txArray.TxOutItem().Equals(tx) && !txArray.OutHash.IsEmpty() {
				// already been signed, we can skip it
				s.logger.Info().Str("tx_id", tx.OutHash.String()).Msgf("already signed. skipping...")
				return nil
			}
		}
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

	if err := chain.BroadcastTx(tx, signedTx); err != nil {
		s.logger.Error().Err(err).Msg("fail to broadcast tx to chain")
		return err
	}

	return nil
}

func (s *Signer) handleYggReturn(height int64, tx types.TxOutItem) (types.TxOutItem, error) {
	chain, err := s.getChain(tx.Chain)
	if err != nil {
		s.logger.Error().Err(err).Msgf("not supported %s", tx.Chain.String())
		return tx, err
	}
	isValid, _ := s.pubkeyMgr.IsValidPoolAddress(tx.ToAddress.String(), tx.Chain)
	if !isValid {
		errInvalidPool := fmt.Errorf("yggdrasil return should return to a valid pool address,%s is not valid", tx.ToAddress.String())
		s.logger.Error().Err(errInvalidPool).Msg("invalid yggdrasil return address")
		return tx, errInvalidPool
	}
	// it is important to set the memo field to `yggdrasil-` , thus chain client can use it to decide leave some gas coin behind to pay the fees
	tx.Memo = thorchain.NewYggdrasilReturn(height).String()
	acct, err := chain.GetAccount(tx.VaultPubKey)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get chain account info")
		return tx, err
	}
	tx.Coins = make(common.Coins, 0)
	for _, coin := range acct.Coins {
		asset, err := common.NewAsset(coin.Denom)
		asset.Chain = tx.Chain
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to parse asset")
			return tx, err
		}
		if coin.Amount > 0 {
			amount := sdk.NewUint(coin.Amount)
			tx.Coins = append(tx.Coins, common.NewCoin(asset, amount))
		}
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
	s.blockScanner.Stop()
	return s.storage.Close()
}
