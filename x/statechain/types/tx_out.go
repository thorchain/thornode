package types

import "gitlab.com/thorchain/bepswap/common"

type TxOut struct {
	Height  string `json:"height"`
	Hash    string `json:"hash"`
	TxArray []struct {
		PoolAddress common.BnbAddress `json:"pool_address"`
		To          string            `json:"to"`
		Coins       common.Coins      `json:"coins"`
	} `json:"tx_array"`
}
