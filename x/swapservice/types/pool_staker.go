package types

import (
	"fmt"
	"strings"
)

type StakerUnit struct {
	StakerID string `json:"staker_id"`
	Units    string `json:"units"`
}

// PoolStakerKeyPrefix all poolstaker key start with this
const PoolStakerKeyPrefix = "poolstaker-"

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
	PoolID     string       `json:"pool_id"`     // pool id
	TotalUnits string       `json:"total_units"` // total units in the pool
	Stakers    []StakerUnit `json:"stakers"`     // key will be staker id , which is the address on binane chain value will be UNITS
}

// NewPoolStaker create a new instance of PoolStaker
func NewPoolStaker(poolID string, totalUnits string) PoolStaker {
	return PoolStaker{
		PoolID:     poolID,
		TotalUnits: totalUnits,
		Stakers:    []StakerUnit{},
	}
}

// String return the human readable string of PoolStaker
func (ps PoolStaker) String() string {
	bs := strings.Builder{}
	bs.WriteString(fmt.Sprintln("pool-id: " + ps.PoolID))
	bs.WriteString(fmt.Sprintln("total units: " + ps.TotalUnits))
	if nil != ps.Stakers {
		for _, stakerUnit := range ps.Stakers {
			bs.WriteString(fmt.Sprintln(stakerUnit.StakerID + " : " + stakerUnit.Units))
		}
	}
	return bs.String()
}
func (ps *PoolStaker) GetStakerUnit(stakerID string) StakerUnit {
	for _, item := range ps.Stakers {
		if item.StakerID == stakerID {
			return item
		}
	}
	return StakerUnit{
		StakerID: stakerID,
		Units:    "0",
	}
}

// RemoveStakerUnit will remove the stakerunit with given staker id from the struct
func (ps *PoolStaker) RemoveStakerUnit(stakerID string) {
	deleteIdx := -1
	for idx, item := range ps.Stakers {
		if item.StakerID == stakerID {
			deleteIdx = idx
		}
	}

	if deleteIdx != -1 {
		ps.Stakers = append(ps.Stakers[:deleteIdx], ps.Stakers[deleteIdx+1:]...)
	}
}

// UpsertStakeUnit
func (ps *PoolStaker) UpsertStakerUnit(stakerUnit StakerUnit) {
	deleteIdx := -1
	for idx, item := range ps.Stakers {
		if item.StakerID == stakerUnit.StakerID {
			deleteIdx = idx
		}
	}
	if deleteIdx != -1 {
		ps.Stakers[deleteIdx] = stakerUnit
		return
	}
	ps.Stakers = append(ps.Stakers, stakerUnit)
}
