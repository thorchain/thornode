package thorchain

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// KeeperBlockGas method to save BlockGas to internal
type KeeperBlockGas interface {
	SaveBlockGas(ctx sdk.Context, blockGas BlockGas) error
	GetBlockGas(ctx sdk.Context) (BlockGas, error)
	RemoveBlockGas(ctx sdk.Context)
}

// SaveBlockGas save the given block gas to kv store
func (k KVStore) SaveBlockGas(ctx sdk.Context, blockGas BlockGas) error {
	key := k.GetKey(ctx, prefixBlockGas, strconv.FormatInt(ctx.BlockHeight(), 10))
	store := ctx.KVStore(k.storeKey)
	buf, err := k.Cdc().MarshalBinaryBare(blockGas)
	if err != nil {
		return err
	}
	store.Set([]byte(key), buf)
	return nil
}

// GetBlockGas get the block gas related to current block height
func (k KVStore) GetBlockGas(ctx sdk.Context) (BlockGas, error) {
	key := k.GetKey(ctx, prefixBlockGas, strconv.FormatInt(ctx.BlockHeight(), 10))
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return NewBlockGas(ctx.BlockHeight()), nil
	}
	buf := store.Get([]byte(key))
	var g BlockGas
	if err := k.Cdc().UnmarshalBinaryBare(buf, &g); err != nil {
		return BlockGas{}, fmt.Errorf("fail to unmarshal event: %w", err)
	}
	return g, nil
}

// RemoveBlockGas remove block gas from the kv store
func (k KVStore) RemoveBlockGas(ctx sdk.Context) {
	key := k.GetKey(ctx, prefixBlockGas, strconv.FormatInt(ctx.BlockHeight(), 10))
	store := ctx.KVStore(k.storeKey)
	store.Delete([]byte(key))
}
