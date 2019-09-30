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
	QueryAdminConfigBnb   = Query{Key: "adminconfig", EndpointTemplate: "/%s/admin/{%s}/{%s}"}
	QueryAdminConfig      = Query{Key: "adminconfig", EndpointTemplate: "/%s/admin/{%s}"}
	QueryPoolIndex        = Query{Key: "poolindex", EndpointTemplate: "/%s/pooltickers"}
	QueryPool             = Query{Key: "pool", EndpointTemplate: "/%s/pool/{%s}"}
	QueryPools            = Query{Key: "pools", EndpointTemplate: "/%s/pools"}
	QueryPoolStakers      = Query{Key: "poolstakers", EndpointTemplate: "/%s/pool/{%s}/stakers"}
	QueryStakerPools      = Query{Key: "stakerpools", EndpointTemplate: "/%s/staker/{%s}"}
	QuerySwapRecord       = Query{Key: "swaprecord", EndpointTemplate: "/%s/swaprecord/{%s}"}
	QueryUnStakeRecord    = Query{Key: "unstakerecord", EndpointTemplate: "/%s/unstakerecord/{%s}"}
	QueryTxIn             = Query{Key: "txin", EndpointTemplate: "/%s/tx/{%s}"}
	QueryTxOutArray       = Query{Key: "txoutarray", EndpointTemplate: "/%s/txoutarray/{%s}"}
	QueryIncompleteEvents = Query{Key: "incomplete_events", EndpointTemplate: ""}
	QueryCompleteEvents   = Query{Key: "complete_events", EndpointTemplate: "/%s/events/{%s}"}
	QueryHeights          = Query{Key: "heights", EndpointTemplate: "/%s/lastblock"}
	QueryObservers        = Query{Key: "observers", EndpointTemplate: "/%s/observers"}
	QueryNodeAccounts     = Query{Key: "nodeaccounts", EndpointTemplate: "/%s/nodeaccounts"}
	QueryNodeAccount      = Query{Key: "nodeaccount", EndpointTemplate: "/%s/nodeaccount/{%s}"}
)

var Queries = []Query{
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
	QueryHeights,
	QueryObservers,
	QueryNodeAccount,
	QueryNodeAccounts,
}
