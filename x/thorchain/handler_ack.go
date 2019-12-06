package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AckHandler is to handle Ack message
type AckHandler struct {
	keeper       Keeper
	poolAddrMgr  *PoolAddressManager
	validatorMgr *ValidatorManager
}

// NewAckHandler create new instance of AckHandler
func NewAckHandler(keeper Keeper, poolAddrMgr *PoolAddressManager, validatorMgr *ValidatorManager) AckHandler {
	return AckHandler{
		keeper:       keeper,
		poolAddrMgr:  poolAddrMgr,
		validatorMgr: validatorMgr,
	}
}

// Run it the main entry point to execute Ack logic
func (ah AckHandler) Run(ctx sdk.Context, msg MsgAck, version semver.Version) sdk.Result {
	ctx.Logger().Info("receive ack to next pool pub key",
		"sender address", msg.Sender.String())
	if err := ah.validate(ctx, msg, version); err != nil {
		ctx.Logger().Error("msg ack failed validation", err)
		return err.Result()
	}
	if err := ah.handle(ctx, msg); err != nil {
		ctx.Logger().Error("fail to process msg ack", err)
		return err.Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (ah AckHandler) validate(ctx sdk.Context, msg MsgAck, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return ah.validateV1(ctx, msg)
	} else {
		return errBadVersion
	}
}

func (ah AckHandler) validateV1(ctx sdk.Context, msg MsgAck) sdk.Error {
	if err := msg.ValidateBasic(); nil != err {
		return err
	}
	if !ah.poolAddrMgr.IsRotateWindowOpen {
		return sdk.ErrUnknownRequest("pool rotation window not open")
	}
	if ah.poolAddrMgr.ObservedNextPoolAddrPubKey.IsEmpty() {
		return sdk.ErrUnknownRequest("did not observe next pool address pub key")
	}

	return nil

}
func (ah AckHandler) handle(ctx sdk.Context, msg MsgAck) sdk.Error {
	chainPubKey := ah.poolAddrMgr.ObservedNextPoolAddrPubKey.GetByChain(msg.Chain)
	if nil == chainPubKey {
		return sdk.ErrUnknownRequest(fmt.Sprintf("THORNode donnot have pool for chain %s", msg.Chain))
	}
	addr, err := chainPubKey.PubKey.GetAddress(msg.Chain)
	if nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to get chain(%s) address from pub key: %w", msg.Chain, err).Error())
	}
	if !addr.Equals(msg.Sender) {
		return sdk.ErrUnknownRequest(fmt.Sprintf("observed next pool address and ack address is different,chain(%s)", msg.Chain))
	}
	ah.poolAddrMgr.currentPoolAddresses.Next = ah.poolAddrMgr.currentPoolAddresses.Next.TryAddKey(chainPubKey)
	ah.poolAddrMgr.ObservedNextPoolAddrPubKey = ah.poolAddrMgr.ObservedNextPoolAddrPubKey.TryRemoveKey(chainPubKey)

	nominatedNode := ah.validatorMgr.Meta.Nominated
	queuedNode := ah.validatorMgr.Meta.Queued
	for _, item := range nominatedNode {
		item.TryAddSignerPubKey(chainPubKey.PubKey)
		if err := ah.keeper.SetNodeAccount(ctx, item); nil != err {
			return sdk.ErrInternal(fmt.Errorf("fail to save node account: %w", err).Error())
		}
	}
	activeNodes, err := ah.keeper.ListActiveNodeAccounts(ctx)
	if nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to get all active node accounts: %w", err).Error())
	}

	for _, item := range activeNodes {
		if queuedNode.Contains(item) {
			// queued node doesn't join the signing committee
			continue
		}
		item.TryAddSignerPubKey(chainPubKey.PubKey)
		if err := ah.keeper.SetNodeAccount(ctx, item); nil != err {
			return sdk.ErrInternal(fmt.Errorf("fail to save node account: %w", err).Error())
		}
	}

	if err := AddGasFees(ctx, ah.keeper, msg.Tx.Gas); nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to add gas fee: %w", err).Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeNexePoolPubKeyConfirmed,
			sdk.NewAttribute("pubkey", ah.poolAddrMgr.currentPoolAddresses.Next.String()),
			sdk.NewAttribute("address", msg.Sender.String()),
			sdk.NewAttribute("chain", msg.Chain.String())))
	// THORNode have a pool address confirmed by a chain
	ah.keeper.SetPoolAddresses(ctx, ah.poolAddrMgr.currentPoolAddresses)
	return nil
}
