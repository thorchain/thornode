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

// IsEmpty check whether the block gas is empty
func (b *BlockGas) IsEmpty() bool {
	return b.GasReimburse.IsEmpty() && b.GasSpend.IsEmpty() && b.GasTopup.IsEmpty()
}

// AddGas to the block gas structure so
func (b *BlockGas) AddGas(gas common.Gas, ty GasType) {
	switch ty {
	case GasTypeTopup:
		b.GasTopup = b.GasTopup.Add(gas)
	case GasTypeReimburse:
		b.GasReimburse = b.GasReimburse.Add(gas)
	case GasTypeSpend:
		b.GasSpend = b.GasSpend.Add(gas)
	}
}
