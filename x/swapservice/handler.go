package swapservice

import (
	"fmt"

	exchange "github.com/jpthor/cosmos-swap/exchange"
	storage "github.com/jpthor/cosmos-swap/storage"
	"github.com/rs/zerolog/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "swapservice" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSetPool:
			return handleMsgSetPool(ctx, keeper, msg)
		case MsgSetTxHash:
			return handleMsgSetTxHash(ctx, keeper, msg)
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

	/////////////////////////////////////////////////////////////////////
	// TODO: this is hacky, should not implement wallet services within the
	// handler
	// TODO: wrap the errors below to be a bit more descriptive
	/////////////////////////////////////////////////////////////////////
	dir := "~/.ssd/wallets"
	ds, err := storage.NewDataStore(dir, log.Logger)
	if nil != err {
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	ws, err := exchange.NewWallets(ds, log.Logger)
	if err != nil {
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	wallet, err := ws.GetWallet(msg.Pool.TokenTicker)
	if err != nil {
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	msg.Pool.Address, err = sdk.AccAddressFromHex(wallet.PublicAddress)
	if err != nil {
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	////////////////////////////////////////////////////////////////////

	if msg.Pool.Empty() {
		return sdk.ErrUnknownRequest("Invalid Pool").Result()
	}

	keeper.SetPool(ctx, msg.Pool)

	return sdk.Result{}
}

func handleMsgSetTxHash(ctx sdk.Context, keeper Keeper, msg MsgSetTxHash) sdk.Result {
	// validate there are not conflicts first
	if keeper.TxDoesExist(ctx, msg.TxHash.Key()) {
		return sdk.ErrUnknownRequest("Conflict").Result()
	}

	// TODO: check tx hash against binance API
	// TODO: mint unit tokens for to user's account

	keeper.SetTxHash(ctx, msg.TxHash)

	return sdk.Result{}
}
