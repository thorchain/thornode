package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

// LeaveHandler a handler to process leave request
// if an operator of THORChain node would like to leave and get their bond back , they have to
// send a Leave request through Binance Chain
type LeaveHandler struct {
	keeper           Keeper
	validatorManager ValidatorManager
	txOut            TxOutStore
}

// NewLeaveHandler create a new LeaveHandler
func NewLeaveHandler(keeper Keeper, validatorManager ValidatorManager, store TxOutStore) LeaveHandler {
	return LeaveHandler{
		keeper:           keeper,
		validatorManager: validatorManager,
		txOut:            store,
	}
}

func (lh LeaveHandler) validate(ctx sdk.Context, msg MsgLeave, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return lh.validateV1(ctx, msg)
	}
	return errBadVersion
}

func (lh LeaveHandler) validateV1(ctx sdk.Context, msg MsgLeave) sdk.Error {
	if err := msg.ValidateBasic(); nil != err {
		return err
	}
	if !isSignedByActiveObserver(ctx, lh.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized("Not authorized")
	}

	return nil
}

// Run execute the handler
func (lh LeaveHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgLeave)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info("receive MsgLeave",
		"sender", msg.Tx.FromAddress.String(),
		"request tx hash", msg.Tx.ID)
	if err := lh.validate(ctx, msg, version); nil != err {
		ctx.Logger().Error("msg leave fail validation", err)
		return err.Result()
	}

	if err := lh.handle(ctx, msg); nil != err {
		ctx.Logger().Error("fail to process msg leave", err)
		return err.Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
func (lh LeaveHandler) handle(ctx sdk.Context, msg MsgLeave) sdk.Error {
	nodeAcc, err := lh.keeper.GetNodeAccountByBondAddress(ctx, msg.Tx.FromAddress)
	if nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to get node account by bond address: %w", err).Error())
	}
	if nodeAcc.IsEmpty() {
		return sdk.ErrUnknownRequest("node account doesn't exist")
	}
	// THORNode add the node to leave queue

	if nodeAcc.Status == NodeActive {
		if nodeAcc.LeaveHeight == 0 {
			nodeAcc.LeaveHeight = ctx.BlockHeight()
		}
	} else {
		// given the node is not active, they should not have Yggdrasil pool either
		// but let's check it anyway just in case
		vault, err := lh.keeper.GetVault(ctx, nodeAcc.NodePubKey.Secp256k1)
		if nil != err {
			return sdk.ErrInternal(fmt.Errorf("fail to get vault pool: %w", err).Error())
		}
		if !vault.IsYggdrasil() {
			return sdk.ErrInternal("the requested vault is NOT a yggdrasil vault")
		}
		if !vault.HasFunds() {
			// node is not active , they are free to leave , refund them
			if err := refundBond(ctx, msg.Tx.ID, nodeAcc, lh.keeper, lh.txOut); err != nil {
				return sdk.ErrInternal(fmt.Errorf("fail to refund bond: %w", err).Error())
			}
		}

		if err := lh.validatorManager.RequestYggReturn(ctx, nodeAcc, lh.txOut); nil != err {
			return sdk.ErrInternal(fmt.Errorf("fail to request yggdrasil return fund: %w", err).Error())
		}

	}
	nodeAcc.RequestedToLeave = true
	if err := lh.keeper.SetNodeAccount(ctx, nodeAcc); nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to save node account to key value store: %w", err).Error())
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent("validator_request_leave",
			sdk.NewAttribute("signer bnb address", msg.Tx.FromAddress.String()),
			sdk.NewAttribute("destination", nodeAcc.BondAddress.String()),
			sdk.NewAttribute("tx", msg.Tx.ID.String())))

	return nil
}
