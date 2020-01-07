package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// THORChain error code start at 101
const (
	// CodeBadVersion error code for bad version
	CodeBadVersion            sdk.CodeType = 101
	CodeInvalidMessage        sdk.CodeType = 102
	CodeConstantsNotAvailable sdk.CodeType = 103
	CodeInvalidVault          sdk.CodeType = 104
	CodeInvalidMemo           sdk.CodeType = 105
	CodeValidationError       sdk.CodeType = 106
	CodeInvalidPoolStatus     sdk.CodeType = 107

	CodeSwapFail                 sdk.CodeType = 108
	CodeSwapFailTradeTarget      sdk.CodeType = 109
	CodeSwapFailNotEnoughFee     sdk.CodeType = 110
	CodeSwapFailZeroEmitAsset    sdk.CodeType = 111
	CodeSwapFailPoolNotExist     sdk.CodeType = 112
	CodeSwapFailInvalidAmount    sdk.CodeType = 113
	CodeSwapFailInvalidBalance   sdk.CodeType = 114
	CodeSwapFailNotEnoughBalance sdk.CodeType = 115

	CodeStakeFailValidation    sdk.CodeType = 120
	CodeFailGetPoolStaker      sdk.CodeType = 122
	CodeStakeMismatchAssetAddr sdk.CodeType = 123
	CodeStakeInvalidPoolAsset  sdk.CodeType = 124

	CodeUnstakeFailValidation sdk.CodeType = 130
	CodeFailAddOutboundTx     sdk.CodeType = 131
	CodeFailSaveEvent         sdk.CodeType = 132
	CodePoolStakerNotExist    sdk.CodeType = 133
	CodeStakerPoolNotExist    sdk.CodeType = 134
	CodeNoStakeUnitLeft       sdk.CodeType = 135
	CodeWithdrawWithin24Hours sdk.CodeType = 136
	CodeUnstakeFail           sdk.CodeType = 137
)

// EmptyAccAddress empty address
var EmptyAccAddress = sdk.AccAddress{}
var notAuthorized = fmt.Errorf("not authorized")
var badVersion = fmt.Errorf("bad version")
var errBadVersion = sdk.NewError(DefaultCodespace, CodeBadVersion, "bad version")
var errInvalidMessage = sdk.NewError(DefaultCodespace, CodeInvalidMessage, "invalid message")
var errConstNotAvailable = sdk.NewError(DefaultCodespace, CodeConstantsNotAvailable, "constant values not available")

// NewHandler returns a handler for "thorchain" type messages.
func NewHandler(keeper Keeper, txOutStore TxOutStore, validatorMgr ValidatorManager, vaultMgr VaultManager) sdk.Handler {
	handlerMap := getHandlerMapping(keeper, txOutStore, validatorMgr, vaultMgr)

	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		version := keeper.GetLowestActiveVersion(ctx)
		constantValues := constants.GetConstantValues(version)
		if nil == constantValues {
			return errConstNotAvailable.Result()
		}
		h, ok := handlerMap[msg.Type()]
		if !ok {
			errMsg := fmt.Sprintf("Unrecognized thorchain Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
		return h.Run(ctx, msg, version, constantValues)
	}
}

func getHandlerMapping(keeper Keeper, txOutStore TxOutStore, validatorMgr ValidatorManager, vaultMgr VaultManager) map[string]MsgHandler {
	// New arch handlers
	m := make(map[string]MsgHandler)
	m[MsgOutboundTx{}.Type()] = NewOutboundTxHandler(keeper)
	m[MsgTssPool{}.Type()] = NewTssHandler(keeper, vaultMgr)
	m[MsgNoOp{}.Type()] = NewNoOpHandler(keeper)
	m[MsgYggdrasil{}.Type()] = NewYggdrasilHandler(keeper, txOutStore, validatorMgr)
	m[MsgEndPool{}.Type()] = NewEndPoolHandler(keeper, txOutStore)
	m[MsgSetNodeKeys{}.Type()] = NewSetNodeKeysHandler(keeper)
	m[MsgSetAdminConfig{}.Type()] = NewSetAdminConfigHandler(keeper)
	m[MsgSwap{}.Type()] = NewSwapHandler(keeper, txOutStore)
	m[MsgReserveContributor{}.Type()] = NewReserveContributorHandler(keeper)
	m[MsgSetPoolData{}.Type()] = NewPoolDataHandler(keeper)
	m[MsgSetVersion{}.Type()] = NewVersionHandler(keeper)
	m[MsgBond{}.Type()] = NewBondHandler(keeper)
	m[MsgObservedTxIn{}.Type()] = NewObservedTxInHandler(keeper, txOutStore, validatorMgr, vaultMgr)
	m[MsgObservedTxOut{}.Type()] = NewObservedTxOutHandler(keeper, txOutStore, validatorMgr, vaultMgr)
	m[MsgLeave{}.Type()] = NewLeaveHandler(keeper, validatorMgr, txOutStore)
	m[MsgAdd{}.Type()] = NewAddHandler(keeper)
	m[MsgSetUnStake{}.Type()] = NewUnstakeHandler(keeper, txOutStore)
	m[MsgSetStakeData{}.Type()] = NewStakeHandler(keeper)
	m[MsgRefundTx{}.Type()] = NewRefundHandler(keeper)
	return m
}

func processOneTxIn(ctx sdk.Context, keeper Keeper, tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	if len(tx.Tx.Coins) == 0 {
		return nil, fmt.Errorf("no coin found")
	}
	memo, err := ParseMemo(tx.Tx.Memo)
	if err != nil {
		return nil, errors.Wrap(err, "fail to parse memo")
	}
	// THORNode should not have one tx across chain, if it is cross chain it should be separate tx
	var newMsg sdk.Msg
	// interpret the memo and initialize a corresponding msg event
	switch m := memo.(type) {
	case CreateMemo:
		newMsg, err = getMsgSetPoolDataFromMemo(ctx, keeper, m, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgSetPoolData from memo")
		}

	case StakeMemo:
		newMsg, err = getMsgStakeFromMemo(ctx, m, tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgStake from memo")
		}

	case WithdrawMemo:
		newMsg, err = getMsgUnstakeFromMemo(m, tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgUnstake from memo")
		}
	case SwapMemo:
		newMsg, err = getMsgSwapFromMemo(m, tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgSwap from memo")
		}
	case AddMemo:
		newMsg, err = getMsgAddFromMemo(m, tx, signer)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get MsgAdd from memo")
		}
	case GasMemo:
		newMsg, err = getMsgNoOpFromMemo(tx, signer)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get MsgNoOp from memo")
		}
	case RefundMemo:
		newMsg, err = getMsgRefundFromMemo(m, tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgRefund from memo")
		}
	case OutboundMemo:
		newMsg, err = getMsgOutboundFromMemo(m, tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgOutbound from memo")
		}
	case BondMemo:
		newMsg, err = getMsgBondFromMemo(m, tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgBond from memo")
		}
	case LeaveMemo:
		newMsg = NewMsgLeave(tx.Tx, signer)
	case YggdrasilFundMemo:
		newMsg = NewMsgYggdrasil(tx.ObservedPubKey, true, tx.Tx.Coins, tx.Tx.ID, signer)
	case YggdrasilReturnMemo:
		newMsg = NewMsgYggdrasil(tx.ObservedPubKey, false, tx.Tx.Coins, tx.Tx.ID, signer)
	case ReserveMemo:
		res := NewReserveContributor(tx.Tx.FromAddress, tx.Tx.Coins[0].Amount)
		newMsg = NewMsgReserveContributor(res, signer)
	default:
		return nil, errors.Wrap(err, "Unable to find memo type")
	}

	if err := newMsg.ValidateBasic(); nil != err {
		return nil, errors.Wrap(err, "invalid msg")
	}
	return newMsg, nil
}

func getMsgNoOpFromMemo(tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	for _, coin := range tx.Tx.Coins {
		if !coin.Asset.IsBNB() {
			return nil, errors.New("Only accepts BNB coins")
		}
	}
	return NewMsgNoOp(signer), nil
}

func getMsgSwapFromMemo(memo SwapMemo, tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	if len(tx.Tx.Coins) > 1 {
		return nil, errors.New("not expecting multiple coins in a swap")
	}
	if memo.Destination.IsEmpty() {
		memo.Destination = tx.Tx.FromAddress
	}

	coin := tx.Tx.Coins[0]
	if memo.Asset.Equals(coin.Asset) {
		return nil, errors.Errorf("swap from %s to %s is noop, refund", memo.Asset.String(), coin.Asset.String())
	}

	// Looks like at the moment THORNode can only process ont ty
	return NewMsgSwap(tx.Tx, memo.GetAsset(), memo.Destination, memo.SlipLimit, signer), nil
}

func getMsgUnstakeFromMemo(memo WithdrawMemo, tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	withdrawAmount := sdk.NewUint(MaxWithdrawBasisPoints)
	if len(memo.GetAmount()) > 0 {
		withdrawAmount = sdk.NewUintFromString(memo.GetAmount())
	}
	return NewMsgSetUnStake(tx.Tx, tx.Tx.FromAddress, withdrawAmount, memo.GetAsset(), signer), nil
}

func getMsgStakeFromMemo(ctx sdk.Context, memo StakeMemo, tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	if len(tx.Tx.Coins) > 2 {
		return nil, errors.New("not expecting more than two coins in a stake")
	}
	runeAmount := sdk.ZeroUint()
	assetAmount := sdk.ZeroUint()
	asset := memo.GetAsset()
	if asset.IsEmpty() {
		return nil, errors.New("Unable to determine the intended pool for this stake")
	}
	if asset.IsRune() {
		return nil, errors.New("invalid pool asset")
	}
	for _, coin := range tx.Tx.Coins {
		ctx.Logger().Info("coin", "asset", coin.Asset.String(), "amount", coin.Amount.String())
		if coin.Asset.IsRune() {
			runeAmount = coin.Amount
		}
		if asset.Equals(coin.Asset) {
			assetAmount = coin.Amount
		}
	}

	if runeAmount.IsZero() && assetAmount.IsZero() {
		return nil, errors.New("did not find any valid coins for stake")
	}

	// when THORNode receive two coins, but THORNode didn't find the coin specify by asset, then user might send in the wrong coin
	if assetAmount.IsZero() && len(tx.Tx.Coins) == 2 {
		return nil, errors.Errorf("did not find %s ", asset)
	}

	runeAddr := tx.Tx.FromAddress
	assetAddr := memo.GetDestination()
	if !runeAddr.IsChain(common.BNBChain) {
		runeAddr = memo.GetDestination()
		assetAddr = tx.Tx.FromAddress
	} else {
		// if it is on BNB chain , while the asset addr is empty, then the asset addr is runeAddr
		if assetAddr.IsEmpty() {
			assetAddr = runeAddr
		}
	}

	return NewMsgSetStakeData(
		tx.Tx,
		asset,
		runeAmount,
		assetAmount,
		runeAddr,
		assetAddr,
		signer,
	), nil
}

func getMsgSetPoolDataFromMemo(ctx sdk.Context, keeper Keeper, memo CreateMemo, signer sdk.AccAddress) (sdk.Msg, error) {
	if keeper.PoolExist(ctx, memo.GetAsset()) {
		return nil, errors.New("pool already exists")
	}
	return NewMsgSetPoolData(
		memo.GetAsset(),
		PoolEnabled, // new pools start in a Bootstrap state
		signer,
	), nil
}

func getMsgAddFromMemo(memo AddMemo, tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	runeAmount := sdk.ZeroUint()
	assetAmount := sdk.ZeroUint()
	for _, coin := range tx.Tx.Coins {
		if coin.Asset.IsRune() {
			runeAmount = coin.Amount
		} else if memo.GetAsset().Equals(coin.Asset) {
			assetAmount = coin.Amount
		}
	}
	return NewMsgAdd(
		tx.Tx,
		memo.GetAsset(),
		runeAmount,
		assetAmount,
		signer,
	), nil
}

func getMsgRefundFromMemo(memo RefundMemo, tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	return NewMsgRefundTx(
		tx,
		memo.GetTxID(),
		signer,
	), nil
}

func getMsgOutboundFromMemo(memo OutboundMemo, tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	return NewMsgOutboundTx(
		tx,
		memo.GetTxID(),
		signer,
	), nil
}

func getMsgBondFromMemo(memo BondMemo, tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	runeAmount := sdk.ZeroUint()
	for _, coin := range tx.Tx.Coins {
		if coin.Asset.IsRune() {
			runeAmount = coin.Amount
		}
	}
	if runeAmount.IsZero() {
		return nil, errors.New("RUNE amount is 0")
	}
	return NewMsgBond(memo.GetNodeAddress(), runeAmount, tx.Tx.ID, tx.Tx.FromAddress, signer), nil
}
