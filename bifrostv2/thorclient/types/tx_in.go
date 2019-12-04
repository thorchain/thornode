package types

import "gitlab.com/thorchain/thornode/common"

type TxIn struct {
	BlockHeight string       `json:"blockHeight"`
	Count       string       `json:"count"`
	Chain       common.Chain `json:"chain"`
	TxArray     []TxInItem   `json:"txArray"`
}

type TxInItem struct {
	Tx                  string       `json:"tx"`
	Memo                string       `json:"memo"`
	Sender              string       `json:"sender"`
	To                  string       `json:"to"` // to adddress
	Coins               common.Coins `json:"coins"`
	Gas                 common.Gas   `json:"gas"`
	ObservedPoolAddress string       `json:"observed_pool_address"`
}
type TxInStatus byte

const (
	Processing TxInStatus = iota
	Failed
)
