package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperRagnarok interface {
	GetRagnarokBlockHeight(_ sdk.Context) (sdk.Uint, error)
	SetRagnarokBlockHeight(_ sdk.Context, _ sdk.Uint)
}

func (k KVStore) GetRagnarokBlockHeight(ctx sdk.Context) (sdk.Uint, error) {
	key := k.GetKey(ctx, prefixRagnarok, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return sdk.ZeroUint(), nil
	}
	var ragnarok sdk.Uint
	buf := store.Get([]byte(key))
	err := k.cdc.UnmarshalBinaryBare(buf, &ragnarok)
	if err != nil {
		return sdk.ZeroUint(), dbError(ctx, "Unmarshal: ragnarok", err)
	}
	return ragnarok, nil
}

func (k KVStore) SetRagnarokBlockHeight(ctx sdk.Context, height sdk.Uint) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixRagnarok, "")
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(height))
}
