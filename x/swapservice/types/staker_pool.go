package types

import (
	"fmt"
	"strings"
)

const StakerPoolKeyPrefix = `stakerpool-`

type StakerPoolItem struct {
	PoolID       string `json:"pool_id"`
	Units        string `json:"units"`
	RuneBalance  string `json:"rune_balance"`
	TokenBalance string `json:"token_balance"`
}

// StakePool represent staker and their activities in the pools
// {
//	"id":"bnbStakerAddr1",
//	"pu":{
//		"pool-BNB":["100", "100", "100"],
//		"pool-XXX":["200", "200", "200"]
//	}
// }
type StakerPool struct {
	StakerID  string           `json:"staker_id"`      // this will be staker's address on binance chain
	PoolUnits []StakerPoolItem `json:"pool_and_units"` // the key of this map will be the pool id , value will bt [UNIT,RUNE,TOKEN]
}

// NewStakerPool create a new instance of StakerPool
func NewStakerPool(id string) StakerPool {
	return StakerPool{
		StakerID:  id,
		PoolUnits: []StakerPoolItem{},
	}
}

// String return a user readable string representation of Staker Pool
func (sp StakerPool) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintln("staker-id: " + sp.StakerID))
	if nil != sp.PoolUnits {
		for _, item := range sp.PoolUnits {
			sb.WriteString(fmt.Sprintf("pool-id: %s, rune:%s ,token:%s ", item.Units, item.RuneBalance, item.TokenBalance))
		}
	}
	return sb.String()
}

func (sp *StakerPool) GetStakerPoolItem(poolID string) StakerPoolItem {
	for _, item := range sp.PoolUnits {
		if item.PoolID == poolID {
			return item
		}
	}
	return StakerPoolItem{
		PoolID:       poolID,
		Units:        "0",
		TokenBalance: "0",
		RuneBalance:  "0",
	}
}

// UpsertStakerPoolItem
func (sp *StakerPool) UpsertStakerPoolItem(stakerPoolItem StakerPoolItem) {
	deleteIdx := -1
	for idx, item := range sp.PoolUnits {
		if item.PoolID == stakerPoolItem.PoolID {
			deleteIdx = idx
		}
	}
	if deleteIdx != -1 {
		sp.PoolUnits[deleteIdx] = stakerPoolItem
		return
	}
	sp.PoolUnits = append(sp.PoolUnits, stakerPoolItem)
}
