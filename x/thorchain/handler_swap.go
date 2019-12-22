package thorchain

import (
	"fmt"

	"github.com/blang/semver"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type SwapHandler struct {
	keeper      Keeper
	txOutStore  TxOutStore
	poolAddrMgr PoolAddressManager
}

func NewSwapHandler(keeper Keeper, txOutStore TxOutStore, poolAddrMgr PoolAddressManager) SwapHandler {
	return SwapHandler{
		keeper:      keeper,
		txOutStore:  txOutStore,
		poolAddrMgr: poolAddrMgr,
	}
}

func (h SwapHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgSwap)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.validate(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	return h.handle(ctx, msg, version, constAccessor)
}

func (h SwapHandler) validate(ctx sdk.Context, msg MsgSwap, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h SwapHandler) validateV1(ctx sdk.Context, msg MsgSwap) error {
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return err
	}

	if !isSignedByActiveObserver(ctx, h.keeper, msg.GetSigners()) {
		ctx.Logger().Error(notAuthorized.Error())
		return notAuthorized
	}
	return nil
}

func (h SwapHandler) handle(ctx sdk.Context, msg MsgSwap, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	ctx.Logger().Info("receive MsgSwap", "request tx hash", msg.Tx.ID, "source asset", msg.Tx.Coins[0].Asset, "target asset", msg.TargetAsset, "signer", msg.Signer.String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg, constAccessor)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return errBadVersion.Result()
	}
}

func (h SwapHandler) handleV1(ctx sdk.Context, msg MsgSwap, constAccessor constants.ConstantValues) sdk.Result {
	globalSlipLimit := constAccessor.GetInt64Value(constants.GlobalSlipLimit)
	gsl := sdk.NewUint(uint64(globalSlipLimit))
	chain := msg.TargetAsset.Chain
	currentAddr := h.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(chain)
	if nil == currentAddr {
		msg := fmt.Sprintf("don't have pool address for chain : %s", chain)
		ctx.Logger().Error(msg)
		return sdk.ErrInternal(msg).Result()
	}
	amount, err := swap(
		ctx,
		h.keeper,
		msg.Tx,
		msg.TargetAsset,
		msg.Destination,
		msg.TradeTarget,
		gsl,
	)
	if err != nil {
		ctx.Logger().Error("fail to process swap message", "error", err)
		return sdk.ErrInternal(err.Error()).Result()
	}

	res, err := h.keeper.Cdc().MarshalBinaryLengthPrefixed(
		struct {
			Asset sdk.Uint `json:"asset"`
		}{
			Asset: amount,
		})

	if nil != err {
		ctx.Logger().Error("fail to encode result to json", "error", err)
		return sdk.ErrInternal("fail to encode result to json").Result()
	}

	toi := &TxOutItem{
		Chain:       currentAddr.Chain,
		InHash:      msg.Tx.ID,
		VaultPubKey: currentAddr.PubKey,
		ToAddress:   msg.Destination,
		Coin:        common.NewCoin(msg.TargetAsset, amount),
	}
	h.txOutStore.AddTxOutItem(ctx, toi)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      res,
		Codespace: DefaultCodespace,
	}
}
