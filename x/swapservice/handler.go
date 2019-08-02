package swapservice

import (
	"fmt"
	"strings"

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

// TODO: this is hacky, should not implement wallet services within the
// handler. Move to a better place
func getWallet(ticker string) (*exchange.Bep2Wallet, error) {
	// TODO: wrap the errors below to be a bit more descriptive
	dir := "~/.ssd/wallets"
	ds, err := storage.NewDataStore(dir, log.Logger)
	if nil != err {
		return nil, err
	}
	ws, err := exchange.NewWallets(ds, log.Logger)
	if err != nil {
		return nil, err
	}

	return ws.GetWallet(ticker)
}

func handleMsgSetPool(ctx sdk.Context, keeper Keeper, msg MsgSetPool) sdk.Result {
	// validate there are not conflicts first
	if keeper.PoolDoesExist(ctx, msg.Pool.Key()) {
		return sdk.ErrUnknownRequest("Conflict").Result()
	}

	wallet, err := getWallet(msg.Pool.TokenTicker)
	if err != nil {
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	msg.Pool.Address, err = sdk.AccAddressFromHex(wallet.PublicAddress)
	if err != nil {
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

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

	txResult, err := exchange.GetTxInfo(msg.TxHash.TxHash)
	if err != nil {
		return sdk.ErrUnknownRequest(
			fmt.Sprintf("Unable to get binance tx info: %s", err.Error()),
		).Result()
	}

	/////////////////////////////////////////////////
	// VALIDATE MEMO ////////////////////////////////
	// Must start with rune
	if !strings.HasPrefix(txResult.Memo(), "rune") {
		// TODO: refund coins back to original wallet
		return sdk.ErrUnknownRequest("Invalid memo: Not rune address").Result()
	}

	// Must be a valid bech32 address
	addr, err := sdk.AccAddressFromHex(txResult.Memo())
	if err != nil {
		// TODO: refund coins back to original wallet
		return sdk.ErrUnknownRequest(
			fmt.Sprintf("Invalid memo: %s", err.Error()),
		).Result()
	}
	/////////////////////////////////////////////////

	// Discover coin
	outputs := txResult.Outputs()
	if len(outputs) == 0 {
		// no outputs
		return sdk.ErrUnknownRequest("No Outputs detected. Try again.").Result()
	}

	for _, output := range outputs {
		// TODO: mint unit tokens for to user's account
		for _, coin := range output.Coins {
			wallet, err := getWallet(coin.Denom)
			if err != nil {
				// TODO: refund coins back to original wallet
				return sdk.ErrUnknownRequest(err.Error()).Result()
			}
			if wallet.PublicAddress != output.Address {
				// addresses don't match
				// TODO should error or something
				continue
			}
			uTokenTicker := fmt.Sprintf("%sU", coin.Denom)

			amt := sdk.NewCoins(
				// TODO: calculate the proper unit toke value (hard coded to 100 for now)
				sdk.NewCoin(uTokenTicker, sdk.NewInt(100)),
			)
			_, err = keeper.coinKeeper.AddCoins(ctx, addr, amt)
			if err != nil {
				// TODO: refund coins back to original wallet
				return sdk.ErrInternal(
					fmt.Sprintf(
						"Unable to Add %s coin. Try again. %s",
						uTokenTicker,
						err.Error(),
					),
				).Result()
			}
		}
	}

	// save that we have successfully handles the transaction
	keeper.SetTxHash(ctx, msg.TxHash)

	return sdk.Result{}
}
