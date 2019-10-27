package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

// StakerUnit staker and their units in the pool
type StakerUnit struct {
	RuneAddress  common.Address `json:"rune_address"`
	AssetAddress common.Address `json:"asset_address"`
	Units        sdk.Uint       `json:"units"`
	PendingRune  sdk.Uint       `json:"pending_rune"` // number of rune coins
}

func (su StakerUnit) Valid() error {
	if su.RuneAddress.IsEmpty() {
		return errors.New("Rune address cannot be empty")
	}
	if su.AssetAddress.IsEmpty() {
		return errors.New("Asset address cannot be empty")
	}

	return nil
}

// PoolStaker
// {
//    "p_id":"pool-BNB",
//    "tu" : "1000",
//    "ss" : {
//        "bnbStaker-1":"100",,
//        "bnbStaker-2":"100"
//    }
//}
type PoolStaker struct {
	Asset      common.Asset `json:"asset"`       // asset
	TotalUnits sdk.Uint     `json:"total_units"` // total units in the pool
	Stakers    []StakerUnit `json:"stakers"`     // key will be staker id , which is the address on binane chain value will be UNITS
}

// NewPoolStaker create a new instance of PoolStaker
func NewPoolStaker(asset common.Asset, totalUnits sdk.Uint) PoolStaker {
	return PoolStaker{
		Asset:      asset,
		TotalUnits: totalUnits,
		Stakers:    []StakerUnit{},
	}
}

func (ps PoolStaker) Valid() error {
	if ps.Asset.IsEmpty() {
		return errors.New("Asset cannot be empty")
	}

	for _, staker := range ps.Stakers {
		if err := staker.Valid(); err != nil {
			return err
		}
	}

	return nil
}

// String return the human readable string of PoolStaker
func (ps PoolStaker) String() string {
	bs := strings.Builder{}
	bs.WriteString(fmt.Sprintln("asset: " + ps.Asset.String()))
	bs.WriteString(fmt.Sprintf("total units: %d", ps.TotalUnits.Uint64()))
	bs.WriteString(fmt.Sprintf("staker count: %d", len(ps.Stakers)))
	if nil != ps.Stakers {
		for _, stakerUnit := range ps.Stakers {
			bs.WriteString(fmt.Sprintln(stakerUnit.RuneAddress.String() + " : " + stakerUnit.Units.String()))
		}
	}
	return bs.String()
}

func (ps *PoolStaker) GetStakerUnit(addr common.Address) StakerUnit {
	for _, item := range ps.Stakers {
		if item.RuneAddress == addr || item.AssetAddress == addr {
			return item
		}
	}
	return StakerUnit{
		Units:       sdk.ZeroUint(),
		PendingRune: sdk.ZeroUint(),
	}
}

// RemoveStakerUnit will remove the stakerunit with given staker id from the struct
func (ps *PoolStaker) RemoveStakerUnit(runeAddr common.Address) {
	deleteIdx := -1
	for idx, item := range ps.Stakers {
		if item.RuneAddress == runeAddr {
			deleteIdx = idx
		}
	}

	if deleteIdx != -1 {
		ps.Stakers = append(ps.Stakers[:deleteIdx], ps.Stakers[deleteIdx+1:]...)
	}
}

// UpsertStakerUnit it check whether the given staker unit is exist in the struct
// if it exist then just update it , otherwise it append it
func (ps *PoolStaker) UpsertStakerUnit(stakerUnit StakerUnit) {
	pos := -1
	for idx, item := range ps.Stakers {
		if item.RuneAddress == stakerUnit.RuneAddress {
			pos = idx
		}
	}
	if pos != -1 {
		ps.Stakers[pos] = stakerUnit
		return
	}
	ps.Stakers = append(ps.Stakers, stakerUnit)
}
