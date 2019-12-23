package types

import (
	"gitlab.com/thorchain/thornode/common"
)

type TxOutItem struct {
	Chain       common.Chain   `json:"chain"`
	ToAddress   common.Address `json:"to"`
	VaultPubKey common.PubKey  `json:"vault_pubkey"`
	SeqNo       uint64         `json:"seq_no"`
	Coins       common.Coins   `json:"coins"`
	Memo        string         `json:"memo"`
	InHash      common.TxID    `json:"in_hash"`
	OutHash     common.TxID    `json:"out_hash"`
}

type TxArrayItem struct {
	Chain       common.Chain   `json:"chain"`
	ToAddress   common.Address `json:"to"`
	VaultPubKey common.PubKey  `json:"vault_pubkey"`
	Coin        common.Coin    `json:"coin"`
	Memo        string         `json:"memo"`
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
		InHash:      tx.InHash,
		OutHash:     tx.OutHash,
	}
}

type TxOut struct {
	Height  string        `json:"height"`
	Hash    string        `json:"hash"`
	Chain   common.Chain  `json:"chain"`
	TxArray []TxArrayItem `json:"tx_array"`
}
