package swapservice

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"gitlab.com/thorchain/bepswap/common"
	q "gitlab.com/thorchain/bepswap/statechain/x/swapservice/query"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		ctx.Logger().Info("query", "path", path[0])
		switch path[0] {
		case q.QueryPool.Key:
			return queryPool(ctx, path[1:], req, keeper)
		case q.QueryPools.Key:
			return queryPools(ctx, req, keeper)
		case q.QueryPoolStakers.Key:
			return queryPoolStakers(ctx, path[1:], req, keeper)
		case q.QueryStakerPools.Key:
			return queryStakerPool(ctx, path[1:], req, keeper)
		case q.QueryPoolIndex.Key:
			return queryPoolIndex(ctx, path[1:], req, keeper)
		case q.QueryTxIn.Key:
			return queryTxIn(ctx, path[1:], req, keeper)
		case q.QueryAdminConfig.Key, q.QueryAdminConfigBnb.Key:
			return queryAdminConfig(ctx, path[1:], req, keeper)
		case q.QueryTxOutArray.Key:
			return queryTxOutArray(ctx, path[1:], req, keeper)
		case q.QueryIncompleteEvents.Key:
			return queryInCompleteEvents(ctx, path[1:], req, keeper)
		case q.QueryCompleteEvents.Key:
			return queryCompleteEvents(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest(
				fmt.Sprintf("unknown swapservice query endpoint: %s", path[0]),
			)
		}
	}
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
	ticker, err := common.NewTicker(path[0])
	if nil != err {
		ctx.Logger().Error("fail to get parse ticker", err)
		return nil, sdk.ErrInternal("fail to parse ticker")
	}
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
	addr, err := common.NewBnbAddress(path[0])
	if nil != err {
		ctx.Logger().Error("fail to parse bnb address", err)
		return nil, sdk.ErrInternal("fail to parse bnb address")
	}

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
func queryPool(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	ticker, err := common.NewTicker(path[0])
	if err != nil {
		ctx.Logger().Error("fail to parse ticker", err)
		return nil, sdk.ErrInternal("Could not parse ticker")
	}
	pool := keeper.GetPool(ctx, ticker)
	if pool.Empty() {
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("pool: %s doesn't exist", path[0]))
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, pool)
	if err != nil {
		return nil, sdk.ErrInternal("could not marshal result to JSON")
	}
	return res, nil
}

func queryPools(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var pools QueryResPools
	iterator := keeper.GetPoolDataIterator(ctx)
	for ; iterator.Valid(); iterator.Next() {
		var pool Pool
		keeper.cdc.MustUnmarshalBinaryBare(iterator.Value(), &pool)
		pools = append(pools, pool)
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, pools)
	if err != nil {
		return nil, sdk.ErrInternal("could not marshal pools result to json")
	}
	return res, nil
}

func queryTxIn(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	hash, err := common.NewTxID(path[0])
	if err != nil {
		ctx.Logger().Error("fail to parse tx id", err)
		return nil, sdk.ErrInternal("fail to parse tx id")
	}
	voter := keeper.GetTxInVoter(ctx, hash)
	res, err := codec.MarshalJSONIndent(keeper.cdc, voter.GetTx(4))
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
	var err error
	key := GetAdminConfigKey(path[0])
	bnb := common.NoBnbAddress
	if len(path) > 1 {
		bnb, err = common.NewBnbAddress(path[1])
		if err != nil {
			ctx.Logger().Error("fail to parse bnb address", err)
			return nil, sdk.ErrInternal("fail to parse bnb address")
		}
	}
	config := NewAdminConfig(key, "", bnb)
	config.Value, err = keeper.GetAdminConfigValue(ctx, key, bnb)
	if nil != err {
		ctx.Logger().Error("fail to get admin config", err)
		return nil, sdk.ErrInternal("fail to get admin config")
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, config)
	if nil != err {
		ctx.Logger().Error("fail to marshal config to json", err)
		return nil, sdk.ErrInternal("fail to marshal config to json")
	}
	return res, nil
}

func queryInCompleteEvents(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	events, err := keeper.GetIncompleteEvents(ctx)
	if nil != err {
		ctx.Logger().Error("fail to get incomplete events", err)
		return nil, sdk.ErrInternal("fail to get incomplete events")
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, events)
	if nil != err {
		ctx.Logger().Error("fail to marshal events to json", err)
		return nil, sdk.ErrInternal("fail to marshal events to json")
	}
	return res, nil
}

func queryCompleteEvents(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	id, err := strconv.ParseInt(path[0], 10, 64)
	if err != nil {
		ctx.Logger().Error("fail to discover id number", err)
		return nil, sdk.ErrInternal("fail to discover id number")
	}

	limit := int64(100) // limit the number of events, aka pagination
	events := make(Events, 0)
	for i := id; i <= id+limit; i++ {
		event, _ := keeper.GetCompletedEvent(ctx, i)
		if !event.Empty() {
			events = append(events, event)
		} else {
			break
		}
	}

	res, err := codec.MarshalJSONIndent(keeper.cdc, events)
	if nil != err {
		ctx.Logger().Error("fail to marshal events to json", err)
		return nil, sdk.ErrInternal("fail to marshal events to json")
	}
	return res, nil
}
