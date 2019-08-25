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
	temp := []string{args[0], q.Key}
	args = append(temp, args[1:]...)
	return fmt.Sprintf("custom/%s", strings.Join(args, "/"))
}

// query endpoints supported by the swapservice Querier
var (
	QueryAdminConfigBnb   Query = Query{"adminconfig", "/%s/admin/{%s}/{%s}"}
	QueryAdminConfig      Query = Query{"adminconfig", "/%s/admin/{%s}"}
	QueryPoolIndex        Query = Query{"poolindex", "/%s/pooltickers"}
	QueryPool             Query = Query{"pool", "/%s/pool/{%s}"}
	QueryPools            Query = Query{"pools", "/%s/pools"}
	QueryPoolStakers      Query = Query{"poolstakers", "/%s/pool/{%s}/stakers"}
	QueryStakerPools      Query = Query{"stakerpools", "/%s/staker/{%s}"}
	QuerySwapRecord       Query = Query{"swaprecord", "/%s/swaprecord/{%s}"}
	QueryUnStakeRecord    Query = Query{"unstakerecord", "/%s/unstakerecord/{%s}"}
	QueryTxIn             Query = Query{"txin", "/%s/tx/{%s}"}
	QueryTxOutArray       Query = Query{"txoutarray", "/%s/txoutarray/{%s}"}
	QueryIncompleteEvents Query = Query{"incomplete_events", ""}
	QueryCompleteEvents   Query = Query{"complete_events", "/%s/complete/{%s}"}
)

var Queries []Query = []Query{
	QueryAdminConfig,
	QueryAdminConfigBnb,
	QueryPool,
	QueryPools,
	QueryPoolStakers,
	QueryStakerPools,
	QueryPoolIndex,
	QuerySwapRecord,
	QueryUnStakeRecord,
	QueryTxIn,
	QueryTxOutArray,
	QueryIncompleteEvents,
	QueryCompleteEvents,
}
