package thorchain

// This file is intended to do orchestration for emitting an event

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

func eventPoolStatusWrapper(ctx sdk.Context, keeper Keeper, pool Pool) error {
	poolEvt := NewEventPool(pool.Asset, pool.Status)
	bytes, err := json.Marshal(poolEvt)
	if err != nil {
		return fmt.Errorf("fail to marshal pool event: %w", err)
	}

	tx := common.Tx{ID: common.BlankTxID}
	evt := NewEvent(poolEvt.Type(), ctx.BlockHeight(), tx, bytes, EventSuccess)
	return keeper.UpsertEvent(ctx, evt)
}
