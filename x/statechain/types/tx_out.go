package types

import (
	ctypes "gitlab.com/thorchain/bepswap/observe/common/types"
)

type TxOut struct {
	Height  string `json:"height"`
	Hash    string `json:"hash"`
	TxArray []struct {
		To    string `json:"to"`
		Coins []ctypes.Coin `json:"coins"`
	} `json:"tx_array"`
}
