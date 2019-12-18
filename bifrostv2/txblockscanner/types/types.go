package types

import (
	"gitlab.com/thorchain/thornode/common"
)

// Block is our simplified/generic block
type Block struct {
	Txs         Txs
	Chain       common.Chain
	BlockHeight uint64
	BlockHash   string
}

type Txs []TxItem

type TxItem struct {
	From  common.Address
	To    common.Address
	Coins common.Coins
	Gas   common.Gas
	Memo  string
}
