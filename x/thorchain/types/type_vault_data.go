package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// VaultData
type VaultData struct {
	BondRewardRune sdk.Uint `json:"bond_reward_rune"` // The total amount of awarded rune for bonders
	TotalBondUnits sdk.Uint `json:"total_bond_units"` // Total amount of bond units
	TotalReserve   sdk.Uint `json:"total_reserve"`    // Total amount of reserves (in rune)
	TotalBEP2Rune  sdk.Uint `json:"total_bep2_rune"`  // Total amount of BEP2 rune held
}

// NewVaultData create a new instance VaultData it is empty though
func NewVaultData() VaultData {
	return VaultData{
		BondRewardRune: sdk.ZeroUint(),
		TotalBondUnits: sdk.ZeroUint(),
		TotalReserve:   sdk.ZeroUint(),
		TotalBEP2Rune:  sdk.ZeroUint(),
	}
}

// calculate node rewards
func (v VaultData) CalcNodeRewards(nodeUnits sdk.Uint) sdk.Uint {
	return common.GetShare(nodeUnits, v.TotalBondUnits, v.BondRewardRune)
}
