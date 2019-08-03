package swapservice

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

// query endpoints supported by the swapservice Querier
const (
	QueryPoolStruct  = "poolstruct"
	QueryPoolDatas   = "pooldatas"
	QueryPoolStakers = "poolstakers"
	QueryStakerPools = "stakerpools"
	QueryPoolIndex   = "poolindex"
	QuerySwapRecord  = "swaprecord"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryPoolStruct:
			return queryPoolStruct(ctx, path[1:], req, keeper)
		case QueryPoolDatas:
			return queryPoolDatas(ctx, req, keeper)
		case QueryPoolStakers:
			return queryPoolStakers(ctx, path[1:], req, keeper)
		case QueryStakerPools:
			return queryStakerPool(ctx, path[1:], req, keeper)
		case QueryPoolIndex:
			return queryPoolIndex(ctx, path[1:], req, keeper)
		case QuerySwapRecord:
			return querySwapRecord(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown swapservice query endpoint")
		}
	}
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
	poolID := types.GetPoolNameFromTicker(ticker)
	ps, err := keeper.GetPoolStaker(ctx, poolID)
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
	if len(poolstruct.PoolID) == 0 {
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("pool: %s doesn't exist", path[0]))
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, poolstruct)
	if err != nil {
		return nil, sdk.ErrInternal("could not marshal result to JSON")
	}
	return res, nil
}

func queryPoolDatas(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var pooldatasList QueryResPoolDatas
	iterator := keeper.GetPoolStructDataIterator(ctx)
	for ; iterator.Valid(); iterator.Next() {
		var poolstruct PoolStruct
		keeper.cdc.MustUnmarshalBinaryBare(iterator.Value(), &poolstruct)
		pooldatasList = append(pooldatasList, poolstruct)
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, pooldatasList)
	if err != nil {
		return nil, sdk.ErrInternal("could not marshal pools result to json")
	}
	return res, nil
}
