package types

import (
	"fmt"

	"gitlab.com/thorchain/thornode/common"
)

// Block is our simplified/generic block
type Block struct {
	Txs         Txs
	Chain       common.Chain
	BlockHeight int64
	BlockHash   string
}

func (b *Block) String() string {
	return fmt.Sprintf("%v: %v: %v ", b.Chain, b.BlockHeight, b.BlockHash)
}

type Txs []TxItem

type TxItem struct {
	From  common.Address
	To    common.Address
	Coins common.Coins
	Gas   common.Gas
	Memo  string
}
