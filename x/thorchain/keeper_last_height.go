package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/common"
)

type KeeperLastHeight interface {
	SetLastSignedHeight(ctx sdk.Context, height sdk.Uint)
	GetLastSignedHeight(ctx sdk.Context) (sdk.Uint, error)
	SetLastChainHeight(ctx sdk.Context, chain common.Chain, height sdk.Uint) error
	GetLastChainHeight(ctx sdk.Context, chain common.Chain) (sdk.Uint, error)
}

func (k KVStore) SetLastSignedHeight(ctx sdk.Context, height sdk.Uint) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixLastSignedHeight, "")
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(height))
}

func (k KVStore) GetLastSignedHeight(ctx sdk.Context) (sdk.Uint, error) {
	var height sdk.Uint
	key := k.GetKey(ctx, prefixLastSignedHeight, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return sdk.ZeroUint(), nil
	}
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &height); nil != err {
		return sdk.ZeroUint(), dbError(ctx, "Unmarshal: last heights", err)
	}
	return height, nil
}

func (k KVStore) SetLastChainHeight(ctx sdk.Context, chain common.Chain, height sdk.Uint) error {
	lastHeight, err := k.GetLastChainHeight(ctx, chain)
	if err != nil {
		return err
	}
	if lastHeight.GT(height) {
		err := errors.Errorf("current block height :%s is larger than %s , block height can't go backward ", lastHeight, height)
		return dbError(ctx, "", err)
	}
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixLastChainHeight, "")
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(height))
	return nil
}

func (k KVStore) GetLastChainHeight(ctx sdk.Context, chain common.Chain) (sdk.Uint, error) {
	var height sdk.Uint
	key := k.GetKey(ctx, prefixLastChainHeight, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return sdk.ZeroUint(), nil
	}
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &height); nil != err {
		return height, dbError(ctx, "Unmarshal: last heights", err)
	}
	return height, nil
}
