package query

import (
	"fmt"
	"strings"
)

type Query struct {
	Key              string
	EndpointTemplate string
	Mainnet          bool
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

// query endpoints supported by the thorchain Querier
var (
	QueryAdminConfigBnb   = Query{Mainnet: false, Key: "adminconfig", EndpointTemplate: "/%s/admin/{%s}/{%s}"}
	QueryAdminConfig      = Query{Mainnet: false, Key: "adminconfig", EndpointTemplate: "/%s/admin/{%s}"}
	QueryPoolIndex        = Query{Mainnet: false, Key: "poolindex", EndpointTemplate: "/%s/pooltickers"}
	QueryChains           = Query{Mainnet: false, Key: "chains", EndpointTemplate: "/%s/chains"}
	QueryPool             = Query{Mainnet: false, Key: "pool", EndpointTemplate: "/%s/pool/{%s}"}
	QueryPools            = Query{Mainnet: false, Key: "pools", EndpointTemplate: "/%s/pools"}
	QueryPoolStakers      = Query{Mainnet: false, Key: "poolstakers", EndpointTemplate: "/%s/pool/{%s}/stakers"}
	QueryStakerPools      = Query{Mainnet: false, Key: "stakerpools", EndpointTemplate: "/%s/staker/{%s}"}
	QuerySwapRecord       = Query{Mainnet: false, Key: "swaprecord", EndpointTemplate: "/%s/swaprecord/{%s}"}
	QueryUnStakeRecord    = Query{Mainnet: false, Key: "unstakerecord", EndpointTemplate: "/%s/unstakerecord/{%s}"}
	QueryTxIn             = Query{Mainnet: false, Key: "txin", EndpointTemplate: "/%s/tx/{%s}"}
	QueryTxOutArray       = Query{Mainnet: true, Key: "txoutarray", EndpointTemplate: "/%s/txoutarray/{%s}"}
	QueryTxOutArrayPubkey = Query{Mainnet: true, Key: "txoutarraypubkey", EndpointTemplate: "/%s/txoutarray/{%s}/{%s}"}
	QueryIncompleteEvents = Query{Mainnet: false, Key: "incomplete_events", EndpointTemplate: ""}
	QueryCompleteEvents   = Query{Mainnet: true, Key: "complete_events", EndpointTemplate: "/%s/events/{%s}"}
	QueryHeights          = Query{Mainnet: false, Key: "heights", EndpointTemplate: "/%s/lastblock"}
	QueryChainHeights     = Query{Mainnet: false, Key: "chainheights", EndpointTemplate: "/%s/lastblock/{%s}"}
	QueryObservers        = Query{Mainnet: false, Key: "observers", EndpointTemplate: "/%s/observers"}
	QueryObserver         = Query{Mainnet: false, Key: "observer", EndpointTemplate: "/%s/observer/{%s}"}
	QueryNodeAccounts     = Query{Mainnet: false, Key: "nodeaccounts", EndpointTemplate: "/%s/nodeaccounts"}
	QueryNodeAccount      = Query{Mainnet: false, Key: "nodeaccount", EndpointTemplate: "/%s/nodeaccount/{%s}"}
	QueryPoolAddresses    = Query{Mainnet: true, Key: "pooladdresses", EndpointTemplate: "/%s/pooladdresses"}
	QueryValidators       = Query{Mainnet: false, Key: "validators", EndpointTemplate: "/%s/validators"}
)

var Queries = []Query{
	QueryAdminConfig,
	QueryAdminConfigBnb,
	QueryPool,
	QueryPools,
	QueryChains,
	QueryPoolStakers,
	QueryStakerPools,
	QueryPoolIndex,
	QuerySwapRecord,
	QueryUnStakeRecord,
	QueryTxIn,
	QueryTxOutArray,
	QueryTxOutArrayPubkey,
	QueryIncompleteEvents,
	QueryCompleteEvents,
	QueryHeights,
	QueryChainHeights,
	QueryObservers,
	QueryObserver,
	QueryNodeAccount,
	QueryNodeAccounts,
	QueryPoolAddresses,
	QueryValidators,
}
