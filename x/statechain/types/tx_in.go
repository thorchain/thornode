package types

import "gitlab.com/thorchain/bepswap/common"

type TxIn struct {
	BlockHeight string     `json:"blockHeight"`
	Count       string     `json:"count"`
	TxArray     []TxInItem `json:"txArray"`
}

type TxInItem struct {
	Tx     string       `json:"tx"`
	Memo   string       `json:"MEMO"`
	Sender string       `json:"sender"`
	Coins  common.Coins `json:"coins"`
}
type TxInStatus byte

const (
	Processing TxInStatus = iota
	Failed
)
