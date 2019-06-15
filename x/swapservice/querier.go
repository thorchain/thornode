package swapservice

import (
	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// query endpoints supported by the swapservice Querier
const (
	QueryPoolStruct  = "poolstruct"
	QueryPoolDatas   = "pooldatas"
	QueryAccStruct   = "accstruct"
	QueryAccDatas    = "accdatas"
	QueryStakeStruct = "stakestruct"
	QueryStakeDatas  = "stakedatas"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryPoolStruct:
			return queryPoolStruct(ctx, path[1:], req, keeper)
		case QueryPoolDatas:
			return queryPoolDatas(ctx, req, keeper)
		case QueryAccStruct:
			return queryAccStruct(ctx, path[1:], req, keeper)
		case QueryAccDatas:
			return queryAccDatas(ctx, req, keeper)
		case QueryStakeStruct:
			return queryStakeStruct(ctx, path[1:], req, keeper)
		case QueryStakeDatas:
			return queryStakeDatas(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown swapservice query endpoint")
		}
	}
}

// nolint: unparam
func queryPoolStruct(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	poolstruct := keeper.GetPoolStruct(ctx, path[0])

	res, err := codec.MarshalJSONIndent(keeper.cdc, poolstruct)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}

// nolint: unparam
func queryAccStruct(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	accstruct := keeper.GetAccStruct(ctx, path[0])

	res, err := codec.MarshalJSONIndent(keeper.cdc, accstruct)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}

// nolint: unparam
func queryStakeStruct(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	stakestruct := keeper.GetStakeStruct(ctx, path[0])

	res, err := codec.MarshalJSONIndent(keeper.cdc, stakestruct)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}

func queryPoolDatas(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var pooldatasList QueryResPoolDatas

	iterator := keeper.GetDatasIterator(ctx)

	for ; iterator.Valid(); iterator.Next() {
		pooldatasList = append(pooldatasList, string(iterator.Key()))
	}

	res, err := codec.MarshalJSONIndent(keeper.cdc, pooldatasList)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}

func queryAccDatas(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var accdatasList QueryResAccDatas

	iterator := keeper.GetDatasIterator(ctx)

	for ; iterator.Valid(); iterator.Next() {
		accdatasList = append(accdatasList, string(iterator.Key()))
	}

	res, err := codec.MarshalJSONIndent(keeper.cdc, accdatasList)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}

func queryStakeDatas(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var accdatasList QueryResStakeDatas

	iterator := keeper.GetDatasIterator(ctx)

	for ; iterator.Valid(); iterator.Next() {
		accdatasList = append(accdatasList, string(iterator.Key()))
	}

	res, err := codec.MarshalJSONIndent(keeper.cdc, accdatasList)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}
