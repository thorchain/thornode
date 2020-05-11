package thorchain

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	CodeFailGetStaker          sdk.CodeType = 122
	CodeStakeMismatchAssetAddr sdk.CodeType = 123
	CodeStakeInvalidPoolAsset  sdk.CodeType = 124
	CodeStakeRUNEOverLimit     sdk.CodeType = 125
	CodeStakeRUNEMoreThanBond  sdk.CodeType = 126

	CodeUnstakeFailValidation sdk.CodeType = 130
	CodeFailAddOutboundTx     sdk.CodeType = 131
	CodeFailSaveEvent         sdk.CodeType = 132
	CodeStakerNotExist        sdk.CodeType = 133
	CodeNoStakeUnitLeft       sdk.CodeType = 135
	CodeUnstakeWithin24Hours  sdk.CodeType = 136
	CodeUnstakeFail           sdk.CodeType = 137
	CodeEmptyChain            sdk.CodeType = 138
	CodeFailEventManager      sdk.CodeType = 139
)

var (
	notAuthorized          = fmt.Errorf("not authorized")
	errInvalidVersion      = fmt.Errorf("bad version")
	errBadVersion          = sdk.NewError(DefaultCodespace, CodeBadVersion, errInvalidVersion.Error())
	errInvalidMessage      = sdk.NewError(DefaultCodespace, CodeInvalidMessage, "invalid message")
	errConstNotAvailable   = sdk.NewError(DefaultCodespace, CodeConstantsNotAvailable, "constant values not available")
	errFailGetEventManager = sdk.NewError(DefaultCodespace, CodeFailEventManager, "fail to get event manager")
)

// NewExternalHandler returns a handler for "thorchain" type messages.
func NewExternalHandler(keeper Keeper,
	versionedTxOutStore VersionedTxOutStore,
	validatorMgr VersionedValidatorManager,
	versionedVaultManager VersionedVaultManager,
	versionedObserverManager VersionedObserverManager,
	versionedGasMgr VersionedGasManager,
	versionedEventManager VersionedEventManager) sdk.Handler {
	handlerMap := getHandlerMapping(keeper, versionedTxOutStore, validatorMgr, versionedVaultManager, versionedObserverManager, versionedGasMgr, versionedEventManager)

	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		version := keeper.GetLowestActiveVersion(ctx)
		constantValues := constants.GetConstantValues(version)
		if constantValues == nil {
			return errConstNotAvailable.Result()
		}
		h, ok := handlerMap[msg.Type()]
		if !ok {
			errMsg := fmt.Sprintf("Unrecognized thorchain Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
		result := h.Run(ctx, msg, version, constantValues)
		if len(ctx.EventManager().Events()) > 0 {
			result.Events = result.Events.AppendEvents(ctx.EventManager().Events())
		}
		return result
	}
}

func getHandlerMapping(keeper Keeper,
	versionedTxOutStore VersionedTxOutStore,
	validatorMgr VersionedValidatorManager,
	versionedVaultManager VersionedVaultManager,
	versionedObserverManager VersionedObserverManager,
	versionedGasMgr VersionedGasManager,
	versionedEventManager VersionedEventManager) map[string]MsgHandler {
	// New arch handlers
	m := make(map[string]MsgHandler)
	m[MsgTssPool{}.Type()] = NewTssHandler(keeper, versionedVaultManager)
	m[MsgSetNodeKeys{}.Type()] = NewSetNodeKeysHandler(keeper)
	m[MsgSetVersion{}.Type()] = NewVersionHandler(keeper)
	m[MsgSetIPAddress{}.Type()] = NewIPAddressHandler(keeper)
	m[MsgNativeTx{}.Type()] = NewNativeTxHandler(keeper, versionedObserverManager, versionedTxOutStore, validatorMgr, versionedVaultManager, versionedGasMgr, versionedEventManager)
	m[MsgObservedTxIn{}.Type()] = NewObservedTxInHandler(keeper, versionedObserverManager, versionedTxOutStore, validatorMgr, versionedVaultManager, versionedGasMgr, versionedEventManager)
	m[MsgObservedTxOut{}.Type()] = NewObservedTxOutHandler(keeper, versionedObserverManager, versionedTxOutStore, validatorMgr, versionedVaultManager, versionedGasMgr, versionedEventManager)
	m[MsgTssKeysignFail{}.Type()] = NewTssKeysignHandler(keeper)
	m[MsgErrataTx{}.Type()] = NewErrataTxHandler(keeper, versionedEventManager)
	m[MsgSend{}.Type()] = NewSendHandler(keeper)
	m[MsgMimir{}.Type()] = NewMimirHandler(keeper)
	return m
}

// NewInternalHandler returns a handler for "thorchain" internal type messages.
func NewInternalHandler(keeper Keeper,
	versionedTxOutStore VersionedTxOutStore,
	validatorMgr VersionedValidatorManager,
	versionedVaultManager VersionedVaultManager,
	versionedObserverManager VersionedObserverManager,
	versionedGasMgr VersionedGasManager,
	versionedEventManager VersionedEventManager) sdk.Handler {
	handlerMap := getInternalHandlerMapping(keeper, versionedTxOutStore, validatorMgr, versionedVaultManager, versionedObserverManager, versionedGasMgr, versionedEventManager)

	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		version := keeper.GetLowestActiveVersion(ctx)
		constantValues := constants.GetConstantValues(version)
		if constantValues == nil {
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

func getInternalHandlerMapping(keeper Keeper,
	versionedTxOutStore VersionedTxOutStore,
	validatorMgr VersionedValidatorManager,
	versionedVaultManager VersionedVaultManager,
	versionedObserverManager VersionedObserverManager,
	versionedGasMgr VersionedGasManager,
	versionedEventManager VersionedEventManager) map[string]MsgHandler {
	// New arch handlers
	m := make(map[string]MsgHandler)
	m[MsgOutboundTx{}.Type()] = NewOutboundTxHandler(keeper, versionedEventManager)
	m[MsgYggdrasil{}.Type()] = NewYggdrasilHandler(keeper, versionedTxOutStore, validatorMgr, versionedEventManager)
	m[MsgSwap{}.Type()] = NewSwapHandler(keeper, versionedTxOutStore, versionedEventManager)
	m[MsgReserveContributor{}.Type()] = NewReserveContributorHandler(keeper, versionedEventManager)
	m[MsgBond{}.Type()] = NewBondHandler(keeper, versionedEventManager)
	m[MsgLeave{}.Type()] = NewLeaveHandler(keeper, validatorMgr, versionedTxOutStore, versionedEventManager)
	m[MsgAdd{}.Type()] = NewAddHandler(keeper, versionedEventManager)
	m[MsgSetUnStake{}.Type()] = NewUnstakeHandler(keeper, versionedTxOutStore, versionedEventManager)
	m[MsgSetStakeData{}.Type()] = NewStakeHandler(keeper, versionedEventManager)
	m[MsgRefundTx{}.Type()] = NewRefundHandler(keeper, versionedEventManager)
	m[MsgMigrate{}.Type()] = NewMigrateHandler(keeper, versionedEventManager)
	m[MsgRagnarok{}.Type()] = NewRagnarokHandler(keeper, versionedEventManager)
	m[MsgSwitch{}.Type()] = NewSwitchHandler(keeper, versionedTxOutStore)
	return m
}

func fetchMemo(ctx sdk.Context, constAccessor constants.ConstantValues, keeper Keeper, tx common.Tx) string {
	if len(tx.Memo) > 0 {
		return tx.Memo
	}

	var memo string
	// attempt to pull memo from tx marker
	hash := tx.Hash()
	marks, _ := keeper.ListTxMarker(ctx, hash) // ignore err
	if len(marks) > 0 {
		// filter out expired tx markers
		period := constAccessor.GetInt64Value(constants.SigningTransactionPeriod) * 3
		marks = marks.FilterByMinHeight(ctx.BlockHeight() - period)

		// if we still have a marker, add the memo
		if len(marks) > 0 {
			var mark TxMarker
			mark, marks = marks.Pop()
			memo = mark.Memo
		}

		// update our marker list
		if err := keeper.SetTxMarkers(ctx, hash, marks); err != nil {
			ctx.Logger().Error("fail to set tx markers", "error", err)
		}
	}
	return memo
}

func processOneTxIn(ctx sdk.Context, keeper Keeper, tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, sdk.Error) {
	if len(tx.Tx.Coins) == 0 {
		return nil, sdk.ErrUnknownRequest("no coin found")
	}

	memo, err := ParseMemo(tx.Tx.Memo)
	if err != nil {
		ctx.Logger().Error("fail to parse memo", "error", err)
		return nil, sdk.NewError(DefaultCodespace, CodeInvalidMemo, err.Error())
	}
	// THORNode should not have one tx across chain, if it is cross chain it should be separate tx
	var newMsg sdk.Msg
	// interpret the memo and initialize a corresponding msg event
	switch m := memo.(type) {
	case StakeMemo:
		newMsg, err = getMsgStakeFromMemo(ctx, m, tx, signer)
		if err != nil {
			return nil, sdk.NewError(DefaultCodespace, CodeInvalidMemo, "invalid stake memo:%s", err.Error())
		}

	case UnstakeMemo:
		newMsg, err = getMsgUnstakeFromMemo(m, tx, signer)
		if err != nil {
			return nil, sdk.NewError(DefaultCodespace, CodeInvalidMemo, "invalid withdraw memo:%s", err.Error())
		}
	case SwapMemo:
		newMsg, err = getMsgSwapFromMemo(m, tx, signer)
		if err != nil {
			return nil, sdk.NewError(DefaultCodespace, CodeInvalidMemo, "invalid swap memo:%s", err.Error())
		}
	case AddMemo:
		newMsg, err = getMsgAddFromMemo(m, tx, signer)
		if err != nil {
			return nil, sdk.NewError(DefaultCodespace, CodeInvalidMemo, "invalid add memo:%s", err.Error())
		}
	case GasMemo:
		newMsg, err = getMsgNoOpFromMemo(tx, signer)
		if err != nil {
			return nil, sdk.NewError(DefaultCodespace, CodeInvalidMemo, "invalid noop memo:%s", err.Error())
		}
	case RefundMemo:
		newMsg, err = getMsgRefundFromMemo(m, tx, signer)
		if err != nil {
			return nil, sdk.NewError(DefaultCodespace, CodeInvalidMemo, "invalid refund memo:%s", err.Error())
		}
	case OutboundMemo:
		newMsg, err = getMsgOutboundFromMemo(m, tx, signer)
		if err != nil {
			return nil, sdk.NewError(DefaultCodespace, CodeInvalidMemo, "invalid outbound memo:%s", err.Error())
		}
	case MigrateMemo:
		newMsg, err = getMsgMigrateFromMemo(m, tx, signer)
		if err != nil {
			return nil, sdk.NewError(DefaultCodespace, CodeInvalidMemo, "invalid migrate memo: %s", err.Error())
		}
	case BondMemo:
		newMsg, err = getMsgBondFromMemo(m, tx, signer)
		if err != nil {
			return nil, sdk.NewError(DefaultCodespace, CodeInvalidMemo, "invalid bond memo:%s", err.Error())
		}
	case RagnarokMemo:
		newMsg, err = getMsgRagnarokFromMemo(m, tx, signer)
		if err != nil {
			return nil, sdk.NewError(DefaultCodespace, CodeInvalidMemo, "invalid ragnarok memo: %s", err.Error())
		}
	case LeaveMemo:
		newMsg = NewMsgLeave(tx.Tx, signer)
	case YggdrasilFundMemo:
		newMsg = NewMsgYggdrasil(tx.Tx, tx.ObservedPubKey, m.GetBlockHeight(), true, tx.Tx.Coins, signer)
	case YggdrasilReturnMemo:
		newMsg = NewMsgYggdrasil(tx.Tx, tx.ObservedPubKey, m.GetBlockHeight(), false, tx.Tx.Coins, signer)
	case ReserveMemo:
		res := NewReserveContributor(tx.Tx.FromAddress, tx.Tx.Coins[0].Amount)
		newMsg = NewMsgReserveContributor(tx.Tx, res, signer)
	case SwitchMemo:
		newMsg = NewMsgSwitch(tx.Tx, memo.GetDestination(), signer)

	default:
		return nil, sdk.NewError(DefaultCodespace, CodeInvalidMemo, "invalid memo")
	}

	if err := newMsg.ValidateBasic(); err != nil {
		return nil, sdk.NewError(DefaultCodespace, CodeInvalidMemo, "invalid message:%s", err.Error())
	}
	return newMsg, nil
}

func getMsgNoOpFromMemo(tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	for _, coin := range tx.Tx.Coins {
		if !coin.Asset.IsBNB() {
			return nil, errors.New("Only accepts BNB coins")
		}
	}
	return NewMsgNoOp(tx, signer), nil
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
		return nil, fmt.Errorf("swap from %s to %s is noop, refund", memo.Asset.String(), coin.Asset.String())
	}

	// Looks like at the moment THORNode can only process ont ty
	return NewMsgSwap(tx.Tx, memo.GetAsset(), memo.Destination, memo.SlipLimit, signer), nil
}

func getMsgUnstakeFromMemo(memo UnstakeMemo, tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	withdrawAmount := sdk.NewUint(MaxUnstakeBasisPoints)
	if len(memo.GetAmount()) > 0 {
		withdrawAmount = sdk.NewUintFromString(memo.GetAmount())
	}
	return NewMsgSetUnStake(tx.Tx, tx.Tx.FromAddress, withdrawAmount, memo.GetAsset(), signer), nil
}

func getMsgStakeFromMemo(ctx sdk.Context, memo StakeMemo, tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	// when staker stake to a pool ,usually it will be two coins, RUNE and the asset of the pool.
	// if it is multi-chain , like NOT Binance chain , it is using two asymmetric staking
	if len(tx.Tx.Coins) > 2 {
		return nil, errors.New("not expecting more than two coins in a stake")
	}
	runeAmount := sdk.ZeroUint()
	assetAmount := sdk.ZeroUint()
	asset := memo.GetAsset()
	if asset.IsEmpty() {
		return nil, errors.New("Unable to determine the intended pool for this stake")
	}
	// There is no dedicate pool for RUNE ,because every pool will have RUNE , that's by design
	if asset.IsRune() {
		return nil, errors.New("invalid pool asset")
	}
	// Extract the Rune amount and the asset amount from the transaction. At least one of them must be
	// nonzero. If we saw two types of coins, one of them must be the asset coin.
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
		return nil, fmt.Errorf("did not find %s ", asset)
	}

	runeAddr := tx.Tx.FromAddress
	assetAddr := memo.GetDestination()
	// this is to cover multi-chain scenario, for example BTC , staker who would like to stake in BTC pool,  will have to complete
	// the stake operation by sending in two asymmetric stake tx, one tx on BTC chain with memo stake:BTC:<RUNE address> ,
	// and another one on Binance chain with stake:BTC , with only RUNE as the coin
	// Thorchain will use the <RUNE address> to match these two together , and consider it as one stake.
	if !runeAddr.IsChain(common.RuneAsset().Chain) {
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

func getMsgMigrateFromMemo(memo MigrateMemo, tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	return NewMsgMigrate(tx, memo.GetBlockHeight(), signer), nil
}

func getMsgRagnarokFromMemo(memo RagnarokMemo, tx ObservedTx, signer sdk.AccAddress) (sdk.Msg, error) {
	return NewMsgRagnarok(tx, memo.GetBlockHeight(), signer), nil
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
	return NewMsgBond(tx.Tx, memo.GetAccAddress(), runeAmount, tx.Tx.FromAddress, signer), nil
}
