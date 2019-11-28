package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

type KeeperChains interface {
	GetChains(ctx sdk.Context) common.Chains
	SupportedChain(ctx sdk.Context, chain common.Chain) bool
	AddChain(ctx sdk.Context, chain common.Chain)
}

func (k KVStore) GetChains(ctx sdk.Context) common.Chains {
	chains := make(common.Chains, 0)
	key := getKey(prefixSupportedChains, "", getVersion(k.GetLowestActiveVersion(ctx), prefixSupportedChains))
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &chains)
	}
	return chains
}

func (k KVStore) SupportedChain(ctx sdk.Context, chain common.Chain) bool {
	for _, ch := range k.GetChains(ctx) {
		if ch.Equals(chain) {
			return true
		}
	}
	return false
}

func (k KVStore) AddChain(ctx sdk.Context, chain common.Chain) {
	key := getKey(prefixSupportedChains, "", getVersion(k.GetLowestActiveVersion(ctx), prefixSupportedChains))
	if k.SupportedChain(ctx, chain) {
		// already added
		return
	}
	chains := k.GetChains(ctx)
	chains = append(chains, chain)
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(chains))
}
