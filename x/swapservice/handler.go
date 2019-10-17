package swapservice

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EmptyAccAddress empty address
var EmptyAccAddress = sdk.AccAddress{}

// NewHandler returns a handler for "swapservice" type messages.
func NewHandler(keeper Keeper, poolAddressMgr *PoolAddressManager, txOutStore *TxOutStore, validatorManager *ValidatorManager) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch m := msg.(type) {
		case MsgSetPoolData:
			return handleMsgSetPoolData(ctx, keeper, m)
		case MsgSetStakeData:
			return handleMsgSetStakeData(ctx, keeper, m)
		case MsgSwap:
			return handleMsgSwap(ctx, keeper, txOutStore, poolAddressMgr, m)
		case MsgAdd:
			return handleMsgAdd(ctx, keeper, m)
		case MsgSetUnStake:
			return handleMsgSetUnstake(ctx, keeper, txOutStore, poolAddressMgr, m)
		case MsgSetTxIn:
			return handleMsgSetTxIn(ctx, keeper, txOutStore, poolAddressMgr, validatorManager, m)
		case MsgSetAdminConfig:
			return handleMsgSetAdminConfig(ctx, keeper, m)
		case MsgOutboundTx:
			return handleMsgOutboundTx(ctx, keeper, poolAddressMgr, m)
		case MsgNoOp:
			return handleMsgNoOp(ctx, keeper, m)
		case MsgEndPool:
			return handleOperatorMsgEndPool(ctx, keeper, txOutStore, poolAddressMgr, m)
		case MsgSetTrustAccount:
			return handleMsgSetTrustAccount(ctx, keeper, m)
		case MsgApply:
			return handleMsgApply(ctx, keeper, m)
		case MsgNextPoolAddress:
			return handleMsgConfirmNextPoolAddress(ctx, keeper, validatorManager, m)
		case MsgLeave:
			return handleMsgLeave(ctx, keeper, txOutStore, poolAddressMgr, m)
		default:
			errMsg := fmt.Sprintf("Unrecognized swapservice Msg type: %v", m)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// isSignedByActiveObserver check whether the signers are all active observer
func isSignedByActiveObserver(ctx sdk.Context, keeper Keeper, signers []sdk.AccAddress) bool {
	if len(signers) == 0 {
		return false
	}
	for _, signer := range signers {
		if !keeper.IsActiveObserver(ctx, signer) {
			return false
		}
	}
	return true
}

func isSignedByActiveNodeAccounts(ctx sdk.Context, keeper Keeper, signers []sdk.AccAddress) bool {
	if len(signers) == 0 {
		return false
	}
	for _, signer := range signers {
		nodeAccount, err := keeper.GetNodeAccount(ctx, signer)
		if err != nil {
			ctx.Logger().Error("unauthorized account", "address", signer.String())
			return false
		}
		if nodeAccount.IsEmpty() {
			ctx.Logger().Error("unauthorized account", "address", signer.String())
			return false
		}
		if nodeAccount.Status != NodeActive {
			ctx.Logger().Error("unauthorized account, node account not active", "address", signer.String(), "status", nodeAccount.Status)
			return false
		}
	}
	return true
}

// handleOperatorMsgEndPool operators decide it is time to end the pool
func handleOperatorMsgEndPool(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore, poolAddrMgr *PoolAddressManager, msg MsgEndPool) sdk.Result {
	if !isSignedByActiveNodeAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "ticker", msg.Ticker)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	ctx.Logger().Info("handle MsgEndPool", "ticker", msg.Ticker, "requester", msg.Requester, "signer", msg.Signer.String())
	poolStaker, err := keeper.GetPoolStaker(ctx, msg.Ticker)
	if nil != err {
		ctx.Logger().Error("fail to get pool staker", err)
		return sdk.ErrInternal(err.Error()).Result()
	}
	// everyone withdraw
	for _, item := range poolStaker.Stakers {
		unstakeMsg := NewMsgSetUnStake(
			item.StakerID,
			sdk.NewUint(10000),
			msg.Ticker,
			msg.RequestTxHash,
			msg.Signer,
		)

		result := handleMsgSetUnstake(ctx, keeper, txOutStore, poolAddrMgr, unstakeMsg)
		if !result.IsOK() {
			ctx.Logger().Error("fail to unstake", "staker", item.StakerID)
			return result
		}
	}
	keeper.SetPoolData(
		ctx,
		msg.Ticker,
		PoolSuspended)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// Handle a message to set pooldata
func handleMsgSetPoolData(ctx sdk.Context, keeper Keeper, msg MsgSetPoolData) sdk.Result {
	if !isSignedByActiveNodeAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "ticker", msg.Ticker)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	ctx.Logger().Info("handleMsgSetPoolData request", "Ticker:"+msg.Ticker)
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	keeper.SetPoolData(
		ctx,
		msg.Ticker,
		msg.Status)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func processStakeEvent(ctx sdk.Context, keeper Keeper, msg MsgSetStakeData, stakeUnits sdk.Uint, eventStatus EventStatus) error {
	var stakeEvt EventStake
	if eventStatus == EventRefund {
		// do not log event if the stake failed
		return nil
	}

	stakeEvt = NewEventStake(
		msg.RuneAmount,
		msg.TokenAmount,
		stakeUnits,
	)
	stakeBytes, err := json.Marshal(stakeEvt)
	if err != nil {
		ctx.Logger().Error("fail to save event", err)
		return errors.Wrap(err, "fail to marshal stake event to json")
	}

	evt := NewEvent(
		stakeEvt.Type(),
		msg.RequestTxHash,
		msg.Ticker,
		stakeBytes,
		eventStatus,
	)
	keeper.AddIncompleteEvents(ctx, evt)
	if eventStatus != EventRefund {
		// since there is no outbound tx for staking, we'll complete the event now
		blankTxID, _ := common.NewTxID(
			"0000000000000000000000000000000000000000000000000000000000000000",
		)
		keeper.CompleteEvents(ctx, []common.TxID{msg.RequestTxHash}, blankTxID)
	}
	return nil
}

// Handle a message to set stake data
func handleMsgSetStakeData(ctx sdk.Context, keeper Keeper, msg MsgSetStakeData) (result sdk.Result) {
	stakeUnits := sdk.ZeroUint()
	defer func() {
		var status EventStatus
		if result.IsOK() {
			status = EventSuccess
		} else {
			status = EventRefund
		}
		if err := processStakeEvent(ctx, keeper, msg, stakeUnits, status); nil != err {
			ctx.Logger().Error("fail to save stake event", "error", err)
			result = sdk.ErrInternal("fail to save stake event").Result()
		}
	}()

	ctx.Logger().Info("handleMsgSetStakeData request", "stakerid:"+msg.Ticker)
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "ticker", msg.Ticker, "request tx hash", msg.RequestTxHash, "public address", msg.PublicAddress)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	pool := keeper.GetPool(ctx, msg.Ticker)
	if pool.Empty() {
		ctx.Logger().Info("pool doesn't exist yet, create a new one", "symbol", msg.Ticker, "creator", msg.PublicAddress)
		pool.Ticker = msg.Ticker
		keeper.SetPool(ctx, pool)
	}
	if err := pool.EnsureValidPoolStatus(msg); nil != err {
		ctx.Logger().Error("check pool status", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	stakeUnits, err := stake(
		ctx,
		keeper,
		msg.Ticker,
		msg.RuneAmount,
		msg.TokenAmount,
		msg.PublicAddress,
		msg.RequestTxHash,
	)
	if err != nil {
		ctx.Logger().Error("fail to process stake message", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// Handle a message to set stake data
func handleMsgSwap(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore, poolAddrMgr *PoolAddressManager, msg MsgSwap) sdk.Result {
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.RequestTxHash, "source ticker", msg.SourceTicker, "target ticker", msg.TargetTicker)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}

	gsl := keeper.GetAdminConfigGSL(ctx, EmptyAccAddress)

	amount, err := swap(
		ctx,
		keeper,
		msg.RequestTxHash,
		msg.SourceTicker,
		msg.TargetTicker,
		msg.Amount,
		msg.Requester,
		msg.Destination,
		msg.RequestTxHash,
		msg.TargetPrice,
		gsl,
	) // If so, set the stake data to the value specified in the msg.
	if err != nil {
		ctx.Logger().Error("fail to process swap message", "error", err)

		return sdk.ErrInternal(err.Error()).Result()
	}

	res, err := keeper.cdc.MarshalBinaryLengthPrefixed(struct {
		Token sdk.Uint `json:"token"`
	}{
		Token: amount,
	})
	if nil != err {
		ctx.Logger().Error("fail to encode result to json", "error", err)
		return sdk.ErrInternal("fail to encode result to json").Result()
	}

	toi := &TxOutItem{
		PoolAddress: poolAddrMgr.GetCurrentPoolAddresses().Current,
		ToAddress:   msg.Destination,
	}
	toi.Coins = append(toi.Coins, common.Coin{
		Denom:  msg.TargetTicker,
		Amount: amount,
	})
	txOutStore.AddTxOutItem(ctx, keeper, toi)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      res,
		Codespace: DefaultCodespace,
	}
}

// handleMsgSetUnstake process unstake
func handleMsgSetUnstake(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore, poolAddrMgr *PoolAddressManager, msg MsgSetUnStake) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgSetUnstake from : %s(%s) unstake (%s)", msg, msg.PublicAddress, msg.WithdrawBasisPoints))
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.RequestTxHash, "public address", msg.PublicAddress, "ticker", msg.Ticker, "withdraw basis points", msg.WithdrawBasisPoints)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := keeper.GetPool(ctx, msg.Ticker).EnsureValidPoolStatus(msg); nil != err {
		ctx.Logger().Error("check pool status", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgSetUnstake", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	runeAmt, tokenAmount, units, err := unstake(ctx, keeper, msg)
	if nil != err {
		ctx.Logger().Error("fail to UnStake", "error", err)
		return sdk.ErrInternal("fail to process UnStake request").Result()
	}
	res, err := keeper.cdc.MarshalBinaryLengthPrefixed(struct {
		Rune  sdk.Uint `json:"rune"`
		Token sdk.Uint `json:"token"`
	}{
		Rune:  runeAmt,
		Token: tokenAmount,
	})
	if nil != err {
		ctx.Logger().Error("fail to marshal result to json", "error", err)
		// if this happen what should we tell the client?
	}

	unstakeEvt := NewEventUnstake(
		runeAmt,
		tokenAmount,
		units,
	)
	unstakeBytes, err := json.Marshal(unstakeEvt)
	if err != nil {
		ctx.Logger().Error("fail to save event", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	evt := NewEvent(
		unstakeEvt.Type(),
		msg.RequestTxHash,
		msg.Ticker,
		unstakeBytes,
		EventSuccess,
	)
	keeper.AddIncompleteEvents(ctx, evt)

	toi := &TxOutItem{
		PoolAddress: poolAddrMgr.currentPoolAddresses.Current,
		ToAddress:   msg.PublicAddress,
	}
	toi.Coins = append(toi.Coins, common.Coin{
		Denom:  common.RuneTicker,
		Amount: runeAmt,
	})
	toi.Coins = append(toi.Coins, common.Coin{
		Denom:  msg.Ticker,
		Amount: tokenAmount,
	})
	txOutStore.AddTxOutItem(ctx, keeper, toi)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      res,
		Codespace: DefaultCodespace,
	}
}

func refundTx(ctx sdk.Context, tx TxIn, store *TxOutStore, keeper Keeper, poolAddrMgr *PoolAddressManager) {
	toi := &TxOutItem{
		ToAddress:   tx.Sender,
		PoolAddress: poolAddrMgr.GetCurrentPoolAddresses().Current,
		Coins:       tx.Coins,
	}

	store.AddTxOutItem(ctx, keeper, toi)
}

// handleMsgConfirmNextPoolAddress , this is the method to handle MsgNextPoolAddress
// MsgNextPoolAddress is a way to prove that the operator has access to the address, and can sign transaction with the given address on binance chain
func handleMsgConfirmNextPoolAddress(ctx sdk.Context, keeper Keeper, validatorManager *ValidatorManager, msg MsgNextPoolAddress) sdk.Result {
	ctx.Logger().Info("receive request to set next pool address", "sender", msg.Sender.String())
	if validatorManager.Meta.Nominated.IsEmpty() {
		return sdk.ErrUnknownRequest("no nominated node yet").Result()
	}

	nominated, err := keeper.GetNodeAccount(ctx, validatorManager.Meta.Nominated.NodeAddress)
	if err != nil {
		return sdk.ErrInternal(fmt.Sprintf("fail to get nominated node,err:%s", err.Error())).Result()
	}
	if !msg.Sender.Equals(nominated.Accounts.SignerBNBAddress) {
		return sdk.ErrUnknownRequest("nominated node has different signer bnb address").Result()
	}

	nominated.SignerActive = true
	if nominated.ObserverActive && nominated.SignerActive {
		// only update their status when both observer and signer are active
		nominated.UpdateStatus(NodeReady)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeNodeReady,
				sdk.NewAttribute("signer bnb address", nominated.Accounts.SignerBNBAddress.String()),
				sdk.NewAttribute("observer bep address", nominated.Accounts.ObserverBEPAddress.String()),
				sdk.NewAttribute("bep consensus pub key", nominated.Accounts.ValidatorBEPConsPubKey)))
	}

	keeper.SetNodeAccount(ctx, nominated)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgSetTxIn gets a binance tx hash, gets the tx/memo, and triggers
// another handler to process the request
func handleMsgSetTxIn(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore, poolAddressMgr *PoolAddressManager, validatorManager *ValidatorManager, msg MsgSetTxIn) sdk.Result {
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		unAuthorizedResult := sdk.ErrUnauthorized("signer is not authorized").Result()
		na, err := keeper.GetNodeAccountByObserver(ctx, msg.Signer)
		if nil != err {
			ctx.Logger().Error("fail to get node account", err, "signer", msg.Signer.String())
			return unAuthorizedResult
		}
		if na.IsEmpty() {
			return unAuthorizedResult
		}

		if na.Status != NodeUnknown &&
			na.Status != NodeDisabled &&
			!na.ObserverActive {
			// tx observed by a standby node, let's mark their observer as active
			na.ObserverActive = true
			keeper.SetNodeAccount(ctx, na)
			return sdk.Result{
				Code:      sdk.CodeOK,
				Codespace: DefaultCodespace,
			}
		}
		return unAuthorizedResult

	}
	activeNodeAccounts, err := keeper.ListActiveNodeAccounts(ctx)
	if nil != err {
		ctx.Logger().Error("fail to get list of active node accounts", err)
		return sdk.ErrInternal("fail to get list of active node accounts").Result()
	}
	handler := NewHandler(keeper, poolAddressMgr, txOutStore, validatorManager)
	for _, tx := range msg.TxIns {
		voter := keeper.GetTxInVoter(ctx, tx.TxID)
		preConsensus := voter.HasConensus(activeNodeAccounts)
		voter.Adds(tx.Txs, msg.Signer)
		postConsensus := voter.HasConensus(activeNodeAccounts)
		keeper.SetTxInVoter(ctx, voter)

		if preConsensus == false && postConsensus == true && !voter.IsProcessed {
			voter.IsProcessed = true
			keeper.SetTxInVoter(ctx, voter)
			txIn := voter.GetTx(activeNodeAccounts)
			m, err := processOneTxIn(ctx, keeper, tx.TxID, txIn, msg.Signer, poolAddressMgr)
			if nil != err {
				ctx.Logger().Error("fail to process txHash", "error", err)
				refundTx(ctx, voter.GetTx(activeNodeAccounts), txOutStore, keeper, poolAddressMgr)
				ee := NewEmptyRefundEvent()
				buf, err := json.Marshal(ee)
				if nil != err {
					return sdk.ErrInternal("fail to marshal EmptyRefund event to json").Result()
				}
				event := NewEvent(ee.Type(), tx.TxID, common.Ticker(""), buf, EventRefund)
				keeper.AddIncompleteEvents(ctx, event)
				continue
			}

			// ignoring the error
			_ = keeper.AddToTxInIndex(ctx, uint64(ctx.BlockHeight()), tx.TxID)
			if err := keeper.SetLastBinanceHeight(ctx, txIn.BlockHeight); nil != err {
				return sdk.ErrInternal("fail to save last binance height to data store err:" + err.Error()).Result()
			}

			result := handler(ctx, m)
			if !result.IsOK() {
				refundTx(ctx, voter.GetTx(activeNodeAccounts), txOutStore, keeper, poolAddressMgr)
			}
		}
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func processOneTxIn(ctx sdk.Context, keeper Keeper, txID common.TxID, tx TxIn, signer sdk.AccAddress, poolAddrMgr *PoolAddressManager) (sdk.Msg, error) {
	if !poolAddrMgr.GetCurrentPoolAddresses().Current.Equals(tx.ObservePoolAddress) {
		return nil, errors.New("tx sent to the wrong pool address")
	}
	memo, err := ParseMemo(tx.Memo)
	if err != nil {
		return nil, errors.Wrap(err, "fail to parse memo")
	}

	var newMsg sdk.Msg
	// interpret the memo and initialize a corresponding msg event
	switch m := memo.(type) {
	case CreateMemo:
		newMsg, err = getMsgSetPoolDataFromMemo(ctx, keeper, m, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgSetPoolData from memo")
		}

	case StakeMemo:
		newMsg, err = getMsgStakeFromMemo(ctx, m, txID, &tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgStake from memo")
		}

	case WithdrawMemo:
		newMsg, err = getMsgUnstakeFromMemo(m, txID, tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgUnstake from memo")
		}
	case SwapMemo:
		newMsg, err = getMsgSwapFromMemo(m, txID, tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgSwap from memo")
		}
	case AddMemo:
		newMsg, err = getMsgAddFromMemo(m, txID, tx, signer)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get MsgAdd from memo")
		}
	case GasMemo:
		newMsg, err = getMsgNoOpFromMemo(tx, signer)
		if err != nil {
			return nil, errors.Wrap(err, "fail to get MsgNoOp from memo")
		}
	case OutboundMemo:
		newMsg, err = getMsgOutboundFromMemo(m, txID, tx.Sender, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgOutbound from memo")
		}
	case ApplyMemo:
		newMsg, err = getMsgApplyFromMemo(m, txID, tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgApply from memo")
		}
	case NextPoolMemo:
		newMsg = NewMsgNextPoolAddress(txID, tx.Sender, signer)
	case LeaveMemo:
		newMsg = NewMsgLeave(m.GetDestination(), txID, tx.Sender, signer)
	default:
		return nil, errors.Wrap(err, "Unable to find memo type")
	}

	if err := newMsg.ValidateBasic(); nil != err {
		return nil, errors.Wrap(err, "invalid msg")
	}
	return newMsg, nil
}

func getMsgNoOpFromMemo(tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	for _, coin := range tx.Coins {
		if !common.IsBNB(coin.Denom) {
			return nil, errors.New("Only accepts BNB coins")
		}
	}
	return NewMsgNoOp(signer), nil
}

func getMsgSwapFromMemo(memo SwapMemo, txID common.TxID, tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	if len(tx.Coins) > 1 {
		return nil, errors.New("not expecting multiple coins in a swap")
	}
	if memo.Destination.IsEmpty() {
		memo.Destination = tx.Sender
	}
	coin := tx.Coins[0]
	// Looks like at the moment we can only process ont ty
	return NewMsgSwap(txID, coin.Denom, memo.GetTicker(), coin.Amount, tx.Sender, memo.Destination, memo.SlipLimit, signer), nil
}

func getMsgUnstakeFromMemo(memo WithdrawMemo, txID common.TxID, tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	withdrawAmount := sdk.NewUint(MaxWithdrawBasisPoints)
	if len(memo.GetAmount()) > 0 {
		withdrawAmount = sdk.NewUintFromString(memo.GetAmount())
	}
	return NewMsgSetUnStake(tx.Sender, withdrawAmount, memo.GetTicker(), txID, signer), nil

}

func getMsgStakeFromMemo(ctx sdk.Context, memo StakeMemo, txID common.TxID, tx *TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	if len(tx.Coins) > 2 {
		return nil, errors.New("not expecting more than two coins in a stake")
	}
	runeAmount := sdk.ZeroUint()
	tokenAmount := sdk.ZeroUint()
	ticker := memo.GetTicker()
	for _, coin := range tx.Coins {
		ctx.Logger().Info("coin", "denom", coin.Denom.String(), "amount", coin.Amount.String())
		if common.IsRune(coin.Denom) {
			runeAmount = coin.Amount
		} else {
			tokenAmount = coin.Amount
			ticker = coin.Denom // override the memo ticker with coin received
		}
	}
	if ticker.IsEmpty() {
		return nil, errors.New("Unable to determine the intended pool for this stake")
	}
	return NewMsgSetStakeData(
		ticker,
		runeAmount,
		tokenAmount,
		tx.Sender,
		txID,
		signer,
	), nil
}

func getMsgSetPoolDataFromMemo(ctx sdk.Context, keeper Keeper, memo CreateMemo, signer sdk.AccAddress) (sdk.Msg, error) {
	if keeper.PoolExist(ctx, memo.GetTicker()) {
		return nil, errors.New("pool already exists")
	}
	return NewMsgSetPoolData(
		memo.GetTicker(),
		PoolEnabled, // new pools start in a Bootstrap state
		signer,
	), nil
}

func getMsgAddFromMemo(memo AddMemo, txID common.TxID, tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	runeAmount := sdk.ZeroUint()
	tokenAmount := sdk.ZeroUint()
	for _, coin := range tx.Coins {
		if common.IsRune(coin.Denom) {
			runeAmount = coin.Amount
		} else if memo.GetTicker().Equals(coin.Denom) {
			tokenAmount = coin.Amount
		}
	}
	return NewMsgAdd(
		memo.GetTicker(),
		runeAmount,
		tokenAmount,
		txID,
		signer,
	), nil
}

func getMsgOutboundFromMemo(memo OutboundMemo, txID common.TxID, sender common.BnbAddress, signer sdk.AccAddress) (sdk.Msg, error) {
	blockHeight := memo.GetBlockHeight()
	return NewMsgOutboundTx(
		txID,
		blockHeight,
		sender,
		signer,
	), nil
}
func getMsgApplyFromMemo(memo ApplyMemo, txID common.TxID, tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	runeAmount := sdk.ZeroUint()
	for _, coin := range tx.Coins {
		if common.IsRune(coin.Denom) {
			runeAmount = coin.Amount
		}
	}
	if runeAmount.IsZero() {
		return nil, errors.New("RUNE amount is 0")
	}
	// later on , we might be able to automatically do a swap for them , but not right now

	return NewMsgApply(memo.GetNodeAddress(), runeAmount, txID, signer), nil
}

// handleMsgAdd
func handleMsgAdd(ctx sdk.Context, keeper Keeper, msg MsgAdd) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgAdd %s", msg.TxID))
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account")
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgAdd", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	pool := keeper.GetPool(ctx, msg.Ticker)
	if pool.Ticker.IsEmpty() {
		return sdk.ErrUnknownRequest(fmt.Sprintf("pool %s not exist", msg.Ticker)).Result()
	}
	if msg.TokenAmount.GT(sdk.ZeroUint()) {
		pool.BalanceToken = pool.BalanceToken.Add(msg.TokenAmount)
	}
	if msg.RuneAmount.GT(sdk.ZeroUint()) {
		pool.BalanceRune = pool.BalanceRune.Add(msg.RuneAmount)
	}

	keeper.SetPool(ctx, pool)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgNoOp doesn't do anything, its a no op
func handleMsgNoOp(ctx sdk.Context, keeper Keeper, msg MsgNoOp) sdk.Result {
	ctx.Logger().Info("receive no op msg")
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgOutboundTx processes outbound tx from our pool
func handleMsgOutboundTx(ctx sdk.Context, keeper Keeper, poolAddressMgr *PoolAddressManager, msg MsgOutboundTx) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgOutboundTx %s at height %d", msg.TxID, msg.Height))
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "signer", msg.GetSigners())
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgOutboundTx", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	// it could
	currentPoolAddr := poolAddressMgr.GetCurrentPoolAddresses()
	if !currentPoolAddr.Current.Equals(msg.Sender) && !currentPoolAddr.Previous.Equals(msg.Sender) {
		ctx.Logger().Error("message sent by unauthorized account")
		return sdk.ErrUnauthorized("Not authorized").Result()
	}

	index, err := keeper.GetTxInIndex(ctx, msg.Height)
	if err != nil {
		ctx.Logger().Error("invalid TxIn Index", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	// iterate over our index and mark each tx as done
	for _, txID := range index {
		voter := keeper.GetTxInVoter(ctx, txID)
		voter.SetDone(msg.TxID)
		keeper.SetTxInVoter(ctx, voter)
	}

	// complete events
	keeper.CompleteEvents(ctx, index, msg.TxID)

	// update txOut record with our TxID that sent funds out of the pool
	txOut, err := keeper.GetTxOut(ctx, msg.Height)
	if err != nil {
		ctx.Logger().Error("unable to get txOut record", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	// Save TxOut back with the TxID only when the TxOut on the block height is not empty
	if !txOut.IsEmpty() {
		txOut.Hash = msg.TxID
		keeper.SetTxOut(ctx, txOut)
	}
	keeper.SetLastSignedHeight(ctx, sdk.NewUint(msg.Height))

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgSetAdminConfig process admin config
func handleMsgSetAdminConfig(ctx sdk.Context, keeper Keeper, msg MsgSetAdminConfig) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgSetAdminConfig %s --> %s", msg.AdminConfig.Key, msg.AdminConfig.Value))
	if !isSignedByActiveNodeAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account")
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgSetAdminConfig", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	keeper.SetAdminConfig(ctx, msg.AdminConfig)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgSetTrustAccount Update node account
func handleMsgSetTrustAccount(ctx sdk.Context, keeper Keeper, msg MsgSetTrustAccount) sdk.Result {
	ctx.Logger().Info("receive MsgSetTrustAccount", "trust account info", msg.TrustAccount.String())
	nodeAccount, err := keeper.GetNodeAccount(ctx, msg.Signer)
	if err != nil {
		ctx.Logger().Error("fail to get node account", "error", err, "address", msg.Signer.String())
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not authorizaed", msg.Signer)).Result()
	}
	if nodeAccount.IsEmpty() {
		ctx.Logger().Error("unauthorized account", "address", msg.Signer.String())
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not authorizaed", msg.Signer)).Result()
	}
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("MsgUpdateNodeAccount is invalid", "error", err)
		return sdk.ErrUnknownRequest("MsgUpdateNodeAccount is invalid").Result()
	}

	// You should not able to update node address when the node is in active mode
	// for example if they update observer address
	if nodeAccount.Status == NodeActive {
		ctx.Logger().Error(fmt.Sprintf("node %s is active, so it can't update itself", nodeAccount.NodeAddress))
		return sdk.ErrUnknownRequest("node is active can't update").Result()
	}
	if nodeAccount.Status == NodeDisabled {
		ctx.Logger().Error(fmt.Sprintf("node %s is disabled, so it can't update itself", nodeAccount.NodeAddress))
		return sdk.ErrUnknownRequest("node is disabled can't update").Result()
	}
	if err := keeper.EnsureTrustAccountUnique(ctx, msg.TrustAccount); nil != err {
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	// Here make sure we don't change the node account's bond
	nodeAccount.Accounts = msg.TrustAccount
	nodeAccount.UpdateStatus(NodeStandby)
	keeper.SetNodeAccount(ctx, nodeAccount)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent("set_trust_account",
			sdk.NewAttribute("bep_address", msg.Signer.String()),
			sdk.NewAttribute("observer_bep_address", msg.TrustAccount.ObserverBEPAddress.String()),
			sdk.NewAttribute("signer_bnb_address", msg.TrustAccount.SignerBNBAddress.String()),
			sdk.NewAttribute("validator_consensus_pub_key", msg.TrustAccount.ValidatorBEPConsPubKey)))
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgApply
func handleMsgApply(ctx sdk.Context, keeper Keeper, msg MsgApply) sdk.Result {
	ctx.Logger().Info("receive MsgApply", "node address", msg.NodeAddress, "txhash", msg.RequestTxHash, "bond", msg.Bond.String())
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "signer", msg.GetSigners())
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgApply", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	nodeAccount, err := keeper.GetNodeAccount(ctx, msg.NodeAddress)
	if nil != err {
		ctx.Logger().Error("fail to get node account", "err", err, "address", msg.NodeAddress)
		return sdk.ErrInternal("fail to get node account").Result()
	}
	if !nodeAccount.IsEmpty() {
		ctx.Logger().Error("node account already exist", "address", msg.NodeAddress, "status", nodeAccount.Status)
		return sdk.ErrUnknownRequest("node account already exist").Result()
	}
	minValidatorBond := keeper.GetAdminConfigMinValidatorBond(ctx, sdk.AccAddress{})
	if msg.Bond.LT(minValidatorBond) {
		ctx.Logger().Error("not enough rune to be whitelisted", "rune", msg.Bond, "min validator bond", minValidatorBond.String())
		return sdk.ErrUnknownRequest("not enough rune to be whitelisted").Result()
	}
	// we don't have the trust account info right now, so leave it empty
	trustAccount := NewTrustAccount(common.NoBnbAddress, sdk.AccAddress{}, "")
	// white list the given bep address
	nodeAccount = NewNodeAccount(msg.NodeAddress, NodeWhiteListed, trustAccount)
	nodeAccount.Bond = msg.Bond
	keeper.SetNodeAccount(ctx, nodeAccount)
	ctx.EventManager().EmitEvent(sdk.NewEvent("new_node", sdk.NewAttribute("address", msg.NodeAddress.String())))
	coinsToMint := keeper.GetAdminConfigWhiteListGasToken(ctx, sdk.AccAddress{})
	// mint some gas token
	err = keeper.supplyKeeper.MintCoins(ctx, ModuleName, coinsToMint)
	if nil != err {
		ctx.Logger().Error("fail to mint gas tokens", "err", err)
	}
	if err := keeper.supplyKeeper.SendCoinsFromModuleToAccount(ctx, ModuleName, msg.NodeAddress, coinsToMint); nil != err {
		ctx.Logger().Error("fail to send newly minted gas token to node address")
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func handleMsgLeave(ctx sdk.Context, keeper Keeper, txOut *TxOutStore, poolAddrMgr *PoolAddressManager, msg MsgLeave) sdk.Result {
	ctx.Logger().Info("receive MsgLeave", "sender", msg.Sender.String(), "request tx hash", msg.RequestTxHash, "destination", msg.Destination)
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "signer", msg.GetSigners())
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgLeave", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	nodeAcc, err := keeper.GetNodeAccountBySignerBNBAddress(ctx, msg.Sender)
	if nil != err {
		ctx.Logger().Error("fail to get node account", "error", err)
		return sdk.ErrInternal("fail to get node account by signer bnb address").Result()
	}
	if nodeAcc.IsEmpty() {
		return sdk.ErrUnknownRequest("node account doesn't exist").Result()
	}
	if nodeAcc.Status == NodeActive {
		return sdk.ErrUnknownRequest("active node can't leave").Result()
	}
	curPoolAddr := poolAddrMgr.GetCurrentPoolAddresses()

	if curPoolAddr.Current.Equals(msg.Sender) || curPoolAddr.Previous.Equals(msg.Sender) || curPoolAddr.Next.Equals(msg.Sender) {
		return sdk.ErrUnknownRequest("address still in use , cannot leave now").Result()
	}
	if nodeAcc.Bond.GT(sdk.ZeroUint()) {
		// refund bond
		txOutItem := &TxOutItem{
			ToAddress:   msg.Destination,
			PoolAddress: poolAddrMgr.GetCurrentPoolAddresses().Current,
			Coins: common.Coins{
				common.NewCoin(common.RuneTicker, nodeAcc.Bond),
			},
		}
		txOut.AddTxOutItem(ctx, keeper, txOutItem)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent("validator_leave",
				sdk.NewAttribute("signer bnb address", msg.Sender.String()),
				sdk.NewAttribute("destination", msg.Destination.String()),
				sdk.NewAttribute("tx", msg.RequestTxHash.String())))
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
