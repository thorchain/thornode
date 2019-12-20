package query

import (
	"fmt"
	"strings"
)

// Query define all the queries
type Query struct {
	Key              string
	EndpointTemplate string
}

// Endpoint return the end point string
func (q Query) Endpoint(args ...string) string {
	count := strings.Count(q.EndpointTemplate, "%s")
	a := args[:count]

	in := make([]interface{}, len(a))
	for i, _ := range in {
		in[i] = a[i]
	}

	return fmt.Sprintf(q.EndpointTemplate, in...)
}

// Path return the path
func (q Query) Path(args ...string) string {
	temp := []string{args[0], q.Key}
	args = append(temp, args[1:]...)
	return fmt.Sprintf("custom/%s", strings.Join(args, "/"))
}

// query endpoints supported by the thorchain Querier
var (
	QueryAdminConfigBnb     = Query{Key: "adminconfig", EndpointTemplate: "/%s/admin/{%s}/{%s}"}
	QueryAdminConfig        = Query{Key: "adminconfig", EndpointTemplate: "/%s/admin/{%s}"}
	QueryChains             = Query{Key: "chains", EndpointTemplate: "/%s/chains"}
	QueryPool               = Query{Key: "pool", EndpointTemplate: "/%s/pool/{%s}"}
	QueryPools              = Query{Key: "pools", EndpointTemplate: "/%s/pools"}
	QueryPoolStakers        = Query{Key: "poolstakers", EndpointTemplate: "/%s/pool/{%s}/stakers"}
	QueryStakerPools        = Query{Key: "stakerpools", EndpointTemplate: "/%s/staker/{%s}"}
	QuerySwapRecord         = Query{Key: "swaprecord", EndpointTemplate: "/%s/swaprecord/{%s}"}
	QueryUnStakeRecord      = Query{Key: "unstakerecord", EndpointTemplate: "/%s/unstakerecord/{%s}"}
	QueryTxIn               = Query{Key: "txin", EndpointTemplate: "/%s/tx/{%s}"}
	QueryKeysignArray       = Query{Key: "keysign", EndpointTemplate: "/%s/keysign/{%s}"}
	QueryKeysignArrayPubkey = Query{Key: "keysignpubkey", EndpointTemplate: "/%s/keysign/{%s}/{%s}"}
	QueryKeygens            = Query{Key: "keygens", EndpointTemplate: "/%s/keygen/{%s}"}
	QueryKeygensPubkey      = Query{Key: "keygenspubkey", EndpointTemplate: "/%s/keygen/{%s}/{%s}"}
	QueryCompleteEvents     = Query{Key: "complete_events", EndpointTemplate: "/%s/events/{%s}"}
	QueryHeights            = Query{Key: "heights", EndpointTemplate: "/%s/lastblock"}
	QueryChainHeights       = Query{Key: "chainheights", EndpointTemplate: "/%s/lastblock/{%s}"}
	QueryObservers          = Query{Key: "observers", EndpointTemplate: "/%s/observers"}
	QueryObserver           = Query{Key: "observer", EndpointTemplate: "/%s/observer/{%s}"}
	QueryNodeAccounts       = Query{Key: "nodeaccounts", EndpointTemplate: "/%s/nodeaccounts"}
	QueryNodeAccount        = Query{Key: "nodeaccount", EndpointTemplate: "/%s/nodeaccount/{%s}"}
	QueryPoolAddresses      = Query{Key: "pooladdresses", EndpointTemplate: "/%s/pool_addresses"}
	QueryVaultData          = Query{Key: "vaultdata", EndpointTemplate: "/%s/vault"}
	QueryVaultPubkeys       = Query{Key: "pubkeys", EndpointTemplate: "/%s/vaults/pubkeys"}
)

// Queries all queries
var Queries = []Query{
	QueryAdminConfig,
	QueryAdminConfigBnb,
	QueryPool,
	QueryPools,
	QueryChains,
	QueryPoolStakers,
	QueryStakerPools,
	QuerySwapRecord,
	QueryUnStakeRecord,
	QueryTxIn,
	QueryKeysignArray,
	QueryKeysignArrayPubkey,
	QueryCompleteEvents,
	QueryHeights,
	QueryChainHeights,
	QueryObservers,
	QueryObserver,
	QueryNodeAccount,
	QueryNodeAccounts,
	QueryPoolAddresses,
	QueryVaultData,
	QueryVaultPubkeys,
	QueryKeygens,
	QueryKeygensPubkey,
}
