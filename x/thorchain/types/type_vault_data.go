package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

type VaultData struct {
	BondRewardRune sdk.Uint   `json:"bond_reward_rune"` // The total amount of awarded rune for bonders
	TotalBondUnits sdk.Uint   `json:"total_bond_units"` // Total amount of bond units
	TotalReserve   sdk.Uint   `json:"total_reserve"`    // Total amount of reserves (in rune)
	Gas            common.Gas `json:"gas"`              // Total gas used (intended to be tracked per block and be repaid via block rewards)
}

func NewVaultData() VaultData {
	return VaultData{
		BondRewardRune: sdk.ZeroUint(),
		TotalBondUnits: sdk.ZeroUint(),
		TotalReserve:   sdk.ZeroUint(),
	}
}
