package types

// PoolStaker
// {
//    "p_id":"pool-BNB",
//    "tu" : "1000",
//    "ss" : {
//        "bnbStaker-1":["100","100","100"],
//        "bnbStaker-2":["100","100","100"]
//    }
//}
type PoolStaker struct {
	PoolID     string               `json:"p_id"` // pool id
	TotalUnits string               `json:"tu"`   // total units in the pool
	Stakers    map[string][3]string `json:"ss"`   // key will be staker id , which is the address on binane chain value will be [UNITS, RUNE, TOKEN]
}
