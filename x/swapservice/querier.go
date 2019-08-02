package swapservice

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// query endpoints supported by the swapservice Querier
const (
	pathPool  = "pool"
	pathPools = "pools"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case pathPool:
			return queryPool(ctx, path[1:], req, keeper)
		case pathPools:
			return queryPools(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown swapservice query endpoint")
		}
	}
}

func queryPool(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	key := path[0]
	pool := keeper.GetPool(ctx, key)

	res, err := codec.MarshalJSONIndent(keeper.cdc, QueryPool(pool))
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}

func queryPools(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {

	var pools QueryPools

	iterator := keeper.GetPoolIterator(ctx)

	for ; iterator.Valid(); iterator.Next() {
		pool := keeper.GetPool(ctx, string(iterator.Key()))
		if !pool.Empty() {
			pools = append(pools, pool)
		}
	}

	res, err := codec.MarshalJSONIndent(keeper.cdc, pools)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}
