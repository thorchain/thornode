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
	for i := range in {
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
	QueryPool               = Query{Key: "pool", EndpointTemplate: "/%s/pool/{%s}"}
	QueryPools              = Query{Key: "pools", EndpointTemplate: "/%s/pools"}
	QueryStakers            = Query{Key: "stakers", EndpointTemplate: "/%s/pool/{%s}/stakers"}
	QueryTxIn               = Query{Key: "txin", EndpointTemplate: "/%s/tx/{%s}"}
	QueryKeysignArray       = Query{Key: "keysign", EndpointTemplate: "/%s/keysign/{%s}"}
	QueryKeysignArrayPubkey = Query{Key: "keysignpubkey", EndpointTemplate: "/%s/keysign/{%s}/{%s}"}
	QueryKeygensPubkey      = Query{Key: "keygenspubkey", EndpointTemplate: "/%s/keygen/{%s}/{%s}"}
	QueryCompEvents         = Query{Key: "comp_events", EndpointTemplate: "/%s/events/{%s}"}
	QueryCompEventsByChain  = Query{Key: "comp_events_chain", EndpointTemplate: "/%s/events/{%s}/{%s}"}
	QueryEventsByTxHash     = Query{Key: "txhash_events", EndpointTemplate: "/%s/events/tx/{%s}"}
	QueryHeights            = Query{Key: "heights", EndpointTemplate: "/%s/lastblock"}
	QueryChainHeights       = Query{Key: "chainheights", EndpointTemplate: "/%s/lastblock/{%s}"}
	QueryObservers          = Query{Key: "observers", EndpointTemplate: "/%s/observers"}
	QueryObserver           = Query{Key: "observer", EndpointTemplate: "/%s/observer/{%s}"}
	QueryNodeAccounts       = Query{Key: "nodeaccounts", EndpointTemplate: "/%s/nodeaccounts"}
	QueryNodeAccount        = Query{Key: "nodeaccount", EndpointTemplate: "/%s/nodeaccount/{%s}"}
	QueryPoolAddresses      = Query{Key: "pooladdresses", EndpointTemplate: "/%s/pool_addresses"}
	QueryVaultData          = Query{Key: "vaultdata", EndpointTemplate: "/%s/vault"}
	QueryVaultsAsgard       = Query{Key: "vaultsasgard", EndpointTemplate: "/%s/vaults/asgard"}
	QueryVaultsYggdrasil    = Query{Key: "vaultsyggdrasil", EndpointTemplate: "/%s/vaults/yggdrasil"}
	QueryVaultPubkeys       = Query{Key: "vaultpubkeys", EndpointTemplate: "/%s/vaults/pubkeys"}
	QueryTSSSigners         = Query{Key: "tsssigner", EndpointTemplate: "/%s/vaults/{%s}/signers"}
	QueryConstantValues     = Query{Key: "constants", EndpointTemplate: "/%s/constants"}
	QueryMimirValues        = Query{Key: "mimirs", EndpointTemplate: "/%s/mimir"}
	QueryBan                = Query{Key: "ban", EndpointTemplate: "/%s/ban/{%s}"}
)

// Queries all queries
var Queries = []Query{
	QueryPool,
	QueryPools,
	QueryStakers,
	QueryTxIn,
	QueryKeysignArray,
	QueryKeysignArrayPubkey,
	QueryEventsByTxHash,
	QueryCompEvents,
	QueryCompEventsByChain,
	QueryHeights,
	QueryChainHeights,
	QueryObservers,
	QueryObserver,
	QueryNodeAccount,
	QueryNodeAccounts,
	QueryPoolAddresses,
	QueryVaultData,
	QueryVaultsAsgard,
	QueryVaultsYggdrasil,
	QueryVaultPubkeys,
	QueryKeygensPubkey,
	QueryTSSSigners,
	QueryConstantValues,
	QueryMimirValues,
	QueryBan,
}
