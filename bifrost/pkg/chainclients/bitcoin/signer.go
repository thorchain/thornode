package bitcoin

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/txsort"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"gitlab.com/thorchain/txscript"

	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
)

// SatsPervBytes it should be enough , this one will only be used if signer can't find any previous UTXO , and fee info from local storage.
const SatsPervBytes = 25

func getBTCPrivateKey(key crypto.PrivKey) (*btcec.PrivateKey, error) {
	priKey, ok := key.(secp256k1.PrivKeySecp256k1)
	if !ok {
		return nil, errors.New("invalid private key type")
	}
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), priKey[:])
	return privateKey, nil
}

func (c *Client) getChainCfg() *chaincfg.Params {
	cn := common.GetCurrentChainNetwork()
	switch cn {
	case common.MockNet:
		return &chaincfg.RegressionNetParams
	case common.TestNet:
		return &chaincfg.TestNet3Params
	case common.MainNet:
		return &chaincfg.MainNetParams
	}
	return nil
}

func (c *Client) getGasCoin(tx stypes.TxOutItem, vSize int64) common.Coin {
	if !tx.MaxGas.IsEmpty() {
		return tx.MaxGas.ToCoins().GetCoin(common.BTCAsset)
	}
	gasRate := int64(SatsPervBytes)
	fee, vBytes, err := c.utxoAccessor.GetTransactionFee()
	if err != nil {
		c.logger.Error().Err(err).Msg("fail to get previous transaction fee from local storage")
		return common.NewCoin(common.BTCAsset, sdk.NewUint(uint64(vSize*gasRate)))
	}
	if fee != 0.0 && vSize != 0 {
		amt, err := btcutil.NewAmount(fee)
		if err != nil {
			c.logger.Err(err).Msg("fail to convert amount from float64 to int64")
		} else {
			gasRate = int64(amt) / int64(vBytes) // sats per vbyte
		}
	}
	return common.NewCoin(common.BTCAsset, sdk.NewUint(uint64(gasRate*vSize)))
}

// SignTx is going to generate the outbound transaction, and also sign it
func (c *Client) SignTx(tx stypes.TxOutItem, height int64) ([]byte, error) {
	if !tx.Chain.Equals(common.BTCChain) {
		return nil, errors.New("not BTC chain")
	}
	sourceAddr, err := tx.VaultPubKey.GetAddress(common.BTCChain)
	if err != nil {
		return nil, fmt.Errorf("fail to get source address: %w", err)
	}

	addr, err := btcutil.DecodeAddress(sourceAddr.String(), c.getChainCfg())
	if err != nil {
		return nil, fmt.Errorf("fail to decode source address(%s): %w", sourceAddr.String(), err)
	}
	sourceScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, fmt.Errorf("fail to get source pay to address script: %w", err)
	}
	txes, err := c.utxoAccessor.GetUTXOs(tx.VaultPubKey)
	if err != nil {
		return nil, fmt.Errorf("fail to get unspent UTXO")
	}
	redeemTx := wire.NewMsgTx(wire.TxVersion)
	totalAmt := float64(0)
	individualAmounts := make(map[chainhash.Hash]btcutil.Amount, len(txes))
	for _, item := range txes {
		// double check that the utxo is still valid
		outputPoint := wire.NewOutPoint(&item.TxID, item.N)
		sourceTxIn := wire.NewTxIn(outputPoint, nil, nil)
		redeemTx.AddTxIn(sourceTxIn)
		totalAmt += item.Value
		amt, err := btcutil.NewAmount(item.Value)
		if err != nil {
			return nil, fmt.Errorf("fail to parse amount(%f): %w", item.Value, err)
		}
		individualAmounts[item.TxID] = amt
	}

	outputAddr, err := btcutil.DecodeAddress(tx.ToAddress.String(), c.getChainCfg())
	if err != nil {
		return nil, fmt.Errorf("fail to decode next address: %w", err)
	}
	buf, err := txscript.PayToAddrScript(outputAddr)
	if err != nil {
		return nil, fmt.Errorf("fail to get pay to address script: %w", err)
	}

	total, err := btcutil.NewAmount(totalAmt)
	if err != nil {
		return nil, fmt.Errorf("fail to parse total amount(%f),err: %w", totalAmt, err)
	}
	vSize := mempool.GetTxVirtualSize(btcutil.NewTx(redeemTx))
	gasCoin := c.getGasCoin(tx, vSize)
	gasAmt := btcutil.Amount(int64(gasCoin.Amount.Uint64()))
	if err := c.utxoAccessor.UpsertTransactionFee(gasAmt.ToBTC(), int32(vSize)); err != nil {
		c.logger.Err(err).Msg("fail to save gas info to UTXO storage")
	}
	coinToCustomer := tx.Coins.GetCoin(common.BTCAsset)

	// pay to customer
	redeemTxOut := wire.NewTxOut(int64(coinToCustomer.Amount.Uint64()), buf)
	redeemTx.AddTxOut(redeemTxOut)

	if len(tx.Memo) != 0 {
		// memo
		nullDataScript, err := txscript.NullDataScript([]byte(tx.Memo))
		if err != nil {
			return nil, fmt.Errorf("fail to generate null data script: %w", err)
		}
		redeemTx.AddTxOut(wire.NewTxOut(0, nullDataScript))
	}
	// balance to ourselves
	// add output to pay the balance back ourselves
	balance := int64(total) - redeemTxOut.Value - int64(gasCoin.Amount.Uint64())
	if balance < 0 {
		return nil, errors.New("not enough balance to pay customer")
	}
	if balance > 0 {
		redeemTx.AddTxOut(wire.NewTxOut(balance, sourceScript))
	}
	txsort.InPlaceSort(redeemTx)
	for idx, txIn := range redeemTx.TxIn {
		sigHashes := txscript.NewTxSigHashes(redeemTx)
		sig := c.ksWrapper.GetSignable(tx.VaultPubKey)
		outputAmount := int64(individualAmounts[txIn.PreviousOutPoint.Hash])
		witness, err := txscript.WitnessSignature(redeemTx, sigHashes, idx, outputAmount, sourceScript, txscript.SigHashAll, sig, true)
		if err != nil {
			var keysignError tss.KeysignError
			if errors.As(err, &keysignError) {
				if len(keysignError.Blame.BlameNodes) == 0 {
					// TSS doesn't know which node to blame
					return nil, err
				}

				// key sign error forward the keysign blame to thorchain
				txID, err := c.bridge.PostKeysignFailure(keysignError.Blame, height, tx.Memo, tx.Coins)
				if err != nil {
					c.logger.Error().Err(err).Msg("fail to post keysign failure to thorchain")
					return nil, err
				} else {
					c.logger.Info().Str("tx_id", txID.String()).Msgf("post keysign failure to thorchain")
					return nil, fmt.Errorf("sent keysign failure to thorchain")
				}
			}
			return nil, fmt.Errorf("fail to get witness: %w", err)
		}

		redeemTx.TxIn[idx].Witness = witness
		flag := txscript.StandardVerifyFlags
		engine, err := txscript.NewEngine(sourceScript, redeemTx, idx, flag, nil, nil, outputAmount)
		if err != nil {
			return nil, fmt.Errorf("fail to create engine: %w", err)
		}
		if err := engine.Execute(); err != nil {
			return nil, fmt.Errorf("fail to execute the script: %w", err)
		}
	}

	var signedTx bytes.Buffer
	if err := redeemTx.Serialize(&signedTx); err != nil {
		return nil, fmt.Errorf("fail to serialize tx to bytes: %w", err)
	}
	// only send the balance back to ourselves
	if balance > 0 {
		if err := c.saveNewUTXO(redeemTx, balance, sourceScript, height, tx.VaultPubKey); nil != err {
			return nil, fmt.Errorf("fail to save the new UTXO to storage: %w", err)
		}
	}
	if err := c.removeSpentUTXO(txes); err != nil {
		return nil, fmt.Errorf("fail to remove already spent transaction output: %w", err)
	}
	return signedTx.Bytes(), nil
}

func (c *Client) removeSpentUTXO(txs []UnspentTransactionOutput) error {
	for _, item := range txs {
		key := item.GetKey()
		if err := c.utxoAccessor.RemoveUTXO(key); err != nil {
			return fmt.Errorf("fail to remove unspent transaction output(%s): %w", key, err)
		}
	}
	return nil
}

// saveUTXO save the newly created UTXO which transfer balance back our own address to storage
func (c *Client) saveNewUTXO(tx *wire.MsgTx, balance int64, script []byte, blockHeight int64, pubKey common.PubKey) error {
	txID := tx.TxHash()
	n := 0
	// find the position of output that we send balance back to ourselves
	for idx, item := range tx.TxOut {
		if item.Value == balance && bytes.Equal(script, item.PkScript) {
			n = idx
			break
		}
	}
	amt := btcutil.Amount(balance)
	return c.utxoAccessor.AddUTXO(NewUnspentTransactionOutput(txID, uint32(n), amt.ToBTC(), blockHeight, pubKey))
}

// BroadcastTx will broadcast the given payload to BTC chain
func (c *Client) BroadcastTx(txOut stypes.TxOutItem, payload []byte) error {
	redeemTx := wire.NewMsgTx(wire.TxVersion)
	buf := bytes.NewBuffer(payload)
	if err := redeemTx.Deserialize(buf); err != nil {
		return fmt.Errorf("fail to deserialize payload: %w", err)
	}
	txHash, err := c.client.SendRawTransaction(redeemTx, true)
	if err != nil {

		if rpcErr, ok := err.(*btcjson.RPCError); ok && rpcErr.Code == btcjson.ErrRPCTxAlreadyInChain {
			// this means the tx had been broadcast to chain, it must be another signer finished quicker then us
			return nil
		}
		return fmt.Errorf("fail to broadcast transaction to chain: %w", err)
	}
	c.logger.Info().Str("hash", txHash.String()).Msg("broadcast to BTC chain successfully")
	return nil
}
