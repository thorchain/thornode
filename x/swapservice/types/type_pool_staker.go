package types

import (
	"fmt"
	"strings"
)

// StakerUnit staker and their units in the pool
type StakerUnit struct {
	StakerID string `json:"staker_id"`
	Units    Amount `json:"units"`
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
	Ticker     Ticker       `json:"ticker"`      // ticker
	TotalUnits Amount       `json:"total_units"` // total units in the pool
	Stakers    []StakerUnit `json:"stakers"`     // key will be staker id , which is the address on binane chain value will be UNITS
}

// NewPoolStaker create a new instance of PoolStaker
func NewPoolStaker(ticker Ticker, totalUnits Amount) PoolStaker {
	return PoolStaker{
		Ticker:     ticker,
		TotalUnits: totalUnits,
		Stakers:    []StakerUnit{},
	}
}

// String return the human readable string of PoolStaker
func (ps PoolStaker) String() string {
	bs := strings.Builder{}
	bs.WriteString(fmt.Sprintln("ticker: " + ps.Ticker.String()))
	bs.WriteString(fmt.Sprintln("total units: " + ps.TotalUnits))
	if nil != ps.Stakers {
		for _, stakerUnit := range ps.Stakers {
			bs.WriteString(fmt.Sprintln(stakerUnit.StakerID + " : " + stakerUnit.Units.String()))
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
		Units:    ZeroAmount,
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

// UpsertStakerUnit it check whether the given staker unit is exist in the struct
// if it exist then just update it , otherwise it append it
func (ps *PoolStaker) UpsertStakerUnit(stakerUnit StakerUnit) {
	pos := -1
	for idx, item := range ps.Stakers {
		if item.StakerID == stakerUnit.StakerID {
			pos = idx
		}
	}
	if pos != -1 {
		ps.Stakers[pos] = stakerUnit
		return
	}
	ps.Stakers = append(ps.Stakers, stakerUnit)
}
