package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

type KeeperChains interface {
	GetChains(ctx sdk.Context) (common.Chains, error)
	SetChains(ctx sdk.Context, chains common.Chains)
}

func (k KVStore) GetChains(ctx sdk.Context) (common.Chains, error) {
	chains := make(common.Chains, 0)
	key := k.GetKey(ctx, prefixSupportedChains, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return chains, nil
	}
	buf := store.Get([]byte(key))
	err := k.cdc.UnmarshalBinaryBare(buf, &chains)
	if err != nil {
		return chains, dbError(ctx, "Unmarshal: chains", err)
	}
	return chains, nil
}

func (k KVStore) SetChains(ctx sdk.Context, chains common.Chains) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixSupportedChains, "")
	chains = chains.Distinct()
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(chains))
}
