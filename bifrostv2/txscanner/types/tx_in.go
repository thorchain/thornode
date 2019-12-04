package types

import "gitlab.com/thorchain/thornode/common"

type TxIn struct {
	BlockHeight uint64
	BlockHash   string
	Chain       common.Chain
}
