package types

import (
	"fmt"
	"strings"
)

// StakeTxDetail represent all the stake activity
// Staker can stake on the same pool for multiple times
type StakeTxDetail struct {
	RequestTxHash string `json:"request_tx_hash"` // the tx hash from binance chain , represent staker send token to the pool
	RuneAmount    string `json:"rune_amount"`     // amount of rune that send in at the time
	TokenAmount   string `json:"token_amount"`    // amount of token that send in at the time
}

// StakerPoolItem represent the staker's activity in a pool
// Staker can stake on multiple pool
type StakerPoolItem struct {
	Ticker       Ticker          `json:"ticker"`
	Units        string          `json:"units"`
	StakeDetails []StakeTxDetail `json:"stake_details"`
}

// StakerPool represent staker and their activities in the pools
// A staker can stake on multiple pool.
// A Staker can stake on the same pool for multiple times
// json representation will be looking like
// {
//    "staker_id": "bnbxasdfaswqerqwe",
//    "pool_and_units": [
//        {
//            "ticker": "BNB",
//            "units": "200",
//            "stake_details": [
//                {
//                    "request_tx_hash": "txhash from binance chain",
//                    "rune_amount": "100",
//                    "token_amount": "100"
//                },
//                {
//                    "request_tx_hash": "another hash",
//                    "rune_amount": "100",
//                    "token_amount": "100"
//                }
//            ]
//        },
//        {
//            "ticker": "BTC",
//            "units": "200",
//            "stake_details": [
//                {
//                    "request_tx_hash": "txhash from binance chain",
//                    "rune_amount": "100",
//                    "token_amount": "100"
//                },
//                {
//                    "request_tx_hash": "another hash",
//                    "rune_amount": "100",
//                    "token_amount": "100"
//                }
//            ]
//        }
//    ]
// }
type StakerPool struct {
	StakerID  string            `json:"staker_id"`      // this will be staker's address on binance chain
	PoolUnits []*StakerPoolItem `json:"pool_and_units"` // the key of this map will be the pool id , value will bt [UNIT,RUNE,TOKEN]
}

// NewStakerPool create a new instance of StakerPool
func NewStakerPool(id string) StakerPool {
	return StakerPool{
		StakerID:  id,
		PoolUnits: []*StakerPoolItem{},
	}
}

// String return a user readable string representation of Staker Pool
func (sp StakerPool) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintln("staker-id: " + sp.StakerID))
	if nil != sp.PoolUnits {
		for _, item := range sp.PoolUnits {
			sb.WriteString(fmt.Sprintf("pool-id: %s, staker unitsL %s", item.Units, item.Units))
		}
	}
	return sb.String()
}

// GetStakerPoolItem return the StakerPoolItem with given pool id
func (sp *StakerPool) GetStakerPoolItem(ticker Ticker) *StakerPoolItem {
	for _, item := range sp.PoolUnits {
		if ticker.Equals(item.Ticker) {
			return item
		}
	}
	return &StakerPoolItem{
		Ticker:       ticker,
		Units:        "0",
		StakeDetails: []StakeTxDetail{},
	}
}

// RemoveStakerPoolItem delete the stakerpoolitem with given pool id from the struct
func (sp *StakerPool) RemoveStakerPoolItem(ticker Ticker) {
	deleteIdx := -1
	for idx, item := range sp.PoolUnits {
		if item.Ticker.Equals(ticker) {
			deleteIdx = idx
		}
	}
	if deleteIdx != -1 {
		if deleteIdx == 0 {
			sp.PoolUnits = []*StakerPoolItem{}
		} else {
			sp.PoolUnits = append(sp.PoolUnits[:deleteIdx], sp.PoolUnits[deleteIdx:]...)
		}
	}
}

// UpsertStakerPoolItem if the given stakerPoolItem exist then it will update , otherwise it will add
func (sp *StakerPool) UpsertStakerPoolItem(stakerPoolItem *StakerPoolItem) {
	pos := -1
	for idx, item := range sp.PoolUnits {
		if item.Ticker.Equals(stakerPoolItem.Ticker) {
			pos = idx
		}
	}
	if pos != -1 {
		sp.PoolUnits[pos] = stakerPoolItem
		return
	}
	sp.PoolUnits = append(sp.PoolUnits, stakerPoolItem)
}

// AddStakerTxDetail to the StakerPool structure
func (spi *StakerPoolItem) AddStakerTxDetail(requestTxHash, runeAmount, tokenAmount string) {
	std := StakeTxDetail{
		RequestTxHash: requestTxHash,
		RuneAmount:    runeAmount,
		TokenAmount:   tokenAmount,
	}
	spi.StakeDetails = append(spi.StakeDetails, std)
}
