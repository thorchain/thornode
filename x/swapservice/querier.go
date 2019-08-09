package swapservice

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// query endpoints supported by the swapservice Querier
const (
	QueryAdminConfig   = "adminconfig"
	QueryPoolStruct    = "poolstruct"
	QueryPoolStructs   = "pools"
	QueryPoolStakers   = "poolstakers"
	QueryStakerPools   = "stakerpools"
	QueryPoolIndex     = "poolindex"
	QuerySwapRecord    = "swaprecord"
	QueryUnStakeRecord = "unstakerecord"
	QueryTxHash        = "txhash"
	QueryTxOutArray    = "txoutarray"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		ctx.Logger().Info("query", "path", path[0])
		switch path[0] {
		case QueryPoolStruct:
			return queryPoolStruct(ctx, path[1:], req, keeper)
		case QueryPoolStructs:
			return queryPoolStructs(ctx, req, keeper)
		case QueryPoolStakers:
			return queryPoolStakers(ctx, path[1:], req, keeper)
		case QueryStakerPools:
			return queryStakerPool(ctx, path[1:], req, keeper)
		case QueryPoolIndex:
			return queryPoolIndex(ctx, path[1:], req, keeper)
		case QuerySwapRecord:
			return querySwapRecord(ctx, path[1:], req, keeper)
		case QueryUnStakeRecord:
			return queryUnStakeRecord(ctx, path[1:], req, keeper)
		case QueryTxHash:
			return queryTxHash(ctx, path[1:], req, keeper)
		case QueryAdminConfig:
			return queryAdminConfig(ctx, path[1:], req, keeper)
		case QueryTxOutArray:
			return queryTxOutArray(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown swapservice query endpoint")
		}
	}
}

// queryUnStakeRecord
func queryUnStakeRecord(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	sr, err := keeper.GetUnStakeRecord(ctx, path[0])
	if nil != err {
		ctx.Logger().Error("fail to get UnStake record", err)
		return nil, sdk.ErrInternal("fail to get UnStake record")
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, sr)
	if nil != err {
		ctx.Logger().Error("fail to marshal UnStake record to json", err)
		return nil, sdk.ErrInternal("fail to marshal UnStake record to json")
	}
	return res, nil
}

// querySwapRecord
func querySwapRecord(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	sr, err := keeper.GetSwapRecord(ctx, path[0])
	if nil != err {
		ctx.Logger().Error("fail to get swaprecord", err)
		return nil, sdk.ErrInternal("fail to get swap record")
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, sr)
	if nil != err {
		ctx.Logger().Error("fail to marshal swap record to json", err)
		return nil, sdk.ErrInternal("fail to marshal swap record to json")
	}
	return res, nil
}

// queryPoolIndex
func queryPoolIndex(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	ps, err := keeper.GetPoolIndex(ctx)
	if nil != err {
		ctx.Logger().Error("fail to get pool index", err)
		return nil, sdk.ErrInternal("fail to get pool index")
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, ps)
	if nil != err {
		ctx.Logger().Error("fail to marshal pool index to json", err)
		return nil, sdk.ErrInternal("fail to marshal pool index to json")
	}
	return res, nil
}

// queryPoolStakers
func queryPoolStakers(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	ticker := path[0]
	ps, err := keeper.GetPoolStaker(ctx, ticker)
	if nil != err {
		ctx.Logger().Error("fail to get pool staker", err)
		return nil, sdk.ErrInternal("fail to get pool staker")
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, ps)
	if nil != err {
		ctx.Logger().Error("fail to marshal pool staker to json", err)
		return nil, sdk.ErrInternal("fail to marshal pool staker to json")
	}
	return res, nil
}

// queryStakerPool
func queryStakerPool(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	addr := path[0]
	ps, err := keeper.GetStakerPool(ctx, addr)
	if nil != err {
		ctx.Logger().Error("fail to get staker pool", err)
		return nil, sdk.ErrInternal("fail to get staker pool")
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, ps)
	if nil != err {
		ctx.Logger().Error("fail to marshal staker pool to json", err)
		return nil, sdk.ErrInternal("fail to marshal staker pool to json")
	}
	return res, nil
}

// nolint: unparam
func queryPoolStruct(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	poolstruct := keeper.GetPoolStruct(ctx, path[0])
	if poolstruct.Empty() {
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("pool: %s doesn't exist", path[0]))
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, poolstruct)
	if err != nil {
		return nil, sdk.ErrInternal("could not marshal result to JSON")
	}
	return res, nil
}

func queryPoolStructs(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var pools QueryResPoolStructs
	iterator := keeper.GetPoolStructDataIterator(ctx)
	for ; iterator.Valid(); iterator.Next() {
		var poolstruct PoolStruct
		keeper.cdc.MustUnmarshalBinaryBare(iterator.Value(), &poolstruct)
		pools = append(pools, poolstruct)
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, pools)
	if err != nil {
		return nil, sdk.ErrInternal("could not marshal pools result to json")
	}
	return res, nil
}

func queryTxHash(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	hash := path[0]
	tx := keeper.GetTxHash(ctx, hash)
	res, err := codec.MarshalJSONIndent(keeper.cdc, tx)
	if nil != err {
		ctx.Logger().Error("fail to marshal tx hash to json", err)
		return nil, sdk.ErrInternal("fail to marshal tx hash to json")
	}
	return res, nil
}

func queryTxOutArray(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	height, err := strconv.ParseInt(path[0], 0, 64)
	if nil != err {
		ctx.Logger().Error("fail to parse block height", err)
		return nil, sdk.ErrInternal("fail to parse block height")
	}
	tx, err := keeper.GetTxOut(ctx, height)
	if nil != err {
		ctx.Logger().Error("fail to get tx out array from key value store", err)
		return nil, sdk.ErrInternal("fail to get tx out array from key value store")
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, tx)
	if nil != err {
		ctx.Logger().Error("fail to marshal tx hash to json", err)
		return nil, sdk.ErrInternal("fail to marshal tx hash to json")
	}
	return res, nil
}

func queryAdminConfig(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	key := path[0]
	config := keeper.GetAdminConfig(ctx, key)
	res, err := codec.MarshalJSONIndent(keeper.cdc, config)
	if nil != err {
		ctx.Logger().Error("fail to marshal config to json", err)
		return nil, sdk.ErrInternal("fail to marshal config to json")
	}
	return res, nil
}
