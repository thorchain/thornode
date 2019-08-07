package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/jpthor/cosmos-swap/exchange"

	"github.com/jpthor/cosmos-swap/config"
	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

// NewHandler returns a handler for "swapservice" type messages.
func NewHandler(keeper Keeper, settings *config.Settings) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSetPoolData:
			return handleMsgSetPoolData(ctx, keeper, msg)
		case MsgSetStakeData:
			return handleMsgSetStakeData(ctx, keeper, msg)
		case MsgSwap:
			return handleMsgSwap(ctx, keeper, settings, msg)
		case types.MsgSwapComplete:
			return handleMsgSetSwapComplete(ctx, keeper, msg)
		case types.MsgSetUnStake:
			return handleMsgSetUnstake(ctx, keeper, msg)
		case types.MsgUnStakeComplete:
			return handleMsgSetUnstakeComplete(ctx, keeper, msg)
		case MsgSetTxHash:
			return handleMsgSetTxHash(ctx, keeper, settings, msg)
		case MsgSetTxHashComplete:
			return handleMsgSetTxHashComplete(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized swapservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func isSignedByTrustAccounts(ctx sdk.Context, keeper Keeper, signers []sdk.AccAddress) bool {
	if len(signers) == 0 {
		return false
	}
	for _, signer := range signers {
		if !keeper.IsTrustAccount(ctx, signer) {
			ctx.Logger().Error("unauthorized account", "address", signer.String())
			return false
		}
	}
	return true
}

// Handle a message to set pooldata
func handleMsgSetPoolData(ctx sdk.Context, keeper Keeper, msg MsgSetPoolData) sdk.Result {
	if isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "pool id", msg.PoolID, "pool address", msg.PoolAddress)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	ctx.Logger().Info("handleMsgSetPoolData request", "poolID:"+msg.PoolID)
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ErrUnknownRequest(err.Error()).Result()
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
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// Handle a message to set stake data
func handleMsgSetStakeData(ctx sdk.Context, keeper Keeper, msg MsgSetStakeData) sdk.Result {
	ctx.Logger().Info("handleMsgSetStakeData request", "stakerid:"+msg.Ticker)
	if isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "ticker", msg.Ticker, "request tx hash", msg.RequestTxHash, "public address", msg.PublicAddress)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	if err := stake(
		ctx,
		keeper,
		msg.Name,
		msg.Ticker,
		msg.Rune,
		msg.Token,
		msg.PublicAddress,
		msg.RequestTxHash); err != nil {
		ctx.Logger().Error("fail to process stake message", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// Handle a message to set stake data
func handleMsgSwap(ctx sdk.Context, keeper Keeper, setting *config.Settings, msg MsgSwap) sdk.Result {
	if isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.RequestTxHash, "source ticker", msg.SourceTicker, "target ticker", msg.TargetTicker)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	amount, err := swap(
		ctx,
		keeper,
		setting,
		msg.SourceTicker,
		msg.TargetTicker,
		msg.Amount,
		msg.Requester,
		msg.Destination,
		msg.RequestTxHash,
		msg.SlipLimit,
	) // If so, set the stake data to the value specified in the msg.
	if err != nil {
		ctx.Logger().Error("fail to process swap message", "error", err)
		return sdk.ErrInternal(err.Error()).Result()
	}
	res, err := keeper.cdc.MarshalBinaryLengthPrefixed(struct {
		Token string `json:"token"`
	}{
		Token: amount,
	})
	if nil != err {
		ctx.Logger().Error("fail to encode result to json", "error", err)
		return sdk.ErrInternal("fail to encode result to json").Result()
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      res,
		Codespace: DefaultCodespace,
	}
}

// handleMsgSetSwapComplete mark a swap as complete , record the tx hash.
func handleMsgSetSwapComplete(ctx sdk.Context, keeper Keeper, msg types.MsgSwapComplete) sdk.Result {
	ctx.Logger().Debug("receive MsgSetSwapComplete", "requestTxHash", msg.RequestTxHash, "paytxhash", msg.PayTxHash)
	if isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.RequestTxHash, "pay tx hash", msg.PayTxHash)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgSwapComplete", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	if err := swapComplete(ctx, keeper, msg.RequestTxHash, msg.PayTxHash); nil != err {
		ctx.Logger().Error("fail to set swap to complete", "requestTxHash", msg.RequestTxHash, "paytxhash", msg.PayTxHash)
		return sdk.ErrInternal("fail to mark a swap to complete").Result()
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgSetUnstake process unstake
func handleMsgSetUnstake(ctx sdk.Context, keeper Keeper, msg types.MsgSetUnStake) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgSetUnstake from : %s(%s) unstake (%s)", msg, msg.PublicAddress, msg.Percentage))
	if isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.RequestTxHash, "public address", msg.PublicAddress, "ticker", msg.Ticker, "percentage", msg.Percentage)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgSetUnstake", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	runeAmt, tokenAmount, err := unstake(ctx, keeper, msg)
	if nil != err {
		ctx.Logger().Error("fail to UnStake", "error", err)
		return sdk.ErrInternal("fail to process UnStake request").Result()
	}
	res, err := keeper.cdc.MarshalBinaryLengthPrefixed(struct {
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
		Code:      sdk.CodeOK,
		Data:      res,
		Codespace: DefaultCodespace,
	}
}

func handleMsgSetUnstakeComplete(ctx sdk.Context, keeper Keeper, msg types.MsgUnStakeComplete) sdk.Result {
	ctx.Logger().Debug("receive MsgUnStakeComplete", "requestTxHash", msg.RequestTxHash, "completeTxHash", msg.CompleteTxHash)
	if isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.RequestTxHash)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgUnStakeComplete", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	if err := unStakeComplete(ctx, keeper, msg.RequestTxHash, msg.CompleteTxHash); nil != err {
		ctx.Logger().Error("fail to set swap to complete", "requestTxHash", msg.RequestTxHash, "completetxhash", msg.CompleteTxHash)
		return sdk.ErrInternal("fail to mark a swap to complete").Result()
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgSetTxHash gets a binance tx hash, gets the tx/memo, and triggers
// another handler to process the request
func handleMsgSetTxHash(ctx sdk.Context, keeper Keeper, setting *config.Settings, msg MsgSetTxHash) sdk.Result {
	// validate there are not conflicts first
	if keeper.CheckTxHash(ctx, msg.TxHash.Key()) {
		return sdk.ErrUnknownRequest("Conflict").Result()
	}

	binanceClient := exchange.NewClient()

	txResult, err := binanceClient.GetTxInfo(msg.TxHash.Request)
	if err != nil {
		return sdk.ErrUnknownRequest(
			fmt.Sprintf("Unable to get binance tx info: %s", err.Error()),
		).Result()
	}

	outputs := txResult.Outputs()
	if len(outputs) == 0 {
		return sdk.ErrUnknownRequest("Invalid tx: no outputs").Result()
	}
	inputs := txResult.Inputs()
	if len(inputs) == 0 {
		return sdk.ErrUnknownRequest("Invalid tx: no inputs").Result()
	}
	address := inputs[0].Address
	memo, err := ParseMemo(txResult.Memo())
	if err != nil {
		return sdk.ErrUnknownRequest(
			fmt.Sprintf("Unable to parse memo: %s", err.Error()),
		).Result()
	}

	handler := NewHandler(keeper, setting)
	var newMsg sdk.Msg

	// interpret the memo and initialize a corresponding msg event
	switch memo.(type) {
	case CreateMemo:
		if keeper.PoolExist(ctx, GetPoolNameFromTicker(memo.GetSymbol())) {
			return sdk.ErrUnknownRequest("Pool already exists").Result()
		}
		newMsg = NewMsgSetPoolData(
			"TODO: Name",
			memo.GetSymbol(),
			"TODO: pool address", // prob can be hard coded since its a single pool
			PoolSuspended,        // new pools start in a suspended state
			msg.Signer,
		)
	case StakeMemo:
		runeAmount := "0"
		tokenAmount := "0"
		for _, output := range outputs {
			for _, coin := range output.Coins {
				if coin.Denom == "RUNE-B1A" {
					runeAmount = fmt.Sprintf("%f", coin.Amount)
				}
				if coin.Denom == memo.GetSymbol() {
					tokenAmount = fmt.Sprintf("%f", coin.Amount)
				}
			}
		}
		newMsg = NewMsgSetStakeData(
			"TODO: Name",
			memo.GetSymbol(),
			tokenAmount,
			runeAmount,
			address,
			msg.TxHash.Request,
			msg.Signer,
		)
	case WithdrawMemo:
		newMsg = NewMsgSetUnStake(
			"TODO: name",
			address,
			memo.GetAmount(),
			memo.GetSymbol(),
			msg.TxHash.Request,
			msg.Signer,
		)
	case SwapMemo:
		coin := outputs[0].Coins[0]
		newMsg = NewMsgSwap(
			msg.TxHash.Request,
			coin.Denom,
			memo.GetSymbol(),
			fmt.Sprintf("%f", coin.Amount),
			address,
			memo.GetDestination(),
			fmt.Sprintf("%f", memo.GetSlipLimit()),
			msg.Signer,
		)
	default:
		return sdk.ErrUnknownRequest("Unable to find memo type").Result()
	}

	// trigger msg event
	result := handler(ctx, newMsg)

	// TODO , Based on the new design we will not send the transaction back to binance chain here, instead
	// the tx will be written into a txArray ,  it will be processed by `signer or auditor` when the block commited
	// once the signer or auditor send it to binance chain, it will use MsgSetTxHashComplete to mark it is done
	// given that I think the follow few lines will not be relevant¬
	// Check if our message was successful, if so, save txhash to kvstore, so
	// we don't duplicate this work.
	if result.IsOK() {
		// retrieve done binance hash from result data and save to store
		msg.TxHash.SetDone(string(result.Data))
		keeper.SetTxHash(ctx, msg.TxHash)
	}

	return result
}

// handleMsgSetTxHashComplete will process
func handleMsgSetTxHashComplete(ctx sdk.Context, keeper Keeper, msg MsgSetTxHashComplete) sdk.Result {
	ctx.Logger().Debug("receive MsgSetTxHashComplete", "requestTxHash", msg.RequestTxHash, "complete tx hash", msg.CompleteTxHash)
	if isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.RequestTxHash, "complete tx hash", msg.CompleteTxHash)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgSetTxHashComplete", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	txHash := NewTxHash(msg.RequestTxHash)
	txHash.SetDone(msg.CompleteTxHash)
	keeper.SetTxHash(ctx, txHash)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
