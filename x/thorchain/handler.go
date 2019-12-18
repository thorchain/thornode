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
)

// EmptyAccAddress empty address
var EmptyAccAddress = sdk.AccAddress{}
var notAuthorized = fmt.Errorf("not authorized")
var badVersion = fmt.Errorf("bad version")
var errBadVersion = sdk.NewError(DefaultCodespace, CodeBadVersion, "bad version")
var errInvalidMessage = sdk.NewError(DefaultCodespace, CodeInvalidMessage, "invalid message")
var errConstNotAvailable = sdk.NewError(DefaultCodespace, CodeConstantsNotAvailable, "constant values not available")

// NewHandler returns a handler for "thorchain" type messages.
func NewHandler(keeper Keeper, poolAddrMgr PoolAddressManager, txOutStore TxOutStore, validatorMgr ValidatorManager) sdk.Handler {
	// Classic Handler
	classic := NewClassicHandler(keeper, poolAddrMgr, txOutStore, validatorMgr)
	handlerMap := getHandlerMapping(keeper, poolAddrMgr, txOutStore, validatorMgr)

	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		version := keeper.GetLowestActiveVersion(ctx)
		constantValues := constants.GetConstantValues(version)
		if nil == constantValues {
			return errConstNotAvailable.Result()
		}
		h, ok := handlerMap[msg.Type()]
		if !ok {
			return classic(ctx, msg)
		}
		return h.Run(ctx, msg, version, constantValues)
	}
}

func getHandlerMapping(keeper Keeper, poolAddrMgr PoolAddressManager, txOutStore TxOutStore, validatorMgr ValidatorManager) map[string]MsgHandler {
	// New arch handlers
	m := make(map[string]MsgHandler)
	m[MsgTssPool{}.Type()] = NewTssHandler(keeper, txOutStore, poolAddrMgr)
	m[MsgNoOp{}.Type()] = NewNoOpHandler(keeper)
	m[MsgYggdrasil{}.Type()] = NewYggdrasilHandler(keeper, txOutStore, poolAddrMgr, validatorMgr)
	m[MsgEndPool{}.Type()] = NewEndPoolHandler(keeper, txOutStore, poolAddrMgr)
	m[MsgSetTrustAccount{}.Type()] = NewSetTrustAccountHandler(keeper)
	m[MsgSetAdminConfig{}.Type()] = NewSetAdminConfigHandler(keeper)
	m[MsgSwap{}.Type()] = NewSwapHandler(keeper, txOutStore, poolAddrMgr)
	m[MsgReserveContributor{}.Type()] = NewReserveContributorHandler(keeper)
	m[MsgSetPoolData{}.Type()] = NewPoolDataHandler(keeper)
	m[MsgSetVersion{}.Type()] = NewVersionHandler(keeper)
	m[MsgBond{}.Type()] = NewBondHandler(keeper)
	m[MsgObservedTxIn{}.Type()] = NewObservedTxInHandler(keeper, txOutStore, poolAddrMgr, validatorMgr)
	m[MsgObservedTxOut{}.Type()] = NewObservedTxOutHandler(keeper, txOutStore, poolAddrMgr, validatorMgr)
	m[MsgLeave{}.Type()] = NewLeaveHandler(keeper, validatorMgr, poolAddrMgr, txOutStore)
	m[MsgAdd{}.Type()] = NewAddHandler(keeper)
	m[MsgSetUnStake{}.Type()] = NewUnstakeHandler(keeper, txOutStore, poolAddrMgr)
	m[MsgSetStakeData{}.Type()] = NewStakeHandler(keeper)
	return m
}

// NewClassicHandler returns a handler for "thorchain" type messages.
func NewClassicHandler(keeper Keeper, poolAddressMgr PoolAddressManager, txOutStore TxOutStore, validatorManager ValidatorManager) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		version := keeper.GetLowestActiveVersion(ctx)
		constAccessor := constants.GetConstantValues(version)
		if nil == constAccessor {
			return errConstNotAvailable.Result()
		}
		switch m := msg.(type) {
		case MsgOutboundTx:
			return handleMsgOutboundTx(ctx, keeper, poolAddressMgr, m)
		default:
			errMsg := fmt.Sprintf("Unrecognized thorchain Msg type: %v", m)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
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

// handleMsgOutboundTx processes outbound tx from our pool
func handleMsgOutboundTx(ctx sdk.Context, keeper Keeper, poolAddressMgr PoolAddressManager, msg MsgOutboundTx) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgOutboundTx %s", msg.Tx.Tx.ID))
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "signer", msg.GetSigners())
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgOutboundTx", "error", err)
		return err.Result()
	}
	currentChainPoolAddr := poolAddressMgr.GetCurrentPoolAddresses().Current.GetByChain(msg.Tx.Tx.Chain)
	if nil == currentChainPoolAddr {
		msg := fmt.Sprintf("THORNode don't have pool for chain %s", msg.Tx.Tx.Chain)
		ctx.Logger().Error(msg)
		return sdk.ErrUnknownRequest(msg).Result()
	}

	currentPoolAddr, err := currentChainPoolAddr.GetAddress()
	if nil != err {
		ctx.Logger().Error("fail to get current pool address", "error", err)
		return sdk.ErrUnknownRequest("fail to get current pool address").Result()
	}
	previousChainPoolAddr := poolAddressMgr.GetCurrentPoolAddresses().Previous.GetByChain(msg.Tx.Tx.Chain)
	previousPoolAddr := common.NoAddress
	if nil != previousChainPoolAddr {
		previousPoolAddr, err = previousChainPoolAddr.GetAddress()
		if nil != err {
			ctx.Logger().Error("fail to get previous pool address", "error", err)
			return sdk.ErrUnknownRequest("fail to get previous pool address").Result()
		}
	}

	if !currentPoolAddr.Equals(msg.Tx.Tx.FromAddress) && !previousPoolAddr.Equals(msg.Tx.Tx.FromAddress) {
		ctx.Logger().Error("message sent by unauthorized account", "sender", msg.Tx.Tx.FromAddress.String(), "current pool addr", currentPoolAddr.String())
		return sdk.ErrUnauthorized("Not authorized").Result()
	}

	voter, err := keeper.GetObservedTxVoter(ctx, msg.InTxID)
	if err != nil {
		ctx.Logger().Error(err.Error())
		return sdk.ErrInternal("fail to get observed tx voter").Result()
	}
	voter.AddOutTx(msg.Tx.Tx)
	keeper.SetObservedTxVoter(ctx, voter)

	// complete events
	if voter.IsDone() {
		err := completeEvents(ctx, keeper, msg.InTxID, voter.OutTxs)
		if err != nil {
			ctx.Logger().Error("unable to complete events", "error", err)
			return sdk.ErrInternal(err.Error()).Result()
		}
	}

	// Apply Gas fees
	if err := AddGasFees(ctx, keeper, msg.Tx); nil != err {
		ctx.Logger().Error("fail to add gas fee", err)
		return sdk.ErrInternal("fail to add gas fee").Result()
	}

	// update txOut record with our TxID that sent funds out of the pool
	txOut, err := keeper.GetTxOut(ctx, uint64(voter.Height))
	if err != nil {
		ctx.Logger().Error("unable to get txOut record", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	// Save TxOut back with the TxID only when the TxOut on the block height is
	// not empty
	if !txOut.IsEmpty() {
		for i, tx := range txOut.TxArray {

			// withdraw , refund etc, one inbound tx might result two outbound txes, THORNode have to correlate outbound tx back to the
			// inbound, and also txitem , thus THORNode could record both outbound tx hash correctly
			// given every tx item will only have one coin in it , given that , THORNode could use that to identify which txit
			if tx.InHash.Equals(msg.InTxID) &&
				tx.OutHash.IsEmpty() &&
				msg.Tx.Tx.Coins.Contains(tx.Coin) {
				txOut.TxArray[i].OutHash = msg.Tx.Tx.ID
			}
		}
		if err := keeper.SetTxOut(ctx, txOut); nil != err {
			ctx.Logger().Error("fail to save tx out", err)
			return sdk.ErrInternal("fail to save tx out").Result()
		}
	}
	keeper.SetLastSignedHeight(ctx, sdk.NewUint(uint64(voter.Height)))

	// If THORNode are sending from a yggdrasil pool, decrement coins on record
	if keeper.VaultExists(ctx, msg.Tx.ObservedPubKey) {
		ygg, err := keeper.GetVault(ctx, msg.Tx.ObservedPubKey)
		if nil != err {
			ctx.Logger().Error("fail to get yggdrasil", err)
			return sdk.ErrInternal("fail to get yggdrasil").Result()
		}
		ygg.SubFunds(msg.Tx.Tx.Coins)
		if err := keeper.SetVault(ctx, ygg); nil != err {
			ctx.Logger().Error("fail to save yggdrasil", err)
			return sdk.ErrInternal("fail to save yggdrasil").Result()
		}
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
