package ethereum

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	ecommon "github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tssp "gitlab.com/thorchain/tss/go-tss/tss"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients/ethereum/types"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
)

// Client is a structure to sign and broadcast tx to Ethereum chain used by signer mostly
type Client struct {
	logger          zerolog.Logger
	cfg             config.ChainConfiguration
	chainID         types.ChainID
	pk              common.PubKey
	client          *ethclient.Client
	kw              *KeySignWrapper
	ethScanner      *BlockScanner
	accts           *EthereumMetaDataStore
	thorchainBridge *thorclient.ThorchainBridge
	blockScanner    *blockscanner.BlockScanner
}

// NewClient create new instance of Ethereum client
func NewClient(thorKeys *thorclient.Keys, cfg config.ChainConfiguration, server *tssp.TssServer, thorchainBridge *thorclient.ThorchainBridge, m *metrics.Metrics) (*Client, error) {
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

	ethPrivateKey, err := getETHPrivateKey(priv)
	if err != nil {
		return nil, err
	}

	keysignWrapper := &KeySignWrapper{
		privKey:       ethPrivateKey,
		pubKey:        pk,
		tssKeyManager: tssKm,
		logger:        log.With().Str("module", "local_signer").Str("chain", common.ETHChain.String()).Logger(),
	}
	ethClient, err := ethclient.Dial(cfg.RPCHost)
	if err != nil {
		return nil, err
	}
	c := &Client{
		logger:          log.With().Str("module", "ethereum").Logger(),
		cfg:             cfg,
		client:          ethClient,
		pk:              pk,
		accts:           NewEthereumMetaDataStore(),
		kw:              keysignWrapper,
		thorchainBridge: thorchainBridge,
	}
	c.InitChainID()

	var path string // if not set later, will in memory storage
	if len(c.cfg.BlockScanner.DBPath) > 0 {
		path = fmt.Sprintf("%s/%s", c.cfg.BlockScanner.DBPath, c.cfg.BlockScanner.ChainID)
	}
	storage, err := blockscanner.NewBlockScannerStorage(path)
	if err != nil {
		return c, fmt.Errorf("fail to create blockscanner storage: %w", err)
	}

	c.ethScanner, err = NewBlockScanner(c.cfg.BlockScanner, storage, c.chainID, c.client, m)
	if err != nil {
		return c, fmt.Errorf("fail to create eth block scanner: %w", err)
	}

	c.blockScanner, err = blockscanner.NewBlockScanner(c.cfg.BlockScanner, storage, m, c.thorchainBridge, c.ethScanner)
	if err != nil {
		return c, fmt.Errorf("fail to create block scanner: %w", err)
	}

	return c, nil
}

func (c *Client) Start(globalTxsQueue chan stypes.TxIn, globalErrataQueue chan stypes.ErrataBlock) {
	c.blockScanner.Start(globalTxsQueue)
}

func (c *Client) Stop() {
	c.blockScanner.Stop()
	c.client.Close()
}

func (c *Client) GetConfig() config.ChainConfiguration {
	return c.cfg
}

// IsTestNet determinate whether we are running on test net by checking the status
func (c *Client) InitChainID() {
	chainID, err := c.client.ChainID(context.Background())
	if err != nil {
		c.logger.Error().Err(err).Msg("Unable to get chain id")
		chainID = big.NewInt(types.Localnet)
	}
	c.chainID = types.ChainID(chainID.Int64())
	vByte = byte(int(vByte) + int(2*c.chainID))
	eipSigner = etypes.NewEIP155Signer(chainID)
}

func (c *Client) GetChain() common.Chain {
	return common.ETHChain
}

func (c *Client) GetHeight() (int64, error) {
	block, err := c.client.BlockByNumber(context.Background(), nil)
	if err != nil {
		return -1, err
	}
	return block.Number().Int64(), nil
}

// GetAddress return current signer address, it will be bech32 encoded address
func (c *Client) GetAddress(poolPubKey common.PubKey) string {
	addr, err := poolPubKey.GetAddress(common.ETHChain)
	if err != nil {
		c.logger.Error().Err(err).Str("pool_pub_key", poolPubKey.String()).Msg("fail to get pool address")
		return ""
	}
	return addr.String()
}

func (c *Client) GetGasFee(count uint64) common.Gas {
	return common.GetETHGasFee(big.NewInt(1), count)
}

func (c *Client) GetGasPrice() (*big.Int, error) {
	return c.client.SuggestGasPrice(context.Background())
}

func (c *Client) GetNonce(addr string) (uint64, error) {
	nonce, err := c.client.NonceAt(context.Background(), ecommon.HexToAddress(addr), nil)
	if err != nil {
		return 0, fmt.Errorf("fail to get account nonce: %w", err)
	}
	return nonce, nil
}

// SignTx sign the the given TxArrayItem
func (c *Client) SignTx(tx stypes.TxOutItem, height int64) ([]byte, error) {
	toAddr := tx.ToAddress.String()

	value := big.NewInt(0)
	for _, coin := range tx.Coins {
		value.Add(value, coin.Amount.BigInt())
	}
	if len(toAddr) == 0 || value.Uint64() == 0 {
		c.logger.Error().Msg("invalid tx params")
		return nil, nil
	}
	fromAddr := c.GetAddress(tx.VaultPubKey)

	currentHeight, err := c.GetHeight()
	if err != nil {
		c.logger.Error().Err(err).Msg("fail to get current Ethereum block height")
		return nil, err
	}
	meta := c.accts.Get(tx.VaultPubKey)
	if currentHeight > meta.BlockHeight {
		nonce, err := c.GetNonce(fromAddr)
		if err != nil {
			return nil, err
		}
		c.accts.Set(tx.VaultPubKey, EthereumMetadata{
			Address:     fromAddr,
			Nonce:       nonce,
			BlockHeight: currentHeight,
		})
	}
	meta = c.accts.Get(tx.VaultPubKey)
	c.logger.Info().Uint64("nonce", meta.Nonce).Msg("account info")

	gasPrice := c.ethScanner.GetGasPrice()
	encodedData := []byte(hex.EncodeToString([]byte(tx.Memo)))
	gasFee := common.GetETHGasFee(big.NewInt(1), uint64(len(tx.Memo)))[0].Amount.Uint64()
	createdTx := etypes.NewTransaction(meta.Nonce, ecommon.HexToAddress(toAddr), value, gasFee, gasPrice, encodedData)

	rawTx, err := c.sign(createdTx, fromAddr, tx.VaultPubKey, currentHeight, tx)
	if err != nil || len(rawTx) == 0 {
		return nil, fmt.Errorf("fail to sign message: %w", err)
	}
	return rawTx, nil
}

// sign is design to sign a given message with keysign party and keysign wrapper
func (c *Client) sign(tx *etypes.Transaction, from string, poolPubKey common.PubKey, height int64, txOutItem stypes.TxOutItem) ([]byte, error) {
	keySignParty, err := c.thorchainBridge.GetKeysignParty(poolPubKey)
	if err != nil {
		c.logger.Error().Err(err).Msg("fail to get keysign party")
		return nil, err
	}
	rawBytes, err := c.kw.Sign(tx, poolPubKey, keySignParty)
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
		txID, err := c.thorchainBridge.PostKeysignFailure(keysignError.Blame, height, txOutItem.Memo, txOutItem.Coins)
		if err != nil {
			c.logger.Error().Err(err).Msg("fail to post keysign failure to thorchain")
			return nil, err
		} else {
			c.logger.Info().Str("tx_id", txID.String()).Msgf("post keysign failure to thorchain")
			return nil, fmt.Errorf("sent keysign failure to thorchain")
		}
	}
	c.logger.Error().Err(err).Msg("fail to sign tx")
	return nil, err
}

// GetAccount gets account by address in eth client
func (c *Client) GetAccount(pkey common.PubKey) (common.Account, error) {
	addr := c.GetAddress(pkey)
	nonce, err := c.GetNonce(addr)
	if err != nil {
		return common.Account{}, err
	}
	balance, err := c.client.BalanceAt(context.Background(), ecommon.HexToAddress(addr), nil)
	if err != nil {
		return common.Account{}, fmt.Errorf("fail to get account nonce: %w", err)
	}
	account := common.NewAccount(int64(nonce), 0, common.AccountCoins{common.AccountCoin{Amount: balance.Uint64(), Denom: "ETH.ETH"}})
	return account, nil
}

// BroadcastTx decodes tx using rlp and broadcasts too Ethereum chain
func (c *Client) BroadcastTx(stx stypes.TxOutItem, hexTx []byte) error {
	var tx *etypes.Transaction = &etypes.Transaction{}
	if err := json.Unmarshal(hexTx, tx); err != nil {
		return err
	}
	if err := c.client.SendTransaction(context.Background(), tx); err != nil {
		return err
	}
	c.accts.NonceInc(stx.VaultPubKey)
	return nil
}
