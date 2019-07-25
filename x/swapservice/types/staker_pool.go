package types

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
