package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "swapservice" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSetPool:
			return handleMsgSetPool(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized swapservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSetPool(ctx sdk.Context, keeper Keeper, msg MsgSetPool) sdk.Result {
	// validate there are not conflicts first
	if keeper.PoolDoesExist(ctx, msg.Pool.Key()) {
		return sdk.ErrUnknownRequest("Conflict").Result()
	}

	keeper.SetPool(ctx, msg.Pool)

	return sdk.Result{}
}

func handleMsgSetTxHash(ctx sdk.Context, keeper Keeper, msg MsgSetTxHash) sdk.Result {
	// validate there are not conflicts first
	if keeper.TxDoesExist(ctx, msg.TxHash.Key()) {
		return sdk.ErrUnknownRequest("Conflict").Result()
	}

	keeper.SetTxHash(ctx, msg.TxHash)

	return sdk.Result{}
}
