package query

import (
	"fmt"
	"strings"
)

type Query struct {
	Key      string
	Endpoint string
}

func (q Query) Sprintf(args ...string) string {
	count := strings.Count(q.Endpoint, "%s")
	return fmt.Sprintf(q.Endpoint, args[:count])
}

// query endpoints supported by the swapservice Querier
var (
	QueryAdminConfig   Query = Query{"adminconfig", "/%s/admin/{%s}"}
	QueryPoolIndex     Query = Query{"poolindex", "/%s/pooltickers"}
	QueryPoolStruct    Query = Query{"pool", "/%s/pool/{%s}"}
	QueryPoolStructs   Query = Query{"pools", "/%s/pools"}
	QueryPoolStakers   Query = Query{"poolstakers", "/%s/pool/{%s}/stakers"}
	QueryStakerPools   Query = Query{"stakerpools", "/%s/staker/{%s}"}
	QuerySwapRecord    Query = Query{"swaprecord", "/%s/swaprecord/{%s}"}
	QueryUnStakeRecord Query = Query{"unstakerecord", "/%s/unstakerecord/{%s}"}
	QueryTxHash        Query = Query{"txhash", "/%s/tx/{%s}"}
	QueryTxOutArray    Query = Query{"txoutarray", "/%s/txoutarray/{%s}"}
)

var Queries []Query = []Query{
	QueryAdminConfig,
	QueryPoolStruct,
	QueryPoolStructs,
	QueryPoolStakers,
	QueryStakerPools,
	QueryPoolIndex,
	QuerySwapRecord,
	QueryUnStakeRecord,
	QueryTxHash,
	QueryTxOutArray,
}
