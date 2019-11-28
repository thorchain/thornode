package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/common"
)

type KeeperLastHeight interface {
	SetLastSignedHeight(ctx sdk.Context, height sdk.Uint)
	GetLastSignedHeight(ctx sdk.Context) (height sdk.Uint)
	SetLastChainHeight(ctx sdk.Context, chain common.Chain, height sdk.Uint) error
	GetLastChainHeight(ctx sdk.Context, chain common.Chain) (height sdk.Uint)
}

func (k KVStore) SetLastSignedHeight(ctx sdk.Context, height sdk.Uint) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixLastSignedHeight, "")
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(height))
}

func (k KVStore) GetLastSignedHeight(ctx sdk.Context) (height sdk.Uint) {
	key := k.GetKey(ctx, prefixLastSignedHeight, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return sdk.ZeroUint()
	}
	bz := store.Get([]byte(key))
	k.cdc.MustUnmarshalBinaryBare(bz, &height)
	return
}

func (k KVStore) SetLastChainHeight(ctx sdk.Context, chain common.Chain, height sdk.Uint) error {
	currentHeight := k.GetLastChainHeight(ctx, chain)
	if currentHeight.GT(height) {
		return errors.Errorf("current block height :%s is larger than %s , block height can't go backward ", currentHeight, height)
	}
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixLastChainHeight, "")
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(height))
	return nil
}

func (k KVStore) GetLastChainHeight(ctx sdk.Context, chain common.Chain) (height sdk.Uint) {
	key := k.GetKey(ctx, prefixLastChainHeight, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return sdk.ZeroUint()
	}
	bz := store.Get([]byte(key))
	k.cdc.MustUnmarshalBinaryBare(bz, &height)
	return
}
