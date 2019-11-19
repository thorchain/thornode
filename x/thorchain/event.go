package thorchain

// This file is intended to do orchestration for emitting an event

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func eventPoolStatusWrapper(ctx sdk.Context, keeper poolStorage, pool Pool) {
	poolEvt := NewEventPool(pool.Asset, pool.Status)
	bytes, err := json.Marshal(poolEvt)
	if err != nil {
		ctx.Logger().Error("failed to marshal pool event", err)
	}
	evt := Event{
		Height: ctx.BlockHeight(),
		Event:  bytes,
		Status: EventSuccess,
	}

	keeper.SetCompletedEvent(ctx, evt)
}
