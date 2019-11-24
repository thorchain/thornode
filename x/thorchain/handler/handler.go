package handler

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Handler interface {
	Validate(ctx sdk.Context, msg sdk.Msg, version int64) error
	Log(ctx sdk.Context, msg sdk.Msg)
	Handle(ctx sdk.Context, msg sdk.Msg, version int64) error
}

// NewHandler returns a handler for "thorchain" type messages.
func NewHandler(keeper Keeper, poolAddressMgr *PoolAddressManager, txOutStore *TxOutStore, validatorManager *ValidatorManager) sdk.Handler {

	poolDataHandler := NewPoolDataHandler(keeper)

	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		version := keeper.GetLowestActiveVersion(ctx)
		switch m := msg.(type) {
		case MsgSetPoolData:
			if err := poolDataHandler.Validate(ctx, m, version); err != nil {
				return sdk.ErrUnauthorized(err.Error()).Result()
			}
			poolDataHandler.Log(ctx, m)
			if err := poolDataHandler.Handle(ctx, m, version); err != nil {
				return sdk.ErrUnauthorized(err.Error()).Result()
			}
		default:
			errMsg := fmt.Sprintf("Unrecognized thorchain Msg type: %v", m)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}

		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}
}
