package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "swapservice" type messages.
func NewHandler(keeper Keeper, txOutStore *TxOutStore) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch m := msg.(type) {
		case MsgSetPoolData:
			return handleMsgSetPoolData(ctx, keeper, m)
		case MsgSetStakeData:
			result := handleMsgSetStakeData(ctx, keeper, m)
			processRefund(result, txOutStore, m)
			return result
		case MsgSwap:
			result := handleMsgSwap(ctx, keeper, txOutStore, m)
			processRefund(result, txOutStore, m)
			return result
		case MsgSwapComplete:
			return handleMsgSetSwapComplete(ctx, keeper, m)
		case MsgSetUnStake:
			return handleMsgSetUnstake(ctx, keeper, txOutStore, m)
		case MsgUnStakeComplete:
			return handleMsgSetUnstakeComplete(ctx, keeper, m)
		case MsgSetTxHash:
			return handleMsgSetTxHash(ctx, keeper, txOutStore, m)
		case MsgSetAdminConfig:
			return handleMsgSetAdminConfig(ctx, keeper, m)
		default:
			errMsg := fmt.Sprintf("Unrecognized swapservice Msg type: %v", m.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// processRefund take in the sdk.Result and decide whether we should refund customer
func processRefund(result sdk.Result, store *TxOutStore, msg sdk.Msg) {
	if result.IsOK() {
		return
	}
	switch m := msg.(type) {
	case MsgSetStakeData:
		toi := &TxOutItem{
			ToAddress: m.PublicAddress,
		}
		toi.Coins = append(toi.Coins, Coin{
			Denom:  RuneTicker,
			Amount: m.RuneAmount,
		})
		toi.Coins = append(toi.Coins, Coin{
			Denom:  m.Ticker,
			Amount: m.TokenAmount,
		})
		store.AddTxOutItem(toi)
	case MsgSwap:
		toi := &TxOutItem{
			ToAddress: m.Requester,
		}
		toi.Coins = append(toi.Coins, Coin{
			Denom:  m.SourceTicker,
			Amount: m.Amount,
		})
		store.AddTxOutItem(toi)
	default:
		return
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
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "ticker", msg.Ticker, "pool address", msg.PoolAddress)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	ctx.Logger().Info("handleMsgSetPoolData request", "Ticker:"+msg.Ticker)
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	keeper.SetPoolData(
		ctx,
		msg.Ticker,
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
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "ticker", msg.Ticker, "request tx hash", msg.RequestTxHash, "public address", msg.PublicAddress)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := keeper.GetPoolStruct(ctx, msg.Ticker).EnsureValidPoolStatus(msg); nil != err {
		ctx.Logger().Error("check pool status", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	if err := stake(
		ctx,
		keeper,
		msg.Ticker,
		msg.RuneAmount,
		msg.TokenAmount,
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
func handleMsgSwap(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore, msg MsgSwap) sdk.Result {
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.RequestTxHash, "source ticker", msg.SourceTicker, "target ticker", msg.TargetTicker)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}

	gslConfig := keeper.GetAdminConfig(ctx, "GSL")
	tslConfig := keeper.GetAdminConfig(ctx, "TSL")

	amount, err := swap(
		ctx,
		keeper,
		msg.SourceTicker,
		msg.TargetTicker,
		msg.Amount,
		msg.Requester,
		msg.Destination,
		msg.RequestTxHash,
		tslConfig.Value,
		gslConfig.Value,
	) // If so, set the stake data to the value specified in the msg.
	if err != nil {
		ctx.Logger().Error("fail to process swap message", "error", err)

		return sdk.ErrInternal(err.Error()).Result()
	}
	res, err := keeper.cdc.MarshalBinaryLengthPrefixed(struct {
		Token Amount `json:"token"`
	}{
		Token: amount,
	})
	if nil != err {
		ctx.Logger().Error("fail to encode result to json", "error", err)
		return sdk.ErrInternal("fail to encode result to json").Result()
	}
	toi := &TxOutItem{
		ToAddress: msg.Destination,
	}
	toi.Coins = append(toi.Coins, Coin{
		Denom:  msg.TargetTicker,
		Amount: amount,
	})
	txOutStore.AddTxOutItem(toi)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      res,
		Codespace: DefaultCodespace,
	}
}

// handleMsgSetSwapComplete mark a swap as complete , record the tx hash.
func handleMsgSetSwapComplete(ctx sdk.Context, keeper Keeper, msg MsgSwapComplete) sdk.Result {
	ctx.Logger().Debug("receive MsgSetSwapComplete", "requestTxHash", msg.RequestTxHash, "paytxhash", msg.PayTxHash)
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
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
func handleMsgSetUnstake(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore, msg MsgSetUnStake) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgSetUnstake from : %s(%s) unstake (%s)", msg, msg.PublicAddress, msg.Percentage))
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.RequestTxHash, "public address", msg.PublicAddress, "ticker", msg.Ticker, "percentage", msg.Percentage)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := keeper.GetPoolStruct(ctx, msg.Ticker).EnsureValidPoolStatus(msg); nil != err {
		ctx.Logger().Error("check pool status", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
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
		Rune  Amount `json:"rune"`
		Token Amount `json:"token"`
	}{
		Rune:  runeAmt,
		Token: tokenAmount,
	})
	if nil != err {
		ctx.Logger().Error("fail to marshal result to json", "error", err)
		// if this happen what should we tell the client?
	}
	toi := &TxOutItem{
		ToAddress: msg.PublicAddress,
	}
	toi.Coins = append(toi.Coins, Coin{
		Denom:  RuneTicker,
		Amount: runeAmt,
	})
	toi.Coins = append(toi.Coins, Coin{
		Denom:  msg.Ticker,
		Amount: tokenAmount,
	})
	txOutStore.AddTxOutItem(toi)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      res,
		Codespace: DefaultCodespace,
	}
}

func handleMsgSetUnstakeComplete(ctx sdk.Context, keeper Keeper, msg MsgUnStakeComplete) sdk.Result {
	ctx.Logger().Debug("receive MsgUnStakeComplete", "requestTxHash", msg.RequestTxHash, "completeTxHash", msg.CompleteTxHash)
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
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

func refundTx(tx TxHash, store *TxOutStore) {
	toi := &TxOutItem{
		ToAddress: tx.Sender,
	}

	for _, item := range tx.Coins {
		toi.Coins = append(toi.Coins, Coin{
			Denom:  Ticker(item.Denom),
			Amount: Amount(item.Amount.String()),
		})
	}
	store.AddTxOutItem(toi)
}

// handleMsgSetTxHash gets a binance tx hash, gets the tx/memo, and triggers
// another handler to process the request
func handleMsgSetTxHash(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore, msg MsgSetTxHash) sdk.Result {

	conflicts := make([]string, 0)
	todo := make([]TxHash, 0)
	for _, tx := range msg.TxHashes {
		// validate there are not conflicts first
		if keeper.CheckTxHash(ctx, tx.Key()) {
			conflicts = append(conflicts, tx.Key())
		} else {
			todo = append(todo, tx)
		}
	}

	handler := NewHandler(keeper, txOutStore)
	for _, tx := range todo {
		memo, err := ParseMemo(tx.Memo)
		if err != nil {
			ctx.Logger().Error("fail to parse memo", "memo", memo, "error", err)
			// skip over message with bad memos
			refundTx(tx, txOutStore)
			continue
		}

		// save the tx to db to stop duplicate request (aka replay attacks)
		keeper.SetTxHash(ctx, tx)

		var newMsg sdk.Msg

		// interpret the memo and initialize a corresponding msg event
		switch memo.(type) {
		case CreateMemo:
			ticker, err := NewTicker(memo.GetSymbol())
			if err != nil {
				return sdk.ErrUnknownRequest(err.Error()).Result()
			}
			if keeper.PoolExist(ctx, ticker) {
				return sdk.ErrUnknownRequest("Pool already exists").Result()
			}
			newMsg = NewMsgSetPoolData(
				ticker,
				"TODO: pool address", // prob can be hard coded since its a single pool
				PoolSuspended,        // new pools start in a suspended state
				msg.Signer,
			)
		case StakeMemo:
			ticker, err := NewTicker(memo.GetSymbol())
			if err != nil {
				return sdk.ErrUnknownRequest(err.Error()).Result()
			}
			runeAmount := ZeroAmount
			tokenAmount := ZeroAmount
			for _, coin := range tx.Coins {
				if coin.Denom == "RUNE-B1A" {
					runeAmount, err = NewAmount(fmt.Sprintf("%f", coin.Amount))
					if err != nil {
						return sdk.ErrUnknownRequest(err.Error()).Result()
					}
				} else {
					tokenAmount, err = NewAmount(fmt.Sprintf("%f", coin.Amount))
					if err != nil {
						return sdk.ErrUnknownRequest(err.Error()).Result()
					}
				}
			}
			newMsg = NewMsgSetStakeData(
				"TODO: Name",
				ticker,
				tokenAmount,
				runeAmount,
				tx.Sender,
				tx.Request,
				msg.Signer,
			)
		case AdminMemo:

			if memo.GetAdminType() == adminPoolStatus {
				ticker, err := NewTicker(memo.GetKey())
				if err != nil {
					return sdk.ErrUnknownRequest(err.Error()).Result()
				}
				pool := keeper.GetPoolStruct(ctx, ticker)
				if pool.Empty() {
					return sdk.ErrUnknownRequest("Pool doesn't exist").Result()
				}
				status := GetPoolStatus(memo.GetValue())
				newMsg = NewMsgSetPoolData(
					ticker,
					pool.PoolAddress,
					status,
					msg.Signer,
				)

			} else if memo.GetAdminType() == adminKey {
				newMsg = NewMsgSetAdminConfig(memo.GetKey(), memo.GetValue(), msg.Signer)
			} else {
				return sdk.ErrUnknownRequest("Invalid admin command type").Result()
			}
		case WithdrawMemo:
			// do nothing. Let the outTx process these
		case SwapMemo:
			// do nothing. Let the outTx process these
		default:
			return sdk.ErrUnknownRequest("Unable to find memo type").Result()
		}

		// trigger msg event (
		handler(ctx, newMsg)
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
		// Data:      []byte(strings.Join(conflicts, ", ")),
	}
}

// handleMsgSetAdminConfig process admin config
func handleMsgSetAdminConfig(ctx sdk.Context, keeper Keeper, msg MsgSetAdminConfig) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgSetAdminConfig %s --> %s", msg.AdminConfig.Key, msg.AdminConfig.Value))
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account")
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgSetAdminConfig", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	keeper.SetAdminConfig(ctx, msg.AdminConfig)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
