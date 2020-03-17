package ethereum

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	chainID            string
	isTestNet          bool
	client             *http.Client
	accts              *EthereumMetaDataStore
	currentBlockHeight int64
	signLock           *sync.Mutex
	tssKeyManager      keys.KeyManager
	localKeyManager    *keyManager
	thorchainBridge    *thorclient.ThorchainBridge
	storage            *LevelDBlockScannerStorage
	signer             etypes.EIP155Signer
	blockScanner       *EthereumBlockScanner
}

// NewEthereum create new instance of Ethereum client
func NewEthereum(thorKeys *thorclient.Keys, cfg config.ChainConfiguration, server *tssp.TssServer, thorchainBridge *thorclient.ThorchainBridge) (*Ethereum, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("rpc host is empty")
	}
	rpcHost := cfg.RPCHost

	tssKm, err := tss.NewKeySign(server)
	if err != nil {
		return nil, fmt.Errorf("fail to create tss signer: %w", err)
	}

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
	localKm := &keyManager{
		privKey: priv,
		addr:    ctypes.AccAddress(priv.PubKey().Address()),
		pubkey:  pk,
	}

	if !strings.HasPrefix(rpcHost, "http") {
		rpcHost = fmt.Sprintf("http://%s", rpcHost)
	}

	ctx := context.Background()
	ethClient, err := ethclient.DialContext(ctx, cfg.ChainHost)
	if err != nil {
		return nil, err
	}

	signer := etypes.NewEIP155Signer(chainId *big.Int)

	return &Ethereum{
		logger:          log.With().Str("module", "Ethereum").Logger(),
		RPCHost:         rpcHost,
		cfg:             cfg,
		accts:           NewBinanceMetaDataStore(),
		client           ethClient,
		signLock:        &sync.Mutex{},
		tssKeyManager:   tssKm,
		localKeyManager: localKm,
		signer:          signer,
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
	e.blockScanner, err = NewEthereumBlockScanner(e.cfg.BlockScanner, startBlockHeight, e.signer, e.isTestNet, pubkeyMgr, m)
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

	chainID, err := e.client.ChainID()
	if err != nil {
		log.Fatal().Msgf("Unable to get chain id: %s\n", err.Msg)
	}

	e.chainID = chainID
	e.isTestNet = e.chainID > 1

	if e.isTestNet {
		types.Network = types.TestNetwork
	} else {
		types.Network = types.ProdNetwork
	}
}

func (e *Ethereum) GetChain() common.Chain {
	return common.ETHChain
}

func (e *Ethereum) GetHeight() (int64, error) {
	block, err := e.client.BlockByNumber(c.ctx, nil)
	return block.Number().Int64(), nil
}

func (e *Ethereum) createTx(fromAddr string, value, nonce uint64) msg.SendMsg {
	nonce := ethclient.NonceAt()
	return 
	addr, err := types.AccAddressFromBech32(fromAddr)
	if err != nil {
		e.logger.Error().Str("address", fromAddr).Err(err).Msg("fail to parse address")
	}
	fromCoins := types.Coins{}
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		fromCoins = fromCoins.Plus(t.Coins)
	}
	return e.createMsg(addr, fromCoins, transfers)
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
	ctx = context.Background()
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
	var payload []msg.Transfer

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
	if currentHeight > e.currentBlockHeight {
		ctx := context.Background()
		nonce, err := e.client.NonceAt(ctx, common.HexToAddress(fromAddr), big.NewInt(currentHeight))
		if err != nil {
			return nil, fmt.Errorf("fail to get account nonce: %w", err)
		}
		atomic.StoreInt64(&e.currentBlockHeight, currentHeight)
		e.accts.Set(tx.VaultPubKey, EthereumMetadata{
			Address: fromAddr,
			Nonce:   nonce,
		})
	}

	meta := e.accts.Get(tx.VaultPubKey)
	e.logger.Info().Int64("nonce", meta.Nonoce).Int64("nonce", meta.Nonce).Msg("account info")

	gasPrice, err := e.getGasPrice()
	if err != nil {
		return, errors.New("failed to get current gas price")
	}
	createdTx := etypes.NewTransaction(meta.Nonce, common.HexToAddress(toAddr), big.NewInt(value), uint64(21000), gasPrice, []byte("ETH.ETH"))

	// why it was just height there? it should be currentHeight
	rawBz, err := e.signWithRetry(createdTx, fromAddr, tx.VaultPubKey, currentHeight, tx)
	if err != nil {
		return nil, fmt.Errorf("fail to sign message: %w", err)
	}

	if len(rawBz) == 0 {
		// the transaction was already signed
		return nil, nil
	}

	// increment sequence number
	e.accts.NonceInc(tx.VaultPubKey)

	hexTx := []byte(hex.EncodeToString(rawBz))
	return hexTx, nil
}

// local sign
func (e *Ethereum) sign(tx etypes.Transaction, poolPubKey common.PubKey, signerPubKeys common.PubKeys) ([]byte, error) {
	if e.localKeyManager.Pubkey().Equals(poolPubKey) {
		return e.localKeyManager.Sign(tx)
	}
	k := e.tssKeyManager.(tss.ThorchainKeyManager)
	return k.SignWithPool(tx, poolPubKey, signerPubKeys)
}

// signWithRetry is design to sign a given message until it success or the same message had been send out by other signer
func (e *Ethereum) signWithRetry(tx etypes.Transaction, from string, poolPubKey common.PubKey, height int64, txOutItem stypes.TxOutItem) ([]byte, error) {
	for {
		keySignParty, err := e.thorchainBridge.GetKeysignParty(poolPubKey)
		if err != nil {
			e.logger.Error().Err(err).Msg("fail to get keysign party")
			continue
		}

		// We get the keysign object from thorchain again to ensure it hasn't
		// been signed already, and we can skip. This helps us not get stuck on
		// a task that we'll never sign, because 2/3rds already has and will
		// never be available to sign again.
		txOut, err := e.thorchainBridge.GetKeysign(height, poolPubKey.String())
		if err != nil {
			e.logger.Error().Err(err).Msg("fail to get keysign items")
			continue
		}
		for _, out := range txOut.Chains {
			for _, tx := range out.TxArray {
				item := tx.TxOutItem()
				if txOutItem.Equals(item) && !tx.OutHash.IsEmpty() {
					// already been signed, we can skip it
					e.logger.Info().Str("tx_id", tx.OutHash.String()).Msgf("already signed. skippping...")
					return nil, nil
				}
			}
		}

		rawBytes, err := e.sign(tx, poolPubKey, keySignParty)
		if err == nil && rawBytes != nil {
			return rawBytes, nil
		}
		var keysignError tss.KeysignError
		if errors.As(err, &keysignError) {
			if len(keysignError.Blame.BlameNodes) == 0 {
				// TSS doesn't know which node to blame
				continue
			}

			// key sign error forward the keysign blame to thorchain
			txID, err := e.thorchainBridge.PostKeysignFailure(keysignError.Blame, height, txOutItem.Memo, txOutItem.Coins)
			if err != nil {
				e.logger.Error().Err(err).Msg("fail to post keysign failure to thorchain")
			} else {
				e.logger.Info().Str("tx_id", txID.String()).Msgf("post keysign failure to thorchain")
			}
			continue
		}
		e.logger.Error().Err(err).Msgf("fail to sign msg with memo: %s", signMsg.Memo)
		// should THORNode give up? let's check the seq no on Ethereum chain
		// keep in mind, when THORNode don't run our own Ethereum full node, THORNode might get rate limited by Ethereum

		acc, err := e.GetAccount(from)
		if err != nil {
			e.logger.Error().Err(err).Msg("fail to get account info from Ethereum chain")
			continue
		}
		if acc.Sequence > tx.Nonce() {
			e.logger.Debug().Msgf("msg with memo: %s , seqNo: %d had been processed", string(tx.Data()), tx.Nonce())
			return nil, nil
		}
	}
}

// GettAccount gets account by address in eth client
func (e *Ethereum) GetAccount(addr string) (common.Account, error) {
	ctx := context.Background()
	nonce, err := e.client.NonceAt(ctx, common.HexToAddress(fromAddr), nil)
	if err != nil {
		return nil, fmt.Errorf("fail to get account nonce: %w", err)
	}
	balance, err := e.client.BalanceAt(ctx, common.HexToAddress(fromAddr), nil)
	if err != nil {	
		return nil, fmt.Errorf("fail to get account nonce: %w", err)
	}
	account := common.NewAccount(int64(nonce), 0, balance.Uint64())
	return account, nil
}

// BroadcastTx decodes tx using rlp and broadcasts too Ethereum chain
func (e *Ethereum) BroadcastTx(hexTx []byte) error {
	var tx etypes.Transaction
	err := rlp.DecodeBytes(hexTx, &tx)
	if err != nil {
		return err
	}
	return ethclient.SendTransaction(tx)
}
