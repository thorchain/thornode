package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// HandlerNextPoolAddress is to process confirm next pool address
// MsgNextPoolAddress is a way to prove that the operator has access to the address, and can sign transaction with the given address on chain
type HandlerNextPoolAddress struct {
	keeper          Keeper
	poolAddrManager PoolAddressManager
	validatorMgr    ValidatorManager
	txOut           TxOutStore
}

// NewHandlerNextPoolAddress create a new handler to process confirm next pool address
func NewHandlerNextPoolAddress(keeper Keeper, poolAddressManager PoolAddressManager, validatorMgr ValidatorManager, store TxOutStore) HandlerNextPoolAddress {
	return HandlerNextPoolAddress{
		keeper:          keeper,
		poolAddrManager: poolAddressManager,
		validatorMgr:    validatorMgr,
		txOut:           store,
	}
}

// Run execute the handler
func (h HandlerNextPoolAddress) Run(ctx sdk.Context, m sdk.Msg, version semver.Version) sdk.Result {
	msg, ok := m.(MsgNextPoolAddress)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info("receive request to set next pool pub key",
		"next pool pub key", msg.NextPoolPubKey.String())

	if err := h.validate(ctx, msg, version); nil != err {
		return err.Result()
	}

	if err := h.handle(ctx, msg, version); nil != err {
		return err.Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (h HandlerNextPoolAddress) validate(ctx sdk.Context, msg MsgNextPoolAddress, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	}
	return errBadVersion
}

func (h HandlerNextPoolAddress) validateV1(ctx sdk.Context, msg MsgNextPoolAddress) sdk.Error {
	if err := msg.ValidateBasic(); nil != err {
		return err
	}
	if !h.poolAddrManager.IsRotateWindowOpen() {
		return sdk.ErrUnknownRequest("pool address rotate window not open yet")
	}

	return nil
}

func (h HandlerNextPoolAddress) handle(ctx sdk.Context, msg MsgNextPoolAddress, version semver.Version) sdk.Error {
	currentPoolAddresses := h.poolAddrManager.GetCurrentPoolAddresses()
	currentChainPoolAddr := currentPoolAddresses.Next.GetByChain(msg.Chain)
	if nil != currentChainPoolAddr {
		return sdk.ErrUnknownRequest(fmt.Sprintf("next pool for chain %s had been observed already", msg.Chain))
	}
	currentAddr := currentPoolAddresses.Current.GetByChain(msg.Chain)
	if nil == currentAddr || currentAddr.IsEmpty() {
		return sdk.ErrUnknownRequest(fmt.Sprintf("THORNode donnot have pool for chain %s", msg.Chain))
	}
	addr, err := currentAddr.PubKey.GetAddress(msg.Chain)
	if nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to get address from pub key, chain(%s): %w", msg.Chain, err).Error())
	}
	// next pool memo need to be initiated by current pool
	if !addr.Equals(msg.Sender) {
		return sdk.ErrInvalidAddress("next pool should be send with current pool address")
	}
	// THORChain observed the next pool address memo, but it has not been confirmed yet
	pkey, err := common.NewPoolPubKey(msg.Chain, 0, msg.NextPoolPubKey)
	if nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to get pool pubkey for chain(%s): %w", msg.Chain, err).Error())
	}

	h.poolAddrManager.SetObservedNextPoolAddrPubKey(h.poolAddrManager.ObservedNextPoolAddrPubKey().TryAddKey(pkey))

	// if THORNode observed a valid nextpool transaction, that means the nominated validator had join the signing committee to generate a new pub key
	// with TSS, if they don't join , then the key won't be generated
	nominatedAccount := h.validatorMgr.Meta().Nominated
	if !nominatedAccount.IsEmpty() {
		for _, item := range nominatedAccount {
			item.SignerActive = true
			if err := h.keeper.SetNodeAccount(ctx, item); nil != err {
				return sdk.ErrInternal(fmt.Errorf("fail to save node account: %w", err).Error())
			}
		}
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeNextPoolPubKeyObserved,
			sdk.NewAttribute("next pool pub key", msg.NextPoolPubKey.String()),
			sdk.NewAttribute("chain", msg.Chain.String())))

	// instruct signer to send an ack back from next pool address
	h.txOut.AddTxOutItem(ctx, &TxOutItem{
		Chain:       common.BNBChain,
		ToAddress:   addr,
		VaultPubKey: msg.NextPoolPubKey,
		Coin:        common.NewCoin(common.BNBAsset, sdk.NewUint(1)),
		Memo:        "ack",
	})
	return nil
}
