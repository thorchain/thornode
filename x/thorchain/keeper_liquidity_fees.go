package thorchain

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperLiquidityFees interface {
	AddToLiquidityFees(ctx sdk.Context, asset common.Asset, fee sdk.Uint) error
	GetTotalLiquidityFees(ctx sdk.Context, height uint64) (sdk.Uint, error)
	GetPoolLiquidityFees(ctx sdk.Context, height uint64, asset common.Asset) (sdk.Uint, error)
}

// AddToLiquidityFees - measure of fees collected in each block
func (k KVStore) AddToLiquidityFees(ctx sdk.Context, asset common.Asset, fee sdk.Uint) error {
	store := ctx.KVStore(k.storeKey)
	currentHeight := uint64(ctx.BlockHeight())

	totalFees, err := k.GetTotalLiquidityFees(ctx, currentHeight)
	if err != nil {
		return err
	}
	poolFees, err := k.GetPoolLiquidityFees(ctx, currentHeight, asset)
	if err != nil {
		return err
	}

	totalFees = totalFees.Add(fee)
	poolFees = poolFees.Add(fee)

	// update total liquidity
	key := k.GetKey(ctx, prefixTotalLiquidityFee, strconv.FormatUint(currentHeight, 10))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(totalFees))

	// update pool liquidity
	key = k.GetKey(ctx, prefixPoolLiquidityFee, fmt.Sprintf("%d-%s", currentHeight, asset.String()))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(poolFees))
	return nil
}

func (k KVStore) getLiquidityFees(ctx sdk.Context, key string) (sdk.Uint, error) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return sdk.ZeroUint(), nil
	}
	buf := store.Get([]byte(key))
	var liquidityFees sdk.Uint

	if err := k.cdc.UnmarshalBinaryBare(buf, &liquidityFees); nil != err {
		return sdk.ZeroUint(), dbError(ctx, "Unmarshal: liquidity fees", err)
	}
	return liquidityFees, nil
}

// GetTotalLiquidityFees - total of all fees collected in each block
func (k KVStore) GetTotalLiquidityFees(ctx sdk.Context, height uint64) (sdk.Uint, error) {
	key := k.GetKey(ctx, prefixTotalLiquidityFee, strconv.FormatUint(height, 10))
	return k.getLiquidityFees(ctx, key)
}

// GetPoolLiquidityFees - total of fees collected in each block per pool
func (k KVStore) GetPoolLiquidityFees(ctx sdk.Context, height uint64, asset common.Asset) (sdk.Uint, error) {
	key := k.GetKey(ctx, prefixPoolLiquidityFee, fmt.Sprintf("%d-%s", height, asset.String()))
	return k.getLiquidityFees(ctx, key)
}
