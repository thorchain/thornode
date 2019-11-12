package types

import "gitlab.com/thorchain/bepswap/thornode/common"

type TxArrayItem struct {
	PoolAddress common.PubKey `json:"pool_address"`
	SeqNo       string        `json:"seq_no"`
	To          string        `json:"to"`
	Coin        common.Coin   `json:"coin"`
	Memo        string        `json:"memo"`
}
type TxOut struct {
	Height  string        `json:"height"`
	Hash    string        `json:"hash"`
	Chain   common.Chain  `json:"chain"`
	TxArray []TxArrayItem `json:"tx_array"`
}
