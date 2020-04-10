package thorchain

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperLiquidityFees interface {
	AddToLiquidityFees(ctx sdk.Context, asset common.Asset, fee sdk.Uint) error
	GetTotalLiquidityFeesInBlock(ctx sdk.Context, height uint64) (sdk.Uint, error)
	GetPoolLiquidityFeesInBlock(ctx sdk.Context, height uint64, asset common.Asset) (sdk.Uint, error)
	GetTotalLiquidityFees(ctx sdk.Context) (sdk.Uint, error)
	GetPoolLiquidityFees(ctx sdk.Context, asset common.Asset) (sdk.Uint, error)
}

// AddToLiquidityFees - measure of fees collected in each block
func (k KVStore) AddToLiquidityFees(ctx sdk.Context, asset common.Asset, fee sdk.Uint) error {
	store := ctx.KVStore(k.storeKey)
	currentHeight := uint64(ctx.BlockHeight())

	totalFeesInBlock, err := k.GetTotalLiquidityFeesInBlock(ctx, currentHeight)
	if err != nil {
		return err
	}
	poolFeesInBlock, err := k.GetPoolLiquidityFeesInBlock(ctx, currentHeight, asset)
	if err != nil {
		return err
	}
	totalFees, err := k.GetTotalLiquidityFees(ctx)
	if err != nil {
		return err
	}
	poolFees, err := k.GetPoolLiquidityFees(ctx, asset)
	if err != nil {
		return err
	}

	totalFeesInBlock = totalFeesInBlock.Add(fee)
	poolFeesInBlock = poolFeesInBlock.Add(fee)
	totalFees = totalFees.Add(fee)
	poolFees = poolFees.Add(fee)

	// update total liquidity fee in block
	key := k.GetKey(ctx, prefixTotalLiquidityFeeInBlock, strconv.FormatUint(currentHeight, 10))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(totalFeesInBlock))

	// update pool liquidity fee in block
	key = k.GetKey(ctx, prefixPoolLiquidityFeeInBlock, fmt.Sprintf("%d-%s", currentHeight, asset.String()))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(poolFeesInBlock))

	// update total liquidity fee
	key = k.GetKey(ctx, prefixTotalLiquidityFee, "")
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(totalFees))

	// update pool liquidity fee
	key = k.GetKey(ctx, prefixPoolLiquidityFee, fmt.Sprintf("%s", asset.String()))
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

	if err := k.cdc.UnmarshalBinaryBare(buf, &liquidityFees); err != nil {
		return sdk.ZeroUint(), dbError(ctx, "Unmarshal: liquidity fees", err)
	}
	return liquidityFees, nil
}

// GetTotalLiquidityFeesInBlock - total of all fees collected in each block
func (k KVStore) GetTotalLiquidityFeesInBlock(ctx sdk.Context, height uint64) (sdk.Uint, error) {
	key := k.GetKey(ctx, prefixTotalLiquidityFeeInBlock, strconv.FormatUint(height, 10))
	return k.getLiquidityFees(ctx, key)
}

// GetPoolLiquidityFees - total of fees collected in each block per pool
func (k KVStore) GetPoolLiquidityFeesInBlock(ctx sdk.Context, height uint64, asset common.Asset) (sdk.Uint, error) {
	key := k.GetKey(ctx, prefixPoolLiquidityFeeInBlock, fmt.Sprintf("%d-%s", height, asset.String()))
	return k.getLiquidityFees(ctx, key)
}

// GetTotalLiquidityFees - total of all fees collected ever
func (k KVStore) GetTotalLiquidityFees(ctx sdk.Context) (sdk.Uint, error) {
	key := k.GetKey(ctx, prefixTotalLiquidityFee, "")
	return k.getLiquidityFees(ctx, key)
}

// GetPoolLiquidityFees - total of fees collected in each per pool ever
func (k KVStore) GetPoolLiquidityFees(ctx sdk.Context, asset common.Asset) (sdk.Uint, error) {
	key := k.GetKey(ctx, prefixPoolLiquidityFee, fmt.Sprintf("%s", asset.String()))
	return k.getLiquidityFees(ctx, key)
}
