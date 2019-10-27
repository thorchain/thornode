package swapservice

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"gitlab.com/thorchain/bepswap/thornode/common"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/bepswap/thornode/x/swapservice/types"
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
		case MsgSetVersion:
			return handleMsgSetVersion(ctx, keeper, m)
		case MsgBond:
			return handleMsgBond(ctx, keeper, m)
		case MsgNextPoolAddress:
			return handleMsgConfirmNextPoolAddress(ctx, keeper, validatorManager, poolAddressMgr, m)
		case MsgLeave:
			return handleMsgLeave(ctx, keeper, txOutStore, poolAddressMgr, m)
		case MsgAck:
			return handleMsgAck(ctx, keeper, validatorManager, m)
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
		ctx.Logger().Error("message signed by unauthorized account", "asset", msg.Asset)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	ctx.Logger().Info("handle MsgEndPool", "asset", msg.Asset, "requester", msg.Requester, "signer", msg.Signer.String())
	poolStaker, err := keeper.GetPoolStaker(ctx, msg.Asset)
	if nil != err {
		ctx.Logger().Error("fail to get pool staker", err)
		return sdk.ErrInternal(err.Error()).Result()
	}
	// everyone withdraw
	for _, item := range poolStaker.Stakers {
		unstakeMsg := NewMsgSetUnStake(
			item.StakerID,
			sdk.NewUint(10000),
			msg.Asset,
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
		msg.Asset,
		PoolSuspended)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// Handle a message to set pooldata
func handleMsgSetPoolData(ctx sdk.Context, keeper Keeper, msg MsgSetPoolData) sdk.Result {
	if !isSignedByActiveNodeAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "asset", msg.Asset.String())
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	ctx.Logger().Info("handleMsgSetPoolData request", "Asset:"+msg.Asset.String())
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	keeper.SetPoolData(
		ctx,
		msg.Asset,
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
		msg.AssetAmount,
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
		msg.Asset,
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

	ctx.Logger().Info("handleMsgSetStakeData request", "stakerid:"+msg.Asset.String())
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "asset", msg.Asset.String(), "request tx hash", msg.RequestTxHash, "public address", msg.PublicAddress)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	pool := keeper.GetPool(ctx, msg.Asset)
	if pool.Empty() {
		ctx.Logger().Info("pool doesn't exist yet, create a new one", "symbol", msg.Asset.String(), "creator", msg.PublicAddress)
		pool.Asset = msg.Asset
		keeper.SetPool(ctx, pool)
	}
	if err := pool.EnsureValidPoolStatus(msg); nil != err {
		ctx.Logger().Error("check pool status", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	stakeUnits, err := stake(
		ctx,
		keeper,
		msg.Asset,
		msg.RuneAmount,
		msg.AssetAmount,
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
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.RequestTxHash, "source asset", msg.SourceAsset, "target asset", msg.TargetAsset)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}

	gsl := keeper.GetAdminConfigGSL(ctx, EmptyAccAddress)

	amount, err := swap(
		ctx,
		keeper,
		msg.RequestTxHash,
		msg.SourceAsset,
		msg.TargetAsset,
		msg.Amount,
		msg.Requester,
		msg.Destination,
		msg.RequestTxHash,
		msg.TradeTarget,
		gsl,
	) // If so, set the stake data to the value specified in the msg.
	if err != nil {
		ctx.Logger().Error("fail to process swap message", "error", err)

		return sdk.ErrInternal(err.Error()).Result()
	}

	res, err := keeper.cdc.MarshalBinaryLengthPrefixed(struct {
		Asset sdk.Uint `json:"asset"`
	}{
		Asset: amount,
	})
	if nil != err {
		ctx.Logger().Error("fail to encode result to json", "error", err)
		return sdk.ErrInternal("fail to encode result to json").Result()
	}

	toi := &TxOutItem{
		PoolAddress: poolAddrMgr.GetCurrentPoolAddresses().Current,
		ToAddress:   msg.Destination,
	}
	toi.Coins = append(toi.Coins, common.NewCoin(
		msg.TargetAsset,
		amount,
	))
	txOutStore.AddTxOutItem(ctx, keeper, toi, true)
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
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.RequestTxHash, "public address", msg.PublicAddress, "asset", msg.Asset, "withdraw basis points", msg.WithdrawBasisPoints)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := keeper.GetPool(ctx, msg.Asset).EnsureValidPoolStatus(msg); nil != err {
		ctx.Logger().Error("check pool status", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgSetUnstake", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	runeAmt, assetAmount, units, err := unstake(ctx, keeper, msg)
	if nil != err {
		ctx.Logger().Error("fail to UnStake", "error", err)
		return sdk.ErrInternal("fail to process UnStake request").Result()
	}
	res, err := keeper.cdc.MarshalBinaryLengthPrefixed(struct {
		Rune  sdk.Uint `json:"rune"`
		Asset sdk.Uint `json:"asset"`
	}{
		Rune:  runeAmt,
		Asset: assetAmount,
	})
	if nil != err {
		ctx.Logger().Error("fail to marshal result to json", "error", err)
		// if this happen what should we tell the client?
	}

	unstakeEvt := NewEventUnstake(
		runeAmt,
		assetAmount,
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
		msg.Asset,
		unstakeBytes,
		EventSuccess,
	)
	keeper.AddIncompleteEvents(ctx, evt)

	toi := &TxOutItem{
		PoolAddress: poolAddrMgr.currentPoolAddresses.Current,
		ToAddress:   msg.PublicAddress,
	}
	// TODO  this will be a problem , because the RUNE symbol on mainnet and testnet will be different
	toi.Coins = append(toi.Coins, common.NewCoin(
		common.RuneAsset(),
		runeAmt,
	))
	toi.Coins = append(toi.Coins, common.NewCoin(
		msg.Asset,
		assetAmount,
	))
	// for unstake , we should deduct fees
	txOutStore.AddTxOutItem(ctx, keeper, toi, true)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      res,
		Codespace: DefaultCodespace,
	}
}

func refundTx(ctx sdk.Context, tx TxIn, store *TxOutStore, keeper Keeper, poolAddr common.Address, deductFee bool) {
	toi := &TxOutItem{
		ToAddress:   tx.Sender,
		PoolAddress: poolAddr,
		Coins:       tx.Coins,
	}

	// If we recognize one of the coins, and therefore able to refund
	// withholding fees, refund all coins.
	for _, coin := range tx.Coins {
		pool := keeper.GetPool(ctx, coin.Asset)
		if common.IsRuneAsset(coin.Asset) || !pool.BalanceRune.IsZero() {
			store.AddTxOutItem(ctx, keeper, toi, deductFee)
			return
		}
	}

	// Since we have assets, we don't have a pool for, we don't know how to
	// refund and withhold for fees. Instead, we'll create a pool with the
	// amount of assets, and associate them with no stakers (meaning up for
	// grabs). This could be like an airdrop scenario, for example.
	// Don't assume this is the first time we've seen this coin (ie second
	// airdrop).
	for _, coin := range tx.Coins {
		pool := keeper.GetPool(ctx, coin.Asset)
		pool.BalanceAsset = pool.BalanceAsset.Add(coin.Amount)
		pool.Asset = coin.Asset
		if pool.BalanceRune.IsZero() {
			pool.Status = PoolBootstrap
		}
		keeper.SetPool(ctx, pool)
	}
}

// handleMsgConfirmNextPoolAddress , this is the method to handle MsgNextPoolAddress
// MsgNextPoolAddress is a way to prove that the operator has access to the address, and can sign transaction with the given address on chain
func handleMsgConfirmNextPoolAddress(ctx sdk.Context, keeper Keeper, validatorManager *ValidatorManager, poolAddrManager *PoolAddressManager, msg MsgNextPoolAddress) sdk.Result {
	ctx.Logger().Info("receive request to set next pool address", "pool address", msg.NextPoolAddr.String())
	if validatorManager.Meta.Nominated.IsEmpty() {
		return sdk.ErrUnknownRequest("no nominated node yet").Result()
	}
	currentPoolAddr := poolAddrManager.GetCurrentPoolAddresses()
	if !currentPoolAddr.Current.Equals(msg.Sender) {
		return sdk.ErrUnknownRequest("next pool should be send with current pool address").Result()
	}
	nominated, err := keeper.GetNodeAccount(ctx, validatorManager.Meta.Nominated.NodeAddress)
	if err != nil {
		return sdk.ErrInternal(fmt.Sprintf("fail to get nominated node,err:%s", err.Error())).Result()
	}
	nominated.Accounts.SignerBNBAddress = msg.NextPoolAddr
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeNextPoolAddress, sdk.NewAttribute("next pool address", msg.NextPoolAddr.String())))

	keeper.SetNodeAccount(ctx, nominated)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgAck
func handleMsgAck(ctx sdk.Context, keeper Keeper, validatorManager *ValidatorManager, msg MsgAck) sdk.Result {
	ctx.Logger().Info("receive ack to next pool address", "sender address", msg.Sender.String())
	if validatorManager.Meta.Nominated.IsEmpty() {
		return sdk.ErrUnknownRequest("no nominated node yet").Result()
	}

	nominated, err := keeper.GetNodeAccount(ctx, validatorManager.Meta.Nominated.NodeAddress)
	if err != nil {
		return sdk.ErrInternal(fmt.Sprintf("fail to get nominated node,err:%s", err.Error())).Result()
	}
	if !nominated.Accounts.SignerBNBAddress.Equals(msg.Sender) {
		return sdk.ErrUnknownRequest("nominated node has different signer address").Result()
	}
	nominated.SignerActive = true
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeSignerAct, sdk.NewAttribute("next pool address", msg.Sender.String())))
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

// handleMsgSetTxIn gets a tx hash, gets the tx/memo, and triggers
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
			currentPoolAddress := poolAddressMgr.GetCurrentPoolAddresses().Current
			if !currentPoolAddress.Equals(txIn.ObservePoolAddress) {
				ctx.Logger().Error("wrong pool address,refund without deduct fee")
				refundTx(ctx, voter.GetTx(activeNodeAccounts), txOutStore, keeper, txIn.ObservePoolAddress, false)
				continue
			}

			var chain common.Chain
			if len(txIn.Coins) > 0 {
				chain = txIn.Coins[0].Asset.Chain
			}

			m, err := processOneTxIn(ctx, keeper, tx.TxID, txIn, msg.Signer)
			if nil != err || chain.IsEmpty() {
				ctx.Logger().Error("fail to process txIn", "error", err, "txhash", tx.TxID.String())
				refundTx(ctx, voter.GetTx(activeNodeAccounts), txOutStore, keeper, currentPoolAddress, true)
				ee := NewEmptyRefundEvent()
				buf, err := json.Marshal(ee)
				if nil != err {
					return sdk.ErrInternal("fail to marshal EmptyRefund event to json").Result()
				}
				event := NewEvent(ee.Type(), tx.TxID, common.Asset{}, buf, EventRefund)
				keeper.AddIncompleteEvents(ctx, event)
				continue
			}

			// ignoring the error
			_ = keeper.AddToTxInIndex(ctx, uint64(ctx.BlockHeight()), tx.TxID)
			if err := keeper.SetLastChainHeight(ctx, chain, txIn.BlockHeight); nil != err {
				return sdk.ErrInternal("fail to save last height to data store err:" + err.Error()).Result()
			}

			result := handler(ctx, m)
			if !result.IsOK() {
				refundTx(ctx, voter.GetTx(activeNodeAccounts), txOutStore, keeper, currentPoolAddress, true)
			}
		}
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func processOneTxIn(ctx sdk.Context, keeper Keeper, txID common.TxID, tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
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
	case BondMemo:
		newMsg, err = getMsgBondFromMemo(m, txID, tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgBond from memo")
		}
	case NextPoolMemo:
		newMsg = NewMsgNextPoolAddress(txID, m.NextPoolAddr, tx.Sender, signer)
	case AckMemo:
		newMsg = types.NewMsgAck(txID, tx.Sender, signer)
	case LeaveMemo:
		newMsg = NewMsgLeave(txID, tx.Sender, signer)
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
		if !common.IsBNBAsset(coin.Asset) {
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
	if memo.Asset.Equals(coin.Asset) {
		return nil, errors.Errorf("swap from %s to %s is noop, refund", memo.Asset.String(), coin.Asset.String())
	}
	// Looks like at the moment we can only process ont ty
	return NewMsgSwap(txID, coin.Asset, memo.GetAsset(), coin.Amount, tx.Sender, memo.Destination, memo.SlipLimit, signer), nil
}

func getMsgUnstakeFromMemo(memo WithdrawMemo, txID common.TxID, tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	withdrawAmount := sdk.NewUint(MaxWithdrawBasisPoints)
	if len(memo.GetAmount()) > 0 {
		withdrawAmount = sdk.NewUintFromString(memo.GetAmount())
	}
	return NewMsgSetUnStake(tx.Sender, withdrawAmount, memo.GetAsset(), txID, signer), nil
}

func getMsgStakeFromMemo(ctx sdk.Context, memo StakeMemo, txID common.TxID, tx *TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	if len(tx.Coins) > 2 {
		return nil, errors.New("not expecting more than two coins in a stake")
	}
	runeAmount := sdk.ZeroUint()
	assetAmount := sdk.ZeroUint()
	asset := memo.GetAsset()
	if asset.IsEmpty() {
		return nil, errors.New("Unable to determine the intended pool for this stake")
	}
	if common.IsRuneAsset(asset) {
		return nil, errors.New("invalid pool asset")
	}
	for _, coin := range tx.Coins {
		ctx.Logger().Info("coin", "asset", coin.Asset.String(), "amount", coin.Amount.String())
		if common.IsRuneAsset(coin.Asset) {
			runeAmount = coin.Amount
		}
		if asset.Equals(coin.Asset) {
			assetAmount = coin.Amount
		}
	}

	if runeAmount.IsZero() && assetAmount.IsZero() {
		return nil, errors.New("did not find any valid coins for stake")
	}

	// when we receive two coins, but we didn't find the coin specify by asset, then user might send in the wrong coin
	if assetAmount.IsZero() && len(tx.Coins) == 2 {
		return nil, errors.Errorf("did not find %s ", asset)
	}

	return NewMsgSetStakeData(
		asset,
		runeAmount,
		assetAmount,
		tx.Sender,
		txID,
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

func getMsgAddFromMemo(memo AddMemo, txID common.TxID, tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	runeAmount := sdk.ZeroUint()
	assetAmount := sdk.ZeroUint()
	for _, coin := range tx.Coins {
		if common.IsRuneAsset(coin.Asset) {
			runeAmount = coin.Amount
		} else if memo.GetAsset().Equals(coin.Asset) {
			assetAmount = coin.Amount
		}
	}
	return NewMsgAdd(
		memo.GetAsset(),
		runeAmount,
		assetAmount,
		txID,
		signer,
	), nil
}

func getMsgOutboundFromMemo(memo OutboundMemo, txID common.TxID, sender common.Address, signer sdk.AccAddress) (sdk.Msg, error) {
	blockHeight := memo.GetBlockHeight()
	return NewMsgOutboundTx(
		txID,
		blockHeight,
		sender,
		signer,
	), nil
}
func getMsgBondFromMemo(memo BondMemo, txID common.TxID, tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	runeAmount := sdk.ZeroUint()
	for _, coin := range tx.Coins {
		if common.IsRuneAsset(coin.Asset) {
			runeAmount = coin.Amount
		}
	}
	if runeAmount.IsZero() {
		return nil, errors.New("RUNE amount is 0")
	}
	return NewMsgBond(memo.GetNodeAddress(), runeAmount, txID, tx.Sender, signer), nil
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

	pool := keeper.GetPool(ctx, msg.Asset)
	if pool.Asset.IsEmpty() {
		return sdk.ErrUnknownRequest(fmt.Sprintf("pool %s not exist", msg.Asset.String())).Result()
	}
	if msg.AssetAmount.GT(sdk.ZeroUint()) {
		pool.BalanceAsset = pool.BalanceAsset.Add(msg.AssetAmount)
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

// handleMsgSetVersion Update the node account registered version
func handleMsgSetVersion(ctx sdk.Context, keeper Keeper, msg MsgSetVersion) sdk.Result {
	ctx.Logger().Info("receive MsgSetVersion", "trust account info", msg.Version, msg.Signer.String())
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
		ctx.Logger().Error("MsgSetVersion is invalid", "error", err)
		return sdk.ErrUnknownRequest("MsgSetVersion is invalid").Result()
	}

	if int(nodeAccount.Version.Float64()) < msg.Version {
		nodeAccount.Version = common.Amount(msg.Version)
	}

	keeper.SetNodeAccount(ctx, nodeAccount)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent("set_version",
			sdk.NewAttribute("bep_address", msg.Signer.String()),
			sdk.NewAttribute("version", fmt.Sprintf("%d", msg.Version))))
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

// handleMsgBond
func handleMsgBond(ctx sdk.Context, keeper Keeper, msg MsgBond) sdk.Result {
	ctx.Logger().Info("receive MsgBond", "node address", msg.NodeAddress, "txhash", msg.RequestTxHash, "bond", msg.Bond.String())
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "signer", msg.GetSigners())
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgBond", "error", err)
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
	trustAccount := NewTrustAccount(common.NoAddress, sdk.AccAddress{}, "")
	// white list the given bep address
	nodeAccount = NewNodeAccount(msg.NodeAddress, NodeWhiteListed, trustAccount, msg.Bond, msg.BondAddress)
	keeper.SetNodeAccount(ctx, nodeAccount)
	ctx.EventManager().EmitEvent(sdk.NewEvent("new_node", sdk.NewAttribute("address", msg.NodeAddress.String())))
	coinsToMint := keeper.GetAdminConfigWhiteListGasAsset(ctx, sdk.AccAddress{})
	// mint some gas asset
	err = keeper.supplyKeeper.MintCoins(ctx, ModuleName, coinsToMint)
	if nil != err {
		ctx.Logger().Error("fail to mint gas assets", "err", err)
	}
	if err := keeper.supplyKeeper.SendCoinsFromModuleToAccount(ctx, ModuleName, msg.NodeAddress, coinsToMint); nil != err {
		ctx.Logger().Error("fail to send newly minted gas asset to node address")
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func handleMsgLeave(ctx sdk.Context, keeper Keeper, txOut *TxOutStore, poolAddrMgr *PoolAddressManager, msg MsgLeave) sdk.Result {
	ctx.Logger().Info("receive MsgLeave", "sender", msg.Sender.String(), "request tx hash", msg.RequestTxHash)
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "signer", msg.GetSigners())
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgLeave", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	nodeAcc, err := keeper.GetNodeAccountByBondAddress(ctx, msg.Sender)
	if nil != err {
		ctx.Logger().Error("fail to get node account", "error", err)
		return sdk.ErrInternal("fail to get node account by bond bnb address").Result()
	}
	if nodeAcc.IsEmpty() {
		return sdk.ErrUnknownRequest("node account doesn't exist").Result()
	}
	if nodeAcc.Status == NodeActive {
		return sdk.ErrUnknownRequest("active node can't leave").Result()
	}
	curPoolAddr := poolAddrMgr.GetCurrentPoolAddresses()
	signerBNBAddress := nodeAcc.Accounts.SignerBNBAddress
	if curPoolAddr.Current.Equals(signerBNBAddress) || curPoolAddr.Previous.Equals(signerBNBAddress) || curPoolAddr.Next.Equals(signerBNBAddress) {
		return sdk.ErrUnknownRequest("address still in use , cannot leave now").Result()
	}
	if nodeAcc.Bond.GT(sdk.ZeroUint()) {
		// refund bond
		txOutItem := &TxOutItem{
			ToAddress:   nodeAcc.BondAddress,
			PoolAddress: poolAddrMgr.GetCurrentPoolAddresses().Current,
			Coins: common.Coins{
				common.NewCoin(common.RuneAsset(), nodeAcc.Bond),
			},
		}
		txOut.AddTxOutItem(ctx, keeper, txOutItem, true)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent("validator_leave",
				sdk.NewAttribute("signer bnb address", msg.Sender.String()),
				sdk.NewAttribute("destination", nodeAcc.BondAddress.String()),
				sdk.NewAttribute("tx", msg.RequestTxHash.String())))
	}
	// disable the node account
	nodeAcc.Bond = sdk.ZeroUint()
	nodeAcc.UpdateStatus(NodeDisabled)
	keeper.SetNodeAccount(ctx, nodeAcc)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
