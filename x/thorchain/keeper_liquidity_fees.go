package thorchain

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

type KeeperLiquidityFees interface {
	AddToLiquidityFees(ctx sdk.Context, pool Pool, fee sdk.Uint) error
	getLiquidityFees(ctx sdk.Context, height uint64, prefix dbPrefix) (sdk.Uint, error)
	GetTotalLiquidityFees(ctx sdk.Context, height uint64) (sdk.Uint, error)
	GetPoolLiquidityFees(ctx sdk.Context, height uint64, pool Pool) (sdk.Uint, error)
}

// AddToLiquidityFees - measure of fees collected in each block
func (k KVStore) AddToLiquidityFees(ctx sdk.Context, pool Pool, fee sdk.Uint) error {
	store := ctx.KVStore(k.storeKey)
	currentHeight := uint64(ctx.BlockHeight())

	totalFees, err := k.GetTotalLiquidityFees(ctx, currentHeight)
	if err != nil {
		return err
	}
	poolFees, err := k.GetPoolLiquidityFees(ctx, currentHeight, pool)
	if err != nil {
		return err
	}

	totalFees = totalFees.Add(fee)
	poolFees = poolFees.Add(fee)
	key := getKey(prefixTotalLiquidityFee, strconv.FormatUint(currentHeight, 10), getVersion(k.GetLowestActiveVersion(ctx), prefixTotalLiquidityFee))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(totalFees))
	strHeightPool := fmt.Sprintf("%s%s", strconv.FormatUint(currentHeight, 10), pool.Asset.String())
	key2 := getKey(prefixPoolLiquidityFee, strHeightPool, getVersion(k.GetLowestActiveVersion(ctx), prefixPoolLiquidityFee))
	store.Set([]byte(key2), k.cdc.MustMarshalBinaryBare(poolFees))
	return nil
}

func (k KVStore) getLiquidityFees(ctx sdk.Context, height uint64, prefix dbPrefix) (sdk.Uint, error) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefix, strconv.FormatUint(height, 10), getVersion(k.GetLowestActiveVersion(ctx), prefix))
	if !store.Has([]byte(key)) {
		return sdk.ZeroUint(), nil
	}
	buf := store.Get([]byte(key))
	var liquidityFees sdk.Uint
	if err := k.cdc.UnmarshalBinaryBare(buf, &liquidityFees); nil != err {
		return sdk.ZeroUint(), errors.Wrap(err, "fail to unmarshal liquidityFees")
	}
	return liquidityFees, nil
}

// GetTotalLiquidityFees - total of all fees collected in each block
func (k KVStore) GetTotalLiquidityFees(ctx sdk.Context, height uint64) (sdk.Uint, error) {
	return k.getLiquidityFees(ctx, height, prefixTotalLiquidityFee)
}

// GetPoolLiquidityFees - total of fees collected in each block per pool
func (k KVStore) GetPoolLiquidityFees(ctx sdk.Context, height uint64, pool Pool) (sdk.Uint, error) {
	return k.getLiquidityFees(ctx, height, prefixPoolLiquidityFee)
}
