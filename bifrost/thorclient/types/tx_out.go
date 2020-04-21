package types

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"gitlab.com/thorchain/thornode/common"
)

type TxOutItem struct {
	Chain       common.Chain   `json:"chain"`
	ToAddress   common.Address `json:"to"`
	VaultPubKey common.PubKey  `json:"vault_pubkey"`
	SeqNo       uint64         `json:"seq_no"`
	Coins       common.Coins   `json:"coins"`
	Memo        string         `json:"memo"`
	MaxGas      common.Gas     `json:"max_gas"`
	InHash      common.TxID    `json:"in_hash"`
	OutHash     common.TxID    `json:"out_hash"`
}

func (tx TxOutItem) Hash() string {
	str := fmt.Sprintf("%s|%s|%s|%s|%s|%s", tx.Chain, tx.ToAddress, tx.VaultPubKey, tx.Coins, tx.Memo, tx.InHash)
	return fmt.Sprintf("%X", sha256.Sum256([]byte(str)))
}

func (tx1 TxOutItem) Equals(tx2 TxOutItem) bool {
	if !tx1.Chain.Equals(tx2.Chain) {
		return false
	}
	if !tx1.VaultPubKey.Equals(tx2.VaultPubKey) {
		return false
	}
	if !tx1.ToAddress.Equals(tx2.ToAddress) {
		return false
	}
	if !tx1.Coins.Equals(tx2.Coins) {
		return false
	}
	if !tx1.InHash.Equals(tx2.InHash) {
		return false
	}
	if !strings.EqualFold(tx1.Memo, tx2.Memo) {
		return false
	}
	return true
}

type TxArrayItem struct {
	Chain       common.Chain   `json:"chain"`
	ToAddress   common.Address `json:"to"`
	VaultPubKey common.PubKey  `json:"vault_pubkey"`
	Coin        common.Coin    `json:"coin"`
	Memo        string         `json:"memo"`
	MaxGas      common.Gas     `json:"max_gas"`
	InHash      common.TxID    `json:"in_hash"`
	OutHash     common.TxID    `json:"out_hash"`
}

func (tx TxArrayItem) TxOutItem() TxOutItem {
	return TxOutItem{
		Chain:       tx.Chain,
		ToAddress:   tx.ToAddress,
		VaultPubKey: tx.VaultPubKey,
		Coins:       common.Coins{tx.Coin},
		Memo:        tx.Memo,
		MaxGas:      tx.MaxGas,
		InHash:      tx.InHash,
		OutHash:     tx.OutHash,
	}
}

type TxOut struct {
	Height  int64         `json:"height,string"`
	Hash    string        `json:"hash"`
	Chain   common.Chain  `json:"chain"`
	TxArray []TxArrayItem `json:"tx_array"`
}

type ChainsTxOut struct {
	Chains map[common.Chain]TxOut `json:"chains"`
}

// GetKey will return a key we can used it to save the infor to level db
func (tai TxArrayItem) GetKey(height int64) string {
	return fmt.Sprintf("%d-%s-%s-%s-%s-%s", height, tai.InHash, tai.VaultPubKey, tai.Memo, tai.Coin, tai.ToAddress)
}
