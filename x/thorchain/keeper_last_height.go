package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperLastHeight interface {
	SetLastSignedHeight(ctx sdk.Context, height int64)
	GetLastSignedHeight(ctx sdk.Context) (int64, error)
	SetLastChainHeight(ctx sdk.Context, chain common.Chain, height int64) error
	GetLastChainHeight(ctx sdk.Context, chain common.Chain) (int64, error)
}

func (k KVStore) SetLastSignedHeight(ctx sdk.Context, height int64) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixLastSignedHeight, "")
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(height))
}

func (k KVStore) GetLastSignedHeight(ctx sdk.Context) (int64, error) {
	var height int64
	key := k.GetKey(ctx, prefixLastSignedHeight, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return 0, nil
	}
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &height); nil != err {
		return 0, dbError(ctx, "Unmarshal: last heights", err)
	}
	return height, nil
}

func (k KVStore) SetLastChainHeight(ctx sdk.Context, chain common.Chain, height int64) error {
	lastHeight, err := k.GetLastChainHeight(ctx, chain)
	if err != nil {
		return err
	}
	if lastHeight > height {
		err := fmt.Errorf("last block height :%d is larger than %d , block height can't go backward ", lastHeight, height)
		return dbError(ctx, "", err)
	}
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixLastChainHeight, "")
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(height))
	return nil
}

func (k KVStore) GetLastChainHeight(ctx sdk.Context, chain common.Chain) (int64, error) {
	var height int64
	key := k.GetKey(ctx, prefixLastChainHeight, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return 0, nil
	}
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &height); nil != err {
		return height, dbError(ctx, "Unmarshal: last heights", err)
	}
	return height, nil
}
