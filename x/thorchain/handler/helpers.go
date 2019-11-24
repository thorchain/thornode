package handler

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var notAuthorized = fmt.Errorf("Not Authorized")

// isSignedByActiveObserver check whether the signers are all active observer
func isSignedByActiveObserver(ctx sdk.Context, keeper Keeper, signers []sdk.AccAddress) error {
	if len(signers) == 0 {
		return notAuthorized
	}
	for _, signer := range signers {
		if !keeper.IsActiveObserver(ctx, signer) {
			return notAuthorized
		}
	}
	return nil
}

func isSignedByActiveNodeAccounts(ctx sdk.Context, keeper Keeper, signers []sdk.AccAddress) error {
	if len(signers) == 0 {
		return notAuthorized
	}
	for _, signer := range signers {
		nodeAccount, err := keeper.GetNodeAccount(ctx, signer)
		if err != nil {
			ctx.Logger().Error("unauthorized account", "address", signer.String())
			return notAuthorized
		}
		if nodeAccount.IsEmpty() {
			ctx.Logger().Error("unauthorized account", "address", signer.String())
			return notAuthorized
		}
		if nodeAccount.Status != NodeActive {
			ctx.Logger().Error("unauthorized account, node account not active", "address", signer.String(), "status", nodeAccount.Status)
			return notAuthorized
		}
	}
	return nil
}
