package bitcoin

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/binance-chain/go-sdk/keys"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	tssp "gitlab.com/thorchain/tss/go-tss/tss"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
)

// Client observes bitcoin chain and allows to sign and broadcast tx
type Client struct {
	logger        zerolog.Logger
	cfg           config.ChainConfiguration
	client        *rpcclient.Client
	chain         common.Chain
	tssKeyManager keys.KeyManager
	privateKey    *btcec.PrivateKey
	utxoAccessor  UnspentTransactionOutputAccessor
	blockScanner  *blockscanner.BlockScanner
}

// NewClient generates a new Client
func NewClient(thorKeys *thorclient.Keys, cfg config.ChainConfiguration, server *tssp.TssServer, thorchainBridge *thorclient.ThorchainBridge, m *metrics.Metrics) (*Client, error) {
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         cfg.ChainHost,
		User:         cfg.UserName,
		Pass:         cfg.Password,
		DisableTLS:   cfg.DisableTLS,
		HTTPPostMode: cfg.HTTPostMode,
	}, nil)
	if err != nil {
		return nil, err
	}
	tssKm, err := tss.NewKeySign(server)
	if err != nil {
		return nil, fmt.Errorf("fail to create tss signer: %w", err)
	}
	thorPrivateKey, err := thorKeys.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("fail to get thor private key")
	}

	btcPrivateKey, err := getBTCPrivateKey(thorPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("fail to get private key for BTC chain: %w", err)
	}

	c := &Client{
		logger:        log.Logger.With().Str("module", "bitcoin").Logger(),
		cfg:           cfg,
		chain:         cfg.ChainID,
		client:        client,
		tssKeyManager: tssKm,
		privateKey:    btcPrivateKey,
	}

	var path string // if not set later, will in memory storage
	if len(c.cfg.BlockScanner.DBPath) > 0 {
		path = fmt.Sprintf("%s/%s", c.cfg.BlockScanner.DBPath, c.cfg.BlockScanner.ChainID)
	}
	storage, err := blockscanner.NewBlockScannerStorage(path)
	if err != nil {
		return c, errors.Wrap(err, "fail to create blockscanner storage")
	}

	c.blockScanner, err = blockscanner.NewBlockScanner(c.cfg.BlockScanner, storage, m, thorchainBridge, c)
	if err != nil {
		return c, errors.Wrap(err, "fail to create block scanner")
	}

	c.utxoAccessor, err = NewUTXOAccessor(path)
	if err != nil {
		return c, errors.Wrap(err, "fail to create utxo accessor")
	}

	return c, nil
}

// Start starts the block scanner
func (c *Client) Start(globalTxsQueue chan types.TxIn) {
	c.blockScanner.Start(globalTxsQueue)
}

// Stop stops the block scanner
func (c *Client) Stop() {
	c.blockScanner.Stop()
}

// GetChain returns BTC Chain
func (c *Client) GetChain() common.Chain {
	return common.BTCChain
}

// GetHeight returns current block height
func (c *Client) GetHeight() (int64, error) {
	return c.client.GetBlockCount()
}

// GetGasFee returns gas fee
func (c *Client) GetGasFee(count uint64) common.Gas {
	return common.Gas{} // TODO not implemented yet
}

// ValidateMetadata validates metadata
func (c *Client) ValidateMetadata(inter interface{}) bool {
	return true // TODO not implemented yet
}

// GetAddress returns address from pubkey
func (c *Client) GetAddress(poolPubKey common.PubKey) string {
	addr, err := poolPubKey.GetAddress(common.BTCChain)
	if err != nil {
		c.logger.Error().Err(err).Str("pool_pub_key", poolPubKey.String()).Msg("fail to get pool address")
		return ""
	}
	return addr.String()
}

// GetAccount returns account with balance for an address
func (c *Client) GetAccount(addr string) (common.Account, error) {
	return common.Account{}, fmt.Errorf("not implemented")
}

// OnObservedTxIn gets called from observer when we have a valid observation
// For bitcoin chain client we want to save the utxo we can spend later to sign
func (c *Client) OnObservedTxIn(txIn types.TxIn) {
	for _, tx := range txIn.TxArray {
		hash, err := chainhash.NewHashFromStr(tx.Tx)
		if err != nil {
			c.logger.Error().Err(err).Str("txID", tx.Tx).Msg("fail to add spendable utxo to storage")
			continue
		}
		value := float64(tx.Coins.GetCoin(common.BTCAsset).Amount.Uint64()) / common.One
		blockHeight, err := strconv.ParseInt(txIn.BlockHeight, 10, 64)
		if err != nil {
			c.logger.Error().Err(err).Str("txID", tx.Tx).Msg("fail to add spendable utxo to storage")
			continue
		}
		utxo := NewUnspentTransactionOutput(*hash, 0, value, blockHeight)
		err = c.utxoAccessor.AddUTXO(utxo)
		if err != nil {
			c.logger.Error().Err(err).Str("txID", tx.Tx).Msg("fail to add spendable utxo to storage")
			continue
		}
	}
}

// FetchTxs retrieves txs for a block height
func (c *Client) FetchTxs(height int64) (types.TxIn, error) {
	block, err := c.getBlock(height)
	if err != nil {
		return types.TxIn{}, errors.Wrap(err, "fail to get block")
	}
	txs, err := c.extractTxs(block)
	if err != nil {
		return types.TxIn{}, errors.Wrap(err, "fail to extract txs from block")
	}
	return txs, nil
}

// getBlock retrieves block from chain for a block height
func (c *Client) getBlock(height int64) (*btcjson.GetBlockVerboseResult, error) {
	hash, err := c.client.GetBlockHash(height)
	if err != nil {
		return &btcjson.GetBlockVerboseResult{}, err
	}
	return c.client.GetBlockVerboseTx(hash)
}

// extractTxs extracts txs from a block to type TxIn
func (c *Client) extractTxs(block *btcjson.GetBlockVerboseResult) (types.TxIn, error) {
	txIn := types.TxIn{
		BlockHeight: strconv.FormatInt(block.Height, 10),
		Chain:       c.GetChain(),
	}
	var txItems []types.TxInItem
	for _, tx := range block.RawTx {
		if c.ignoreTx(&tx) {
			continue
		}
		sender, err := c.getSender(&tx)
		if err != nil {
			return types.TxIn{}, errors.Wrap(err, "fail to get sender from tx")
		}
		memo, err := c.getMemo(&tx)
		if err != nil {
			return types.TxIn{}, errors.Wrap(err, "fail to get memo from tx")
		}
		gas, err := c.getGas(&tx)
		if err != nil {
			return types.TxIn{}, errors.Wrap(err, "fail to get gas from tx")
		}
		amount := uint64(tx.Vout[0].Value * common.One)
		txItems = append(txItems, types.TxInItem{
			Tx:     tx.Txid,
			Sender: sender,
			To:     tx.Vout[0].ScriptPubKey.Addresses[0],
			Coins: common.Coins{
				common.NewCoin(common.BTCAsset, sdk.NewUint(amount)),
			},
			Memo: memo,
			Gas:  gas,
		})
	}
	txIn.TxArray = txItems
	txIn.Count = strconv.Itoa(len(txItems))
	return txIn, nil
}

// ignoreTx checks if we can already ignore a tx according to preset rules
//
// we expect array of "vout" for a BTC to have this format
// vout:0 is our vault
// vout:1 is any any change back to themselves
// vout:2 is OP_RETURN (first 80 bytes)
// vout:3 is OP_RETURN (next 80 bytes)
//
// Rules to ignore a tx are:
// - vout:0 doesn't have coins (value)
// - vout:0 doesn't have address
// - count vouts > 4
// - count vouts with coins (value) > 2
// - no OP_RETURN presents in tx vouts
//
func (c *Client) ignoreTx(tx *btcjson.TxRawResult) bool {
	if len(tx.Vin) == 0 || len(tx.Vout) == 0 || len(tx.Vout) > 4 {
		return true
	}
	if tx.Vout[0].Value == 0 || tx.Vin[0].Txid == "" {
		return true
	}
	// TODO check what we do if get multiple addresses
	if len(tx.Vout[0].ScriptPubKey.Addresses) != 1 {
		return true
	}
	countOPReturn := 0
	countWithCoins := 0
	for _, vout := range tx.Vout {
		if vout.Value > 0 {
			countWithCoins++
		}
		if strings.HasPrefix(vout.ScriptPubKey.Asm, "OP_RETURN") {
			countOPReturn++
		}
	}
	if countOPReturn == 0 || countOPReturn > 2 || countWithCoins > 2 {
		return true
	}
	return false
}

// getSender returns sender address for a btc tx, using vin:0
func (c *Client) getSender(tx *btcjson.TxRawResult) (string, error) {
	if len(tx.Vin) == 0 {
		return "", fmt.Errorf("no vin available in tx")
	}
	txHash, err := chainhash.NewHashFromStr(tx.Vin[0].Txid)
	if err != nil {
		return "", fmt.Errorf("fail to get tx hash from tx id string")
	}
	vinTx, err := c.client.GetRawTransactionVerbose(txHash)
	if err != nil {
		return "", fmt.Errorf("fail to query raw tx from btcd")
	}
	vout := vinTx.Vout[tx.Vin[0].Vout]
	if len(vout.ScriptPubKey.Addresses) == 0 {
		return "", fmt.Errorf("no address available in vout")
	}
	return vout.ScriptPubKey.Addresses[0], nil
}

// getMemo returns memo for a btc tx, using vout OP_RETURN
func (c *Client) getMemo(tx *btcjson.TxRawResult) (string, error) {
	var opreturns string
	for _, vout := range tx.Vout {
		if strings.HasPrefix(vout.ScriptPubKey.Asm, "OP_RETURN") {
			opreturn := strings.Split(vout.ScriptPubKey.Asm, " ")
			opreturns += opreturn[1]
		}
	}
	decoded, err := hex.DecodeString(opreturns)
	if err != nil {
		return "", fmt.Errorf("fail to decode OP_RETURN string")
	}
	return string(decoded), nil
}

// getGas returns gas for a btc tx (sum vin - sum vout)
func (c *Client) getGas(tx *btcjson.TxRawResult) (common.Gas, error) {
	var sumVin float64 = 0
	for _, vin := range tx.Vin {
		txHash, err := chainhash.NewHashFromStr(tx.Vin[0].Txid)
		if err != nil {
			return common.Gas{}, fmt.Errorf("fail to get tx hash from tx id string")
		}
		vinTx, err := c.client.GetRawTransactionVerbose(txHash)
		if err != nil {
			return common.Gas{}, fmt.Errorf("fail to query raw tx from btcd")
		}
		sumVin += vinTx.Vout[vin.Vout].Value
	}
	var sumVout float64 = 0
	for _, vout := range tx.Vout {
		sumVout += vout.Value
	}
	totalGas := uint64((sumVin - sumVout) * common.One)
	return common.Gas{
		common.NewCoin(common.BTCAsset, sdk.NewUint(totalGas)),
	}, nil
}
