package types

import (
	"fmt"
	"strings"
)

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
	PoolID     string            `json:"p_id"` // pool id
	TotalUnits string            `json:"tu"`   // total units in the pool
	Stakers    map[string]string `json:"ss"`   // key will be staker id , which is the address on binane chain value will be UNITS
}

// NewPoolStaker create a new instance of PoolStaker
func NewPoolStaker(poolID string, totalUnits string) PoolStaker {
	return PoolStaker{
		PoolID:     poolID,
		TotalUnits: totalUnits,
		Stakers:    make(map[string]string),
	}
}

// String return the human readable string of PoolStaker
func (ps PoolStaker) String() string {
	bs := strings.Builder{}
	bs.WriteString(fmt.Sprintln("pool-id: " + ps.PoolID))
	bs.WriteString(fmt.Sprintln("total units: " + ps.TotalUnits))
	if nil != ps.Stakers {
		for key, unit := range ps.Stakers {
			bs.WriteString(fmt.Sprintln(key + " : " + unit))
		}
	}
	return bs.String()
}
