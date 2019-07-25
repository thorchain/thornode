package types

import (
	"fmt"
	"strings"
)

// StakePool represent staker and their activities in the pools
// {
//	"id":"bnbStakerAddr1",
//	"pu":{
//		"pool-BNB":["100", "100", "100"],
//		"pool-XXX":["200", "200", "200"]
//	}
// }
type StakerPool struct {
	StakerID  string               `json:"id"` // this will be staker's address on binance chain
	PoolUnits map[string][3]string `json:"pu"` // the key of this map will be the pool id , value will bt [UNIT,RUNE,TOKEN]
}

// NewStakerPool create a new instance of StakerPool
func NewStakerPool(id string) StakerPool {
	return StakerPool{
		StakerID:  id,
		PoolUnits: make(map[string][3]string),
	}
}

// String return a user readable string representation of Staker Pool
func (sp StakerPool) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintln("staker-id: " + sp.StakerID))
	if nil != sp.PoolUnits {
		for key, item := range sp.PoolUnits {
			sb.WriteString(fmt.Sprintf("%s - %s ", key, strings.Join(item[:], ",")))
		}
	}
	return sb.String()
}
