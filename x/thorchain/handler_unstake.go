package thorchain

import (
	"encoding/json"
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// UnstakeHandler
type UnstakeHandler struct {
	keeper      Keeper
	txOutStore  TxOutStore
	poolAddrMgr PoolAddressManager
}

// NewUnstakeHandler create a new instance of UnstakeHandler to process unstake request
func NewUnstakeHandler(keeper Keeper, txOutStore TxOutStore, poolAddrMgr PoolAddressManager) UnstakeHandler {
	return UnstakeHandler{
		keeper:      keeper,
		txOutStore:  txOutStore,
		poolAddrMgr: poolAddrMgr,
	}
}
func (uh UnstakeHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgSetUnStake)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info(fmt.Sprintf("receive MsgSetUnstake from : %s(%s) unstake (%s)", msg, msg.RuneAddress, msg.WithdrawBasisPoints))

	if err := uh.validate(ctx, msg, version); err != nil {
		ctx.Logger().Error("msg ack failed validation", err)
		return err.Result()
	}
	data, err := uh.handle(ctx, msg)
	if err != nil {
		ctx.Logger().Error("fail to process msg unstake", err)
		return err.Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      data,
		Codespace: DefaultCodespace,
	}
}

func (uh UnstakeHandler) validate(ctx sdk.Context, msg MsgSetUnStake, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return uh.validateV1(ctx, msg)
	} else {
		return errBadVersion
	}
}

func (uh UnstakeHandler) validateV1(ctx sdk.Context, msg MsgSetUnStake) sdk.Error {
	if err := msg.ValidateBasic(); nil != err {
		return err
	}
	if !isSignedByActiveObserver(ctx, uh.keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account",
			"request tx hash", msg.Tx.ID,
			"rune address", msg.RuneAddress,
			"asset", msg.Asset,
			"withdraw basis points", msg.WithdrawBasisPoints)
		return sdk.ErrUnauthorized("not authorized")
	}

	pool, err := uh.keeper.GetPool(ctx, msg.Asset)
	if err != nil {
		return sdk.ErrInternal(fmt.Errorf("fail to get pool(%s): %w", msg.Asset, err).Error())
	}

	if err := pool.EnsureValidPoolStatus(msg); nil != err {
		return sdk.ErrUnknownRequest(fmt.Errorf("fail to check pool status: %w", err).Error())
	}

	return nil
}

func (uh UnstakeHandler) handle(ctx sdk.Context, msg MsgSetUnStake) ([]byte, sdk.Error) {
	bnbPoolAddr := uh.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(common.BNBChain)
	if nil == bnbPoolAddr {
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("THORNode don't have pool for chain : %s ", common.BNBChain))
	}
	currentAddr := uh.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(msg.Asset.Chain)
	if nil == currentAddr {
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("THORNode don't have pool for chain : %s ", msg.Asset.Chain))
	}
	poolStaker, err := uh.keeper.GetPoolStaker(ctx, msg.Asset)
	if nil != err {
		return nil, sdk.ErrInternal(fmt.Errorf("fail to get pool staker: %w", err).Error())
	}
	stakerUnit := poolStaker.GetStakerUnit(msg.RuneAddress)

	runeAmt, assetAmount, units, err := unstake(ctx, uh.keeper, msg)
	if nil != err {
		return nil, sdk.ErrInternal(fmt.Errorf("fail to process UnStake request: %w", err).Error())
	}
	res, err := uh.keeper.Cdc().MarshalBinaryLengthPrefixed(struct {
		Rune  sdk.Uint `json:"rune"`
		Asset sdk.Uint `json:"asset"`
	}{
		Rune:  runeAmt,
		Asset: assetAmount,
	})
	if nil != err {
		return nil, sdk.ErrInternal(fmt.Errorf("fail to marshal result to json: %w", err).Error())
	}

	unstakeEvt := NewEventUnstake(
		msg.Asset,
		units,
		int64(msg.WithdrawBasisPoints.Uint64()),
		sdk.ZeroDec(), // TODO: What is Asymmetry, how to calculate it?
	)
	unstakeBytes, err := json.Marshal(unstakeEvt)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Errorf("fail to save event: %w", err).Error())
	}
	evt := NewEvent(
		unstakeEvt.Type(),
		ctx.BlockHeight(),
		msg.Tx,
		unstakeBytes,
		EventSuccess,
	)

	if err := uh.keeper.AddIncompleteEvents(ctx, evt); err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}

	toi := &TxOutItem{
		Chain:       common.BNBChain,
		InHash:      msg.Tx.ID,
		VaultPubKey: bnbPoolAddr.PubKey,
		ToAddress:   stakerUnit.RuneAddress,
		Coin:        common.NewCoin(common.RuneAsset(), runeAmt),
	}
	uh.txOutStore.AddTxOutItem(ctx, toi)

	toi = &TxOutItem{
		Chain:       msg.Asset.Chain,
		InHash:      msg.Tx.ID,
		VaultPubKey: currentAddr.PubKey,
		ToAddress:   stakerUnit.AssetAddress,
		Coin:        common.NewCoin(msg.Asset, assetAmount),
	}
	// for unstake , THORNode should deduct fees
	uh.txOutStore.AddTxOutItem(ctx, toi)
	return res, nil
}
