package types

import "gitlab.com/thorchain/bepswap/thornode/common"

type TxOut struct {
	Height  string `json:"height"`
	Hash    string `json:"hash"`
	TxArray []struct {
		PoolAddress common.PubKey `json:"pool_address"`
		To          string        `json:"to"`
		Coins       common.Coins  `json:"coins"`
	} `json:"tx_array"`
}
