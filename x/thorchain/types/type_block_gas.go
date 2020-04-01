package types

import "gitlab.com/thorchain/thornode/common"

// BlockGas is to record all the gas happened in a block
type BlockGas struct {
	Height       int64      `json:"height"`
	GasSpend     common.Gas `json:"gas_spend"`
	GasReimburse common.Gas `json:"gas_reimburse"`
	GasTopup     common.Gas `json:"gas_topup"`
}

// NewBlockGas create a new instance of BlockGas
func NewBlockGas(height int64) BlockGas {
	return BlockGas{
		Height: height,
	}
}
