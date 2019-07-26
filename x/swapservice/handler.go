package swapservice

import (
	"fmt"
	"log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "swapservice" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSetPoolData:
			return handleMsgSetPoolData(ctx, keeper, msg)
		case MsgSetAccData:
			return handleMsgSetAccData(ctx, keeper, msg)
		case MsgSetStakeData:
			return handleMsgSetStakeData(ctx, keeper, msg)
		case MsgSwap:
			return handleMsgSwap(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized swapservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle a message to set pooldata
func handleMsgSetPoolData(ctx sdk.Context, keeper Keeper, msg MsgSetPoolData) sdk.Result {
	if err := msg.ValidateBasic(); nil != err {
		return sdk.Result{
			Code: sdk.CodeUnknownRequest,
			Data: []byte(err.Error()),
		}
	}
	keeper.SetPoolData(
		ctx,
		msg.PoolID,
		msg.TokenName,
		msg.Ticker,
		msg.BalanceRune,
		msg.BalanceToken,
		msg.PoolAddress,
		msg.Status)
	// TODO return appropriate messages to client
	return sdk.Result{}
}

// Handle a message to set acc data
func handleMsgSetAccData(ctx sdk.Context, keeper Keeper, msg MsgSetAccData) sdk.Result {
	// TODO: Validate the message
	keeper.SetAccData(
		ctx,
		msg.AccID,
		msg.Name,
		msg.Ticker,
		msg.Amount,
	) // If so, set the acc data to the value specified in the msg.
	return sdk.Result{} // return
}

// Handle a message to set stake data
func handleMsgSetStakeData(ctx sdk.Context, keeper Keeper, msg MsgSetStakeData) sdk.Result {
	// TODO: Validate the message
	fmt.Println()
	log.Printf("Setting stake: %s", msg.Name)
	err := stake(
		ctx,
		keeper,
		msg.Name,
		msg.Ticker,
		msg.Rune,
		msg.Token,
	)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	return sdk.Result{}
}

// Handle a message to set stake data
func handleMsgSwap(ctx sdk.Context, keeper Keeper, msg MsgSwap) sdk.Result {
	// TODO: Validate the message
	err := swap(
		ctx,
		keeper,
		msg.SourceTicker,
		msg.TargetTicker,
		msg.Amount,
		msg.Requester,
		msg.Destination,
	) // If so, set the stake data to the value specified in the msg.
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	return sdk.Result{} // return
}
