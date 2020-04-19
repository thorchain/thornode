package types

import "gitlab.com/thorchain/thornode/common"

type ErrataBlock struct {
	Height int64
	Txs    []ErrataTx
}

type ErrataTx struct {
	Height int64
	TxID   common.TxID
	Chain  common.Chain
}
