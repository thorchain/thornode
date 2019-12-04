package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperValidatorMeta interface {
	SetValidatorMeta(ctx sdk.Context, meta ValidatorMeta) error
	GetValidatorMeta(ctx sdk.Context) (ValidatorMeta, error)
}

// SetValidatorMeta save the ValidatorMeta information into kv store
func (k KVStore) SetValidatorMeta(ctx sdk.Context, meta ValidatorMeta) error {
	key := k.GetKey(ctx, prefixValidatorMeta, "")
	store := ctx.KVStore(k.storeKey)
	buf, err := k.cdc.MarshalBinaryBare(meta)
	if nil != err {
		return dbError(ctx, "fail to marshal validator meta to binary", err)
	}
	store.Set([]byte(key), buf)
	return nil
}

// GetValidatorMeta get ValidatorMeta information from kv store
func (k KVStore) GetValidatorMeta(ctx sdk.Context) (ValidatorMeta, error) {
	var meta ValidatorMeta
	key := k.GetKey(ctx, prefixValidatorMeta, "")
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		if err := k.cdc.UnmarshalBinaryBare(buf, &meta); nil != err {
			return meta, fmt.Errorf("fail to unmarshal validator meta: %w", err)
		}
	}
	return meta, nil
}
