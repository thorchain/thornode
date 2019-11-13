package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type VaultData struct {
	BondRewardRune sdk.Uint `json:"bond_reward_rune"` // The total amount of awarded rune for bonders
	BondUnits      sdk.Uint `json:"bond_units"`       // Total amount of bond units
	TotalReserve   sdk.Uint `json:"total_reserve"`    // Total amount of reserves (in rune)
}
