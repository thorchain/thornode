package bitcoin

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/txsort"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"gitlab.com/thorchain/txscript"

	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
)

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

func (c *Client) getLastOutput(inputTxId, sourceAddr string) (btcjson.Vout, error) {
	txHash, err := chainhash.NewHashFromStr(inputTxId)
	if err != nil {
		return btcjson.Vout{}, fmt.Errorf("fail to parse (%s) as chain hash: %w", inputTxId, err)
	}
	txRaw, err := c.client.GetRawTransactionVerbose(txHash)
	if err != nil {
		return btcjson.Vout{}, fmt.Errorf("fail to get the raw transactional: %w", err)
	}
	for _, item := range txRaw.Vout {
		for _, addr := range item.ScriptPubKey.Addresses {
			if addr == sourceAddr {
				return item, nil
			}
		}
	}
	return btcjson.Vout{}, errors.New("not found")
}

func getGasCoin(tx stypes.TxOutItem) common.Coin {
	return tx.MaxGas.ToCoins().GetCoin(common.BTCAsset)
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
	txes, err := c.utxoAccessor.GetUTXOs()
	if err != nil {
		return nil, fmt.Errorf("fail to get unspent UTXO")
	}

	redeemTx := wire.NewMsgTx(wire.TxVersion)
	totalAmt := float64(0)
	individualAmounts := make([]btcutil.Amount, len(txes))
	for idx, item := range txes {
		// double check that the utxo is still valid
		outputPoint := wire.NewOutPoint(&item.TxID, item.N)
		sourceTxIn := wire.NewTxIn(outputPoint, nil, nil)
		redeemTx.AddTxIn(sourceTxIn)
		totalAmt += item.Value
		amt, err := btcutil.NewAmount(item.Value)
		if err != nil {
			return nil, fmt.Errorf("fail to parse amount(%f): %w", item.Value, err)
		}
		individualAmounts[idx] = amt
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
	gasCoin := getGasCoin(tx)
	coinToCustomer := tx.Coins.GetCoin(common.BTCAsset)

	// pay to customer
	redeemTxOut := wire.NewTxOut(int64(coinToCustomer.Amount.Uint64()), buf)
	redeemTx.AddTxOut(redeemTxOut)

	// memo
	nullDataScript, err := txscript.NullDataScript([]byte(tx.Memo))
	if err != nil {
		return nil, fmt.Errorf("fail to generate null data script: %w", err)
	}
	redeemTx.AddTxOut(wire.NewTxOut(0, nullDataScript))

	// balance to ourselves
	// add output to pay the balance back ourselves
	balance := int64(total) - redeemTxOut.Value - int64(gasCoin.Amount.Uint64())
	if balance < 0 {
		return nil, errors.New("not enough balance to pay customer")
	}

	redeemTx.AddTxOut(wire.NewTxOut(balance, sourceScript))

	for idx := range redeemTx.TxIn {
		sigHashes := txscript.NewTxSigHashes(redeemTx)
		sig := c.ksWrapper.GetSignable(tx.VaultPubKey)
		witness, err := txscript.WitnessSignature(redeemTx, sigHashes, idx, int64(individualAmounts[idx]), sourceScript, txscript.SigHashAll, sig, true)
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
		engine, err := txscript.NewEngine(sourceScript, redeemTx, idx, flag, nil, nil, int64(individualAmounts[idx]))
		if err != nil {
			return nil, fmt.Errorf("fail to create engine: %w", err)
		}
		if err := engine.Execute(); err != nil {
			return nil, fmt.Errorf("fail to execute the script: %w", err)
		}
	}
	finalTx := txsort.Sort(redeemTx)
	var signedTx bytes.Buffer
	if err := finalTx.Serialize(&signedTx); err != nil {
		return nil, fmt.Errorf("fail to serialize tx to bytes: %w", err)
	}
	if err := c.saveNewUTXO(finalTx, balance, sourceScript, height); nil != err {
		return nil, fmt.Errorf("fail to save the new UTXO to storage: %w", err)
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
func (c *Client) saveNewUTXO(tx *wire.MsgTx, balance int64, script []byte, blockHeight int64) error {
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
	return c.utxoAccessor.AddUTXO(NewUnspentTransactionOutput(txID, uint32(n), amt.ToBTC(), blockHeight))
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
		return fmt.Errorf("fail to broadcast transaction to chain: %w", err)
	}
	c.logger.Info().Str("hash", txHash.String()).Msg("broadcast to BTC chain successfully")
	return nil
}
