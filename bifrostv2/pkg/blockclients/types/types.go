package types

import "gitlab.com/thorchain/thornode/common"

type ObservedBlockAndTxs struct {
	Txs         Txs
	Chain       common.Chain
	BlockHeight uint64
	BlockHash   string
}

type Tx struct {
	Hash        string
	FromAddress common.Address
	ToAddress   common.Address
	Coins       common.Coins
	Gas         common.Gas
	Memo        string
}

type Txs []Tx
