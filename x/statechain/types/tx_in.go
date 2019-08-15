package types

import (
	ctypes "gitlab.com/thorchain/bepswap/observe/common/types"
)

type TxIn struct {
	BlockHeight string     `json:"blockHeight"`
	Count       string     `json:"count"`
	TxArray     []TxInItem `json:"txArray"`
}

type TxInItem struct {
	Tx     string        `json:"tx"`
	Memo   string        `json:"MEMO"`
	Sender string        `json:"sender"`
	Coins  []ctypes.Coin `json:"coins"`
}
