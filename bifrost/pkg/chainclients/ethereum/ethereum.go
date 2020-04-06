package ethereum

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/ethclient"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	etypes "github.com/ethereum/go-ethereum/core/types"


	pkerrors "github.com/pkg/errors"	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tssp "gitlab.com/thorchain/tss/go-tss/tss"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
)

// Ethereum is a structure to sign and broadcast tx to Ethereum chain used by signer mostly
type Ethereum struct {
	logger             zerolog.Logger
	RPCHost            string
	cfg                config.ChainConfiguration
	chainID            int64
	isTestNet          bool
	client             *ethclient.Client
	accts              *EthereumMetaDataStore
	currentBlockHeight int64
	signLock           *sync.Mutex
	keyManager         *KeyManager
	thorchainBridge    *thorclient.ThorchainBridge
	blockScanner       *EthereumBlockScanner
}

// NewEthereum create new instance of Ethereum client
func NewEthereum(thorKeys *thorclient.Keys, cfg config.ChainConfiguration, server *tssp.TssServer, thorchainBridge *thorclient.ThorchainBridge) (*Ethereum, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("rpc host is empty")
	}
	rpcHost := cfg.RPCHost

	priv, err := thorKeys.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("fail to get private key: %w", err)
	}

	pk, err := common.NewPubKeyFromCrypto(priv.PubKey())
	if err != nil {
		return nil, fmt.Errorf("fail to get pub key: %w", err)
	}
	if thorchainBridge == nil {
		return nil, errors.New("thorchain bridge is nil")
	}
	localKm := &KeyManager{
		privKey: priv,
		addr:    ecommon.HexToAddress(strings.ToLower(priv.PubKey().Address().String())),
		pubkey:  pk,
		server:  server,
		logger:  log.With().Str("module", "tss_signer").Logger(),
	}

	if !strings.HasPrefix(rpcHost, "http") {
		rpcHost = fmt.Sprintf("http://%s", rpcHost)
	}

	ctx := context.Background()
	ethClient, err := ethclient.DialContext(ctx, cfg.ChainHost)
	if err != nil {
		return nil, err
	}

	return &Ethereum{
		logger:          log.With().Str("module", "Ethereum").Logger(),
		RPCHost:         rpcHost,
		cfg:             cfg,
		accts:           NewEthereumMetaDataStore(),
		client:          ethClient,
		signLock:        &sync.Mutex{},
		keyManager:      localKm,
		thorchainBridge: thorchainBridge,
	}, nil
}

func (e *Ethereum) initBlockScanner(pubkeyMgr pubkeymanager.PubKeyValidator, m *metrics.Metrics) error {
	e.checkIsTestNet()

	var err error
	startBlockHeight := int64(0)
	if !e.cfg.BlockScanner.EnforceBlockHeight {
		startBlockHeight, err = e.thorchainBridge.GetLastObservedInHeight(common.ETHChain)
		if err != nil {
			return pkerrors.Wrap(err, "fail to get start block height from thorchain")
		}
		if startBlockHeight == 0 {
			startBlockHeight, err = e.GetHeight()
			if err != nil {
				return pkerrors.Wrap(err, "fail to get Ethereum height")
			}
			e.logger.Info().Int64("height", startBlockHeight).Msg("Current block height is indeterminate; using current height from Ethereum.")
		}
	} else {
		startBlockHeight = e.cfg.BlockScanner.StartBlockHeight
	}
	e.blockScanner, err = NewEthereumBlockScanner(e.cfg.BlockScanner, startBlockHeight, e.isTestNet, e.client, pubkeyMgr, m)
	if err != nil {
		return pkerrors.Wrap(err, "fail to create block scanner")
	}
	return nil
}

func (e *Ethereum) Start(globalTxsQueue chan stypes.TxIn, pubkeyMgr pubkeymanager.PubKeyValidator, m *metrics.Metrics) error {
	err := e.initBlockScanner(pubkeyMgr, m)
	if err != nil {
		e.logger.Error().Err(err).Msg("fail to init block scanner")
		return err
	}
	e.blockScanner.Start(globalTxsQueue)
	return nil
}

func (e *Ethereum) Stop() error {
	return e.blockScanner.Stop()
}

// IsTestNet determinate whether we are running on test net by checking the status
func (e *Ethereum) checkIsTestNet() {
	// Cached data after first call
	if e.isTestNet {
		return
	}
	ctx := context.Background()
	chainID, err := e.client.ChainID(ctx)
	if err != nil {
		log.Fatal().Msgf("Unable to get chain id")
		return
	}

	e.chainID = chainID.Int64()
	e.isTestNet = e.chainID > 1
}

func (e *Ethereum) GetChain() common.Chain {
	return common.ETHChain
}

func (e *Ethereum) GetHeight() (int64, error) {
	ctx := context.Background()
	block, err := e.client.BlockByNumber(ctx, nil)
	if err != nil {
		return -1, nil
	}
	return block.Number().Int64(), nil
}

// GetAddress return current signer address, it will be bech32 encoded address
func (e *Ethereum) GetAddress(poolPubKey common.PubKey) string {
	addr, err := poolPubKey.GetAddress(common.ETHChain)
	if err != nil {
		e.logger.Error().Err(err).Str("pool_pub_key", poolPubKey.String()).Msg("fail to get pool address")
		return ""
	}
	return addr.String()
}

func (e *Ethereum) getGasPrice() (*big.Int, error) {
	ctx := context.Background()
	return e.client.SuggestGasPrice(ctx)
}

func (e *Ethereum) GetGasFee(count uint64) common.Gas {
	return common.GetETHGasFee()
}

func (e *Ethereum) ValidateMetadata(inter interface{}) bool {
	meta := inter.(EthereumMetadata)
	acct := e.accts.GetByAccount(meta.Address)
	return acct.Address == meta.Address && acct.Nonce == meta.Nonce
}

// SignTx sign the the given TxArrayItem
func (e *Ethereum) SignTx(tx stypes.TxOutItem, height int64) ([]byte, error) {
	e.signLock.Lock()
	defer e.signLock.Unlock()

	toAddr := tx.ToAddress.String()

	value := uint64(0)
	for _, coin := range tx.Coins {
		value += coin.Amount.Uint64()
	}
	if len(toAddr) == 0 || value == 0 {
		e.logger.Error().Msg("invalid tx params")
		return nil, nil
	}
	fromAddr := e.GetAddress(tx.VaultPubKey)

	currentHeight, err := e.GetHeight()
	if err != nil {
		e.logger.Error().Err(err).Msg("fail to get current Ethereum block height")
		return nil, err
	}

	meta := e.accts.Get(tx.VaultPubKey)
	if currentHeight > meta.BlockHeight {
		ctx := context.Background()
		nonce, err := e.client.NonceAt(ctx, ecommon.HexToAddress(fromAddr), big.NewInt(currentHeight))
		if err != nil {
			return nil, fmt.Errorf("fail to get account nonce: %w", err)
		}
		atomic.StoreInt64(&e.currentBlockHeight, currentHeight)
		e.accts.Set(tx.VaultPubKey, EthereumMetadata{
			Address:     fromAddr,
			Nonce:       nonce,
			BlockHeight: currentHeight,
		})
	}

	meta = e.accts.Get(tx.VaultPubKey)
	e.logger.Info().Uint64("nonce", meta.Nonce).Msg("account info")

	gasPrice, err := e.getGasPrice()
	if err != nil {
		return nil, errors.New("failed to get current gas price")
	}
	createdTx := etypes.NewTransaction(meta.Nonce, ecommon.HexToAddress(toAddr), big.NewInt(int64(value)), uint64(21000), gasPrice, []byte("ETH.ETH"))

	// why it was just height there? it should be currentHeight
	rawBz, err := e.signMsg(createdTx, fromAddr, tx.VaultPubKey, currentHeight, tx)
	if err != nil {
		return nil, fmt.Errorf("fail to sign message: %w", err)
	}

	if len(rawBz) == 0 {
		// the transaction was already signed
		return nil, nil
	}

	// increment sequence number
	e.accts.NonceInc(tx.VaultPubKey)
	return rawBz, nil
}

// local sign
func (e *Ethereum) sign(tx *etypes.Transaction, poolPubKey common.PubKey, signerPubKeys common.PubKeys) ([]byte, error) {
	if e.keyManager.Pubkey().Equals(poolPubKey) {
		return e.keyManager.Sign(tx)
	}
	return e.keyManager.SignWithPool(tx, poolPubKey, signerPubKeys)
}

// signWithRetry is design to sign a given message until it success or the same message had been send out by other signer
func (e *Ethereum) signMsg(tx *etypes.Transaction, from string, poolPubKey common.PubKey, height int64, txOutItem stypes.TxOutItem) ([]byte, error) {
	keySignParty, err := e.thorchainBridge.GetKeysignParty(poolPubKey)
	if err != nil {
		e.logger.Error().Err(err).Msg("fail to get keysign party")
		return nil, err
	}
	rawBytes, err := e.sign(tx, poolPubKey, keySignParty)
	if err == nil && rawBytes != nil {
		return rawBytes, nil
	}
	var keysignError tss.KeysignError
	if errors.As(err, &keysignError) {
		if len(keysignError.Blame.BlameNodes) == 0 {
			// TSS doesn't know which node to blame
			return nil, err
		}

		// key sign error forward the keysign blame to thorchain
		txID, err := e.thorchainBridge.PostKeysignFailure(keysignError.Blame, height, txOutItem.Memo, txOutItem.Coins)
		if err != nil {
			e.logger.Error().Err(err).Msg("fail to post keysign failure to thorchain")
			return nil, err
		} else {
			e.logger.Info().Str("tx_id", txID.String()).Msgf("post keysign failure to thorchain")
			return nil, fmt.Errorf("sent keysign failure to thorchain")
		}
	}
	e.logger.Error().Err(err).Msg("fail to sign tx")
	// should THORNode give up? let's check the seq no on binance chain
	// keep in mind, when THORNode don't run our own binance full node, THORNode might get rate limited by binance
	return nil, err
}

// GettAccount gets account by address in eth client
func (e *Ethereum) GetAccount(addr string) (common.Account, error) {
	ctx := context.Background()
	nonce, err := e.client.NonceAt(ctx, ecommon.HexToAddress(addr), nil)
	if err != nil {
		return common.Account{}, fmt.Errorf("fail to get account nonce: %w", err)
	}
	balance, err := e.client.BalanceAt(ctx, ecommon.HexToAddress(addr), nil)
	if err != nil {	
		return common.Account{}, fmt.Errorf("fail to get account nonce: %w", err)
	}
	account := common.NewAccount(int64(nonce), 0, common.AccountCoins{common.AccountCoin{Amount: balance.Uint64(), Denom: "ETH.ETH"}})
	return account, nil
}

// BroadcastTx decodes tx using rlp and broadcasts too Ethereum chain
func (e *Ethereum) BroadcastTx(stx stypes.TxOutItem, hexTx []byte) error {
	var tx *etypes.Transaction = &etypes.Transaction{}
	err := rlp.DecodeBytes(hexTx, tx)
	if err != nil {
		return err
	}
	e.accts.NonceInc(stx.VaultPubKey)
	ctx := context.Background()
	return e.client.SendTransaction(ctx, tx)
}
