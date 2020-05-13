package bitcoin

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tssp "gitlab.com/thorchain/tss/go-tss/tss"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	btypes "gitlab.com/thorchain/thornode/bifrost/blockscanner/types"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
)

// BlockCacheSize the number of block meta that get store in storage.
const BlockCacheSize = 100

// Client observes bitcoin chain and allows to sign and broadcast tx
type Client struct {
	logger            zerolog.Logger
	cfg               config.ChainConfiguration
	client            *rpcclient.Client
	chain             common.Chain
	privateKey        *btcec.PrivateKey
	blockScanner      *blockscanner.BlockScanner
	blockMetaAccessor BlockMetaAccessor
	ksWrapper         *KeySignWrapper
	bridge            *thorclient.ThorchainBridge
	globalErrataQueue chan<- types.ErrataBlock
	nodePubKey        common.PubKey
}

// NewClient generates a new Client
func NewClient(thorKeys *thorclient.Keys, cfg config.ChainConfiguration, server *tssp.TssServer, bridge *thorclient.ThorchainBridge, m *metrics.Metrics) (*Client, error) {
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         cfg.RPCHost,
		User:         cfg.UserName,
		Pass:         cfg.Password,
		DisableTLS:   cfg.DisableTLS,
		HTTPPostMode: cfg.HTTPostMode,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("fail to create bitcoin rpc client: %w", err)
	}
	tssKm, err := tss.NewKeySign(server)
	if err != nil {
		return nil, fmt.Errorf("fail to create tss signer: %w", err)
	}
	thorPrivateKey, err := thorKeys.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("fail to get THORChain private key: %w", err)
	}

	btcPrivateKey, err := getBTCPrivateKey(thorPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("fail to convert private key for BTC: %w", err)
	}
	ksWrapper, err := NewKeySignWrapper(btcPrivateKey, bridge, tssKm)
	if err != nil {
		return nil, fmt.Errorf("fail to create keysign wrapper: %w", err)
	}
	nodePubKey, err := common.NewPubKeyFromCrypto(thorKeys.GetSignerInfo().GetPubKey())
	if err != nil {
		return nil, fmt.Errorf("fail to get the node pubkey: %w", err)
	}

	c := &Client{
		logger:     log.Logger.With().Str("module", "bitcoin").Logger(),
		cfg:        cfg,
		chain:      cfg.ChainID,
		client:     client,
		privateKey: btcPrivateKey,
		ksWrapper:  ksWrapper,
		bridge:     bridge,
		nodePubKey: nodePubKey,
	}

	var path string // if not set later, will in memory storage
	if len(c.cfg.BlockScanner.DBPath) > 0 {
		path = fmt.Sprintf("%s/%s", c.cfg.BlockScanner.DBPath, c.cfg.BlockScanner.ChainID)
	}
	storage, err := blockscanner.NewBlockScannerStorage(path)
	if err != nil {
		return c, fmt.Errorf("fail to create blockscanner storage: %w", err)
	}

	c.blockScanner, err = blockscanner.NewBlockScanner(c.cfg.BlockScanner, storage, m, bridge, c)
	if err != nil {
		return c, fmt.Errorf("fail to create block scanner: %w", err)
	}

	c.blockMetaAccessor, err = NewLevelDBBlockMetaAccessor(storage.GetInternalDb())
	if err != nil {
		return c, fmt.Errorf("fail to create utxo accessor: %w", err)
	}

	return c, nil
}

// Start starts the block scanner
func (c *Client) Start(globalTxsQueue chan types.TxIn, globalErrataQueue chan types.ErrataBlock) {
	c.blockScanner.Start(globalTxsQueue)
	c.globalErrataQueue = globalErrataQueue
}

// Stop stops the block scanner
func (c *Client) Stop() {
	c.blockScanner.Stop()
}

// GetConfig - get the chain configuration
func (c *Client) GetConfig() config.ChainConfiguration {
	return c.cfg
}

// GetChain returns BTC Chain
func (c *Client) GetChain() common.Chain {
	return common.BTCChain
}

// GetHeight returns current block height
func (c *Client) GetHeight() (int64, error) {
	return c.client.GetBlockCount()
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
func (c *Client) GetAccount(pkey common.PubKey) (common.Account, error) {
	acct := common.Account{}
	blockMetas, err := c.blockMetaAccessor.GetBlockMetas()
	if err != nil {
		return acct, fmt.Errorf("fail to get block meta: %w", err)
	}
	total := 0.0
	for _, item := range blockMetas {
		for _, utxo := range item.GetUTXOs(pkey) {
			total += utxo.Value
		}
	}

	totalAmt, err := btcutil.NewAmount(total)
	if err != nil {
		return acct, fmt.Errorf("fail to convert total amount: %w", err)
	}
	return common.NewAccount(0, 0, common.AccountCoins{
		common.AccountCoin{
			Amount: uint64(totalAmt),
			Denom:  common.BTCAsset.String(),
		},
	}), nil
}

// OnObservedTxIn gets called from observer when we have a valid observation
// For bitcoin chain client we want to save the utxo we can spend later to sign
func (c *Client) OnObservedTxIn(txIn types.TxInItem, blockHeight int64) {
	hash, err := chainhash.NewHashFromStr(txIn.Tx)
	if err != nil {
		c.logger.Error().Err(err).Str("txID", txIn.Tx).Msg("fail to add spendable utxo to storage")
		return
	}
	value := float64(txIn.Coins.GetCoin(common.BTCAsset).Amount.Uint64()) / common.One
	blockMeta, err := c.blockMetaAccessor.GetBlockMeta(blockHeight)
	if nil != err {
		c.logger.Err(err).Msgf("fail to get block meta on block height(%d)", blockHeight)
	}
	if nil == blockMeta {
		c.logger.Error().Msgf("can't get block meta for height: %d", blockHeight)
		return
	}
	utxo := NewUnspentTransactionOutput(*hash, 0, value, blockHeight, txIn.ObservedVaultPubKey)
	blockMeta.AddUTXO(utxo)
	if err := c.blockMetaAccessor.SaveBlockMeta(blockHeight, blockMeta); err != nil {
		c.logger.Err(err).Msgf("fail to save block meta to storage,block height(%d)", blockHeight)
	}
}

func (c *Client) processReorg(block *btcjson.GetBlockVerboseTxResult) error {
	previousHeight := block.Height - 1
	prevBlockMeta, err := c.blockMetaAccessor.GetBlockMeta(previousHeight)
	if err != nil {
		return fmt.Errorf("fail to get block meta of height(%d) : %w", previousHeight, err)
	}
	if prevBlockMeta == nil {
		return nil
	}
	// the block's previous hash need to be the same as the block hash chain client recorded in block meta
	// blockMetas[PreviousHeight].BlockHash == Block.PreviousHash
	if strings.EqualFold(prevBlockMeta.BlockHash, block.PreviousHash) {
		return nil
	}

	c.logger.Info().Msgf("re-org detected, current block height:%d ,previous block hash is : %s , however block meta at height: %d, block hash is %s", block.Height, block.PreviousHash, prevBlockMeta.Height, prevBlockMeta.BlockHash)
	return c.reConfirmTx()
}

// reConfirmTx will be kicked off only when chain client detected a re-org on bitcoin chain
// it will read through all the block meta data from local storage , and go through all the UTXOes.
// For each UTXO , it will send a RPC request to bitcoin chain , double check whether the TX exist or not
// if the tx still exist , then it is all good, if a transaction previous we detected , however doesn't exist anymore , that means
// the transaction had been removed from chain,  chain client should report to thorchain
func (c *Client) reConfirmTx() error {
	blockMetas, err := c.blockMetaAccessor.GetBlockMetas()
	if err != nil {
		return fmt.Errorf("fail to get block metas from local storage: %w", err)
	}

	for _, blockMeta := range blockMetas {
		var errataTxs []types.ErrataTx
		for _, utxo := range blockMeta.UnspentTransactionOutputs {
			txID := utxo.TxID.String()
			if c.confirmTx(&utxo.TxID) {
				c.logger.Info().Msgf("block height: %d, tx: %s still exist", blockMeta.Height, txID)
				continue
			}
			// this means the tx doesn't exist in chain ,thus should errata it
			errataTxs = append(errataTxs, types.ErrataTx{
				TxID:  common.TxID(txID),
				Chain: common.BTCChain,
			})
			// remove the UTXO from block meta , so signer will not spend it
			blockMeta.RemoveUTXO(utxo.GetKey())
		}
		if len(errataTxs) == 0 {
			continue
		}
		c.globalErrataQueue <- types.ErrataBlock{
			Height: blockMeta.Height,
			Txs:    errataTxs,
		}
		// Let's get the block again to fix the block hash
		r, err := c.getBlock(blockMeta.Height)
		if err != nil {
			c.logger.Err(err).Msgf("fail to get block verbose tx result: %d", blockMeta.Height)
		}
		blockMeta.PreviousHash = r.PreviousHash
		blockMeta.BlockHash = r.Hash
		if err := c.blockMetaAccessor.SaveBlockMeta(blockMeta.Height, blockMeta); err != nil {
			c.logger.Err(err).Msgf("fail to save block meta of height: %d ", blockMeta.Height)
		}
	}
	return nil
}

// confirmTx check a tx is valid on chain post reorg
func (c *Client) confirmTx(txHash *chainhash.Hash) bool {
	// first check if tx is in mempool, just signed it for example
	// if no error it means its valid mempool tx and move on
	_, err := c.client.GetMempoolEntry(txHash.String())
	if err == nil {
		return true
	}
	// then get raw tx and check if it has confirmations or not
	// if no confirmation and not in mempool then invalid
	result, err := c.client.GetTransaction(txHash)
	if err != nil {
		if rpcErr, ok := err.(*btcjson.RPCError); ok && rpcErr.Code == btcjson.ErrRPCNoTxInfo {
			return false
		}
		return true
	}
	if result.Confirmations == 0 {
		return false
	}
	return true
}

// FetchTxs retrieves txs for a block height
func (c *Client) FetchTxs(height int64) (types.TxIn, error) {
	block, err := c.getBlock(height)
	if err != nil {
		time.Sleep(c.cfg.BlockScanner.BlockHeightDiscoverBackoff)
		if rpcErr, ok := err.(*btcjson.RPCError); ok && rpcErr.Code == btcjson.ErrRPCInvalidParameter {
			// this means the tx had been broadcast to chain, it must be another signer finished quicker then us
			return types.TxIn{}, btypes.UnavailableBlock
		}
		return types.TxIn{}, fmt.Errorf("fail to get block: %w", err)
	}
	if err := c.processReorg(block); err != nil {
		c.logger.Err(err).Msg("fail to process bitcoin re-org")
	}
	blockMeta, err := c.blockMetaAccessor.GetBlockMeta(block.Height)
	if err != nil {
		return types.TxIn{}, fmt.Errorf("fail to get block meta from storage: %w", err)
	}
	if blockMeta == nil {
		blockMeta = NewBlockMeta(block.PreviousHash, block.Height, block.Hash)
	} else {
		blockMeta.PreviousHash = block.PreviousHash
		blockMeta.BlockHash = block.Hash
	}

	if err := c.blockMetaAccessor.SaveBlockMeta(block.Height, blockMeta); err != nil {
		return types.TxIn{}, fmt.Errorf("fail to save block meta into storage: %w", err)
	}
	pruneHeight := height - BlockCacheSize
	if pruneHeight > 0 {
		defer func() {
			if err := c.blockMetaAccessor.PruneBlockMeta(pruneHeight); err != nil {
				c.logger.Err(err).Msgf("fail to prune block meta, height(%d)", pruneHeight)
			}
		}()
	}
	txs, err := c.extractTxs(block)
	if err != nil {
		return types.TxIn{}, fmt.Errorf("fail to extract txs from block: %w", err)
	}
	return txs, nil
}

// getBlock retrieves block from chain for a block height
func (c *Client) getBlock(height int64) (*btcjson.GetBlockVerboseTxResult, error) {
	hash, err := c.client.GetBlockHash(height)
	if err != nil {
		return &btcjson.GetBlockVerboseTxResult{}, err
	}
	return c.client.GetBlockVerboseTx(hash)
}

// extractTxs extracts txs from a block to type TxIn
func (c *Client) extractTxs(block *btcjson.GetBlockVerboseTxResult) (types.TxIn, error) {
	txIn := types.TxIn{
		BlockHeight: strconv.FormatInt(block.Height, 10),
		Chain:       c.GetChain(),
	}
	var txItems []types.TxInItem
	for _, tx := range block.Tx {
		if c.ignoreTx(&tx) {
			continue
		}
		sender, err := c.getSender(&tx)
		if err != nil {
			return types.TxIn{}, fmt.Errorf("fail to get sender from tx: %w", err)
		}
		memo, err := c.getMemo(&tx)
		if err != nil {
			return types.TxIn{}, fmt.Errorf("fail to get memo from tx: %w", err)
		}
		gas, err := c.getGas(&tx)
		if err != nil {
			return types.TxIn{}, fmt.Errorf("fail to get gas from tx: %w", err)
		}

		output := c.getOutput(sender, &tx)
		amount := uint64(output.Value * common.One)
		txItems = append(txItems, types.TxInItem{
			Tx:     tx.Txid,
			Sender: sender,
			To:     output.ScriptPubKey.Addresses[0],
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
// OP_RETURN is mandatory only on inbound tx
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
	countWithOutput := 0
	for _, vout := range tx.Vout {
		if vout.Value > 0 {
			countWithOutput++
		}
	}
	if countWithOutput > 2 {
		return true
	}
	return false
}

// getOutput retrieve the correct output for both inbound
// outbound tx.
// logic is if FROM == TO then its an outbound change output
// back to the vault and we need to select the other output
// as Bifrost already filtered the txs to only have here
// txs with max 2 outputs with values
func (c *Client) getOutput(sender string, tx *btcjson.TxRawResult) btcjson.Vout {
	for _, vout := range tx.Vout {
		if vout.Value > 0 && vout.ScriptPubKey.Addresses[0] != sender {
			return vout
		}
	}
	return btcjson.Vout{}
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
	var sumVin uint64 = 0
	for _, vin := range tx.Vin {
		txHash, err := chainhash.NewHashFromStr(vin.Txid)
		if err != nil {
			return common.Gas{}, fmt.Errorf("fail to get tx hash from tx id string")
		}
		vinTx, err := c.client.GetRawTransactionVerbose(txHash)
		if err != nil {
			return common.Gas{}, fmt.Errorf("fail to query raw tx from bitcoin node")
		}
		sumVin += uint64(vinTx.Vout[vin.Vout].Value * common.One)
	}
	var sumVout uint64 = 0
	for _, vout := range tx.Vout {
		sumVout += uint64(vout.Value * common.One)
	}
	totalGas := sumVin - sumVout
	return common.Gas{
		common.NewCoin(common.BTCAsset, sdk.NewUint(totalGas)),
	}, nil
}
