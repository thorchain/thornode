package query

import (
	"fmt"
	"strings"
)

type Query struct {
	Key              string
	EndpointTemplate string
}

func (q Query) Endpoint(args ...string) string {
	count := strings.Count(q.EndpointTemplate, "%s")
	a := args[:count]

	in := make([]interface{}, len(a))
	for i, _ := range in {
		in[i] = a[i]
	}

	return fmt.Sprintf(q.EndpointTemplate, in...)
}

func (q Query) Path(args ...string) string {
	path := fmt.Sprintf("custom/%s/%s", args[0], q.Key)
	if len(args) > 1 {
		path = fmt.Sprintf("%s/%s", path, args[1])
	}
	return path
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
