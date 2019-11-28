package thorchain

import sdk "github.com/cosmos/cosmos-sdk/types"

type KeeperValidatorMeta interface {
	SetValidatorMeta(ctx sdk.Context, meta ValidatorMeta)
	GetValidatorMeta(ctx sdk.Context) ValidatorMeta
}

func (k KVStore) SetValidatorMeta(ctx sdk.Context, meta ValidatorMeta) {
	key := k.GetKey(ctx, prefixValidatorMeta, "")
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(meta))
}

func (k KVStore) GetValidatorMeta(ctx sdk.Context) ValidatorMeta {
	var meta ValidatorMeta
	key := k.GetKey(ctx, prefixValidatorMeta, "")
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &meta)
	}
	return meta
}
