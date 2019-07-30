package swapservice

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

// NewHandler returns a handler for "swapservice" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSetPoolData:
			return handleMsgSetPoolData(ctx, keeper, msg)
		case MsgSetStakeData:
			return handleMsgSetStakeData(ctx, keeper, msg)
		case MsgSwap:
			return handleMsgSwap(ctx, keeper, msg)
		case types.MsgSetUnStake:
			return handleMsgSetUnstake(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized swapservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle a message to set pooldata
func handleMsgSetPoolData(ctx sdk.Context, keeper Keeper, msg MsgSetPoolData) sdk.Result {
	ctx.Logger().Info("handleMsgSetPoolData request", "poolID:"+msg.PoolID)
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
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
	return sdk.Result{
		Code: sdk.CodeOK,
	}
}

// Handle a message to set stake data
func handleMsgSetStakeData(ctx sdk.Context, keeper Keeper, msg MsgSetStakeData) sdk.Result {
	ctx.Logger().Info("handleMsgSetStakeData request", "stakerid:"+msg.Ticker)
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.Result{
			Code: sdk.CodeUnknownRequest,
			Data: []byte(err.Error()),
		}
	}
	if err := stake(
		ctx,
		keeper,
		msg.Name,
		msg.Ticker,
		msg.Rune,
		msg.Token,
		msg.PublicAddress); err != nil {
		ctx.Logger().Error("fail to process stake message", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	return sdk.Result{
		Code: sdk.CodeOK,
	}
}

// Handle a message to set stake data
func handleMsgSwap(ctx sdk.Context, keeper Keeper, msg MsgSwap) sdk.Result {
	amount, err := swap(
		ctx,
		keeper,
		msg.SourceTicker,
		msg.TargetTicker,
		msg.Amount,
		msg.Requester,
		msg.Destination,
	) // If so, set the stake data to the value specified in the msg.
	if err != nil {
		ctx.Logger().Error("fail to process swap message", err)
		return sdk.ErrInternal(err.Error()).Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      []byte(amount),
		Codespace: "swap",
	}
}

// handleMsgSetUnstake process unstake
func handleMsgSetUnstake(ctx sdk.Context, keeper Keeper, msg types.MsgSetUnStake) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgSetUnstake from : %s(%s) unstake (%s)", msg, msg.PublicAddress, msg.Percentage))
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgSetUnstake", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	runeAmt, tokenAmount, err := unstake(ctx, keeper, msg)
	if nil != err {
		ctx.Logger().Error("fail to unstake", "error", err)
		return sdk.ErrInternal("fail to process unstake request").Result()
	}
	res, err := json.Marshal(struct {
		Rune  string `json:"rune"`
		Token string `json:"token"`
	}{
		Rune:  runeAmt,
		Token: tokenAmount,
	})
	if nil != err {
		ctx.Logger().Error("fail to marshal result to json", "error", err)
		// if this happen what should we tell the client?
	}
	return sdk.Result{
		Code: sdk.CodeOK,
		Data: res,
	}
}
