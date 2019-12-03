package thorchain

import (
	"encoding/json"
	stdErrors "errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

// EmptyAccAddress empty address
var EmptyAccAddress = sdk.AccAddress{}
var notAuthorized = fmt.Errorf("Not Authorized")
var badVersion = fmt.Errorf("Bad version")

// NewHandler returns a handler for "thorchain" type messages.
func NewHandler(keeper Keeper, poolAddressMgr *PoolAddressManager, txOutStore *TxOutStore, validatorManager *ValidatorManager) sdk.Handler {

	// Classic Handler
	classic := NewClassicHandler(keeper, poolAddressMgr, txOutStore, validatorManager)

	// New arch handlers
	poolDataHandler := NewPoolDataHandler(keeper)

	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		version := keeper.GetLowestActiveVersion(ctx)
		switch m := msg.(type) {
		case MsgSetPoolData:
			return poolDataHandler.Run(ctx, m, version)
		default:
			return classic(ctx, msg)
		}
	}
}

// NewClassicHandler returns a handler for "thorchain" type messages.
func NewClassicHandler(keeper Keeper, poolAddressMgr *PoolAddressManager, txOutStore *TxOutStore, validatorManager *ValidatorManager) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch m := msg.(type) {
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
			return handleMsgNoOp(ctx)
		case MsgEndPool:
			return handleOperatorMsgEndPool(ctx, keeper, txOutStore, poolAddressMgr, m)
		case MsgSetTrustAccount:
			return handleMsgSetTrustAccount(ctx, keeper, m)
		case MsgSetVersion:
			return handleMsgSetVersion(ctx, keeper, m)
		case MsgBond:
			return handleMsgBond(ctx, keeper, m)
		case MsgYggdrasil:
			return handleMsgYggdrasil(ctx, keeper, txOutStore, poolAddressMgr, validatorManager, m)
		case MsgNextPoolAddress:
			return handleMsgConfirmNextPoolAddress(ctx, keeper, poolAddressMgr, validatorManager, txOutStore, m)
		case MsgLeave:
			return handleMsgLeave(ctx, keeper, txOutStore, validatorManager, m)
		case MsgAck:
			return handleMsgAck(ctx, keeper, poolAddressMgr, validatorManager, m)
		case MsgReserveContributor:
			return handleMsgReserveContributor(ctx, keeper, m)
		default:
			errMsg := fmt.Sprintf("Unrecognized thorchain Msg type: %v", m)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// handleOperatorMsgEndPool operators decide it is time to end the pool
func handleOperatorMsgEndPool(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore, poolAddrMgr *PoolAddressManager, msg MsgEndPool) sdk.Result {
	if !isSignedByActiveNodeAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "asset", msg.Asset)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	ctx.Logger().Info("handle MsgEndPool", "asset", msg.Asset, "requester", msg.Tx.FromAddress, "signer", msg.Signer.String())
	poolStaker, err := keeper.GetPoolStaker(ctx, msg.Asset)
	if nil != err {
		ctx.Logger().Error("fail to get pool staker", err)
		return sdk.ErrInternal(err.Error()).Result()
	}

	// everyone withdraw
	for _, item := range poolStaker.Stakers {
		unstakeMsg := NewMsgSetUnStake(
			msg.Tx,
			item.RuneAddress,
			sdk.NewUint(10000),
			msg.Asset,
			msg.Signer,
		)

		result := handleMsgSetUnstake(ctx, keeper, txOutStore, poolAddrMgr, unstakeMsg)
		if !result.IsOK() {
			ctx.Logger().Error("fail to unstake", "staker", item.RuneAddress)
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

func processStakeEvent(ctx sdk.Context, keeper Keeper, msg MsgSetStakeData, stakeUnits sdk.Uint, eventStatus EventStatus) error {
	var stakeEvt EventStake
	if eventStatus == EventRefund {
		// do not log event if the stake failed
		return nil
	}

	stakeEvt = NewEventStake(
		msg.Asset,
		stakeUnits,
	)
	stakeBytes, err := json.Marshal(stakeEvt)
	if err != nil {
		ctx.Logger().Error("fail to save event", err)
		return errors.Wrap(err, "fail to marshal stake event to json")
	}

	evt := NewEvent(
		stakeEvt.Type(),
		ctx.BlockHeight(),
		msg.Tx,
		stakeBytes,
		eventStatus,
	)

	if err := keeper.AddIncompleteEvents(ctx, evt); err != nil {
		return err
	}

	if eventStatus != EventRefund {
		// since there is no outbound tx for staking, we'll complete the event now
		tx := common.Tx{ID: common.BlankTxID}
		err := completeEvents(ctx, keeper, msg.Tx.ID, common.Txs{tx})
		if err != nil {
			ctx.Logger().Error("unable to complete events", "error", err)
			return err
		}
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

	ctx.Logger().Info("handleMsgSetStakeData request", "stakerid", msg.Asset.String())
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "asset", msg.Asset.String(), "request tx hash", msg.Tx.ID, "rune address", msg.RuneAddress)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	pool := keeper.GetPool(ctx, msg.Asset)
	if pool.Empty() {
		ctx.Logger().Info("pool doesn't exist yet, create a new one", "symbol", msg.Asset.String(), "creator", msg.RuneAddress)
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
		msg.RuneAddress,
		msg.AssetAddress,
		msg.Tx.ID,
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
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.Tx.ID, "source asset", msg.Tx.Coins[0].Asset, "target asset", msg.TargetAsset)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	gsl := keeper.GetAdminConfigGSL(ctx, EmptyAccAddress)
	chain := msg.TargetAsset.Chain
	currentAddr := poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(chain)
	if nil == currentAddr {
		msg := fmt.Sprintf("don't have pool address for chain : %s", chain)
		ctx.Logger().Error(msg)
		return sdk.ErrInternal(msg).Result()
	}
	amount, err := swap(
		ctx,
		keeper,
		msg.Tx,
		msg.TargetAsset,
		msg.Destination,
		msg.TradeTarget,
		gsl,
	) // If so, set the stake data to the value specified in the msg.
	if err != nil {
		ctx.Logger().Error("fail to process swap message", "error", err)

		return sdk.ErrInternal(err.Error()).Result()
	}

	res, err := keeper.Cdc().MarshalBinaryLengthPrefixed(struct {
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
		PoolAddress: currentAddr.PubKey,
		ToAddress:   msg.Destination,
		Coin:        common.NewCoin(msg.TargetAsset, amount),
	}
	txOutStore.AddTxOutItem(ctx, keeper, toi, false)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      res,
		Codespace: DefaultCodespace,
	}
}

// handleMsgSetUnstake process unstake
func handleMsgSetUnstake(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore, poolAddrMgr *PoolAddressManager, msg MsgSetUnStake) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgSetUnstake from : %s(%s) unstake (%s)", msg, msg.RuneAddress, msg.WithdrawBasisPoints))
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.Tx.ID, "rune address", msg.RuneAddress, "asset", msg.Asset, "withdraw basis points", msg.WithdrawBasisPoints)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}

	bnbPoolAddr := poolAddrMgr.currentPoolAddresses.Current.GetByChain(common.BNBChain)
	if nil == bnbPoolAddr {
		msg := fmt.Sprintf("THORNode don't have pool for chain : %s ", common.BNBChain)
		ctx.Logger().Error(msg)
		return sdk.ErrUnknownRequest(msg).Result()
	}
	currentAddr := poolAddrMgr.currentPoolAddresses.Current.GetByChain(msg.Asset.Chain)
	if nil == currentAddr {
		msg := fmt.Sprintf("THORNode don't have pool for chain : %s ", msg.Asset.Chain)
		ctx.Logger().Error(msg)
		return sdk.ErrUnknownRequest(msg).Result()
	}
	if err := keeper.GetPool(ctx, msg.Asset).EnsureValidPoolStatus(msg); nil != err {
		ctx.Logger().Error("check pool status", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgSetUnstake", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	poolStaker, err := keeper.GetPoolStaker(ctx, msg.Asset)
	if nil != err {
		ctx.Logger().Error("fail to get pool staker", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	stakerUnit := poolStaker.GetStakerUnit(msg.RuneAddress)

	runeAmt, assetAmount, units, err := unstake(ctx, keeper, msg)
	if nil != err {
		ctx.Logger().Error("fail to UnStake", "error", err)
		return sdk.ErrInternal("fail to process UnStake request").Result()
	}
	res, err := keeper.Cdc().MarshalBinaryLengthPrefixed(struct {
		Rune  sdk.Uint `json:"rune"`
		Asset sdk.Uint `json:"asset"`
	}{
		Rune:  runeAmt,
		Asset: assetAmount,
	})
	if nil != err {
		ctx.Logger().Error("fail to marshal result to json", "error", err)
		// if this happen what should THORNode tell the client?
	}

	unstakeEvt := NewEventUnstake(
		msg.Asset,
		units,
		0,             // TODO: make this real data
		sdk.ZeroDec(), // TODO: make this real data
	)
	unstakeBytes, err := json.Marshal(unstakeEvt)
	if err != nil {
		ctx.Logger().Error("fail to save event", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	evt := NewEvent(
		unstakeEvt.Type(),
		ctx.BlockHeight(),
		msg.Tx,
		unstakeBytes,
		EventSuccess,
	)

	if err := keeper.AddIncompleteEvents(ctx, evt); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}

	toi := &TxOutItem{
		Chain:       common.BNBChain,
		InHash:      msg.Tx.ID,
		PoolAddress: bnbPoolAddr.PubKey,
		ToAddress:   stakerUnit.RuneAddress,
		Coin:        common.NewCoin(common.RuneAsset(), runeAmt),
	}
	txOutStore.AddTxOutItem(ctx, keeper, toi, false)

	toi = &TxOutItem{
		Chain:       msg.Asset.Chain,
		InHash:      msg.Tx.ID,
		PoolAddress: currentAddr.PubKey,
		ToAddress:   stakerUnit.AssetAddress,
		Coin:        common.NewCoin(msg.Asset, assetAmount),
	}
	// for unstake , THORNode should deduct fees
	txOutStore.AddTxOutItem(ctx, keeper, toi, false)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      res,
		Codespace: DefaultCodespace,
	}
}

// handleMsgConfirmNextPoolAddress , this is the method to handle MsgNextPoolAddress
// MsgNextPoolAddress is a way to prove that the operator has access to the address, and can sign transaction with the given address on chain
func handleMsgConfirmNextPoolAddress(ctx sdk.Context, keeper Keeper, poolAddrManager *PoolAddressManager, validatorMgr *ValidatorManager, txOut *TxOutStore, msg MsgNextPoolAddress) sdk.Result {
	ctx.Logger().Info("receive request to set next pool pub key", "pool pub key", msg.NextPoolPubKey.String())
	if err := msg.ValidateBasic(); nil != err {
		return err.Result()
	}
	if !poolAddrManager.IsRotateWindowOpen {
		return sdk.ErrUnknownRequest("pool address rotate window not open yet").Result()
	}
	currentPoolAddresses := poolAddrManager.GetCurrentPoolAddresses()
	currentChainPoolAddr := currentPoolAddresses.Next.GetByChain(msg.Chain)
	if nil != currentChainPoolAddr {
		ctx.Logger().Error(fmt.Sprintf("next pool for chain %s had been confirmed already", msg.Chain))
		return sdk.ErrUnknownRequest(fmt.Sprintf("next pool for chain %s had been confirmed already", msg.Chain)).Result()
	}
	currentAddr := currentPoolAddresses.Current.GetByChain(msg.Chain)
	if nil == currentAddr || currentAddr.IsEmpty() {
		msg := fmt.Sprintf("THORNode donnot have pool for chain %s", msg.Chain)
		ctx.Logger().Error(msg)
		return sdk.ErrUnknownRequest(msg).Result()
	}
	addr, err := currentAddr.PubKey.GetAddress(msg.Chain)
	if nil != err {
		ctx.Logger().Error("fail to get address from pub key", "chain", msg.Chain, err)
		return sdk.ErrInternal("fail to get address from pub key").Result()
	}

	// nextpool memo need to be initiated by current pool
	if !addr.Equals(msg.Sender) {
		return sdk.ErrUnknownRequest("next pool should be send with current pool address").Result()
	}
	// thorchain observed the next pool address memo, but it has not been confirmed yet
	pkey, err := common.NewPoolPubKey(msg.Chain, 0, msg.NextPoolPubKey)
	if nil != err {
		ctx.Logger().Error("fail to get pool pubkey", "chain", msg.Chain, err)
		return sdk.ErrInternal("fail to get pool pubkey").Result()
	}

	poolAddrManager.ObservedNextPoolAddrPubKey = poolAddrManager.ObservedNextPoolAddrPubKey.TryAddKey(pkey)

	// if THORNode observed a valid nextpool transaction, that means the nominated validator had join the signing committee to generate a new pub key
	// with TSS, if they don't join , then the key won't be generated
	nominatedAccount := validatorMgr.Meta.Nominated
	if !nominatedAccount.IsEmpty() {
		for _, item := range nominatedAccount {
			item.SignerActive = true
			keeper.SetNodeAccount(ctx, item)
		}
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeNextPoolPubKeyObserved,
			sdk.NewAttribute("next pool pub key", msg.NextPoolPubKey.String()),
			sdk.NewAttribute("chain", msg.Chain.String())))

	txOut.AddTxOutItem(ctx, keeper, &TxOutItem{
		Chain:       common.BNBChain,
		ToAddress:   addr,
		PoolAddress: msg.NextPoolPubKey,
		Coin:        common.NewCoin(common.BNBAsset, sdk.NewUint(1)),
		Memo:        "ack",
	}, true)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgReserveContributor
func handleMsgReserveContributor(ctx sdk.Context, keeper Keeper, msg MsgReserveContributor) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgReserveContributor from : %s reserve %s (%s)", msg, msg.Contributor.Address.String(), msg.Contributor.Amount.String()))
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account")
		return sdk.ErrUnauthorized("Not authorized").Result()
	}

	reses := keeper.GetReservesContributors(ctx)
	reses = reses.Add(msg.Contributor)
	keeper.SetReserveContributors(ctx, reses)

	vault := keeper.GetVaultData(ctx)
	vault.TotalReserve = vault.TotalReserve.Add(msg.Contributor.Amount)
	keeper.SetVaultData(ctx, vault)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgAck
func handleMsgAck(ctx sdk.Context, keeper Keeper, poolAddrMgr *PoolAddressManager, validatorMgr *ValidatorManager, msg MsgAck) sdk.Result {
	ctx.Logger().Info("receive ack to next pool pub key", "sender address", msg.Sender.String())
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid ack msg", "err", err)
		return err.Result()
	}

	if !poolAddrMgr.IsRotateWindowOpen {
		return sdk.ErrUnknownRequest("pool rotation window not open").Result()
	}

	if poolAddrMgr.ObservedNextPoolAddrPubKey.IsEmpty() {
		return sdk.ErrUnknownRequest("didn't observe next pool address pub key").Result()
	}
	chainPubKey := poolAddrMgr.ObservedNextPoolAddrPubKey.GetByChain(msg.Chain)
	if nil == chainPubKey {
		msg := fmt.Sprintf("THORNode donnot have pool for chain %s", msg.Chain)
		ctx.Logger().Error(msg)
		return sdk.ErrUnknownRequest(msg).Result()
	}
	addr, err := chainPubKey.PubKey.GetAddress(msg.Chain)
	if nil != err {
		ctx.Logger().Error("fail to get address from pub key", "chain", msg.Chain, err)
		return sdk.ErrInternal("fail to get address from pub key").Result()
	}
	if !addr.Equals(msg.Sender) {
		ctx.Logger().Error("observed next pool address and ack address is different", "chain", msg.Chain)
		return sdk.ErrUnknownRequest("observed next pool address and ack address is different").Result()
	}

	poolAddrMgr.currentPoolAddresses.Next = poolAddrMgr.currentPoolAddresses.Next.TryAddKey(chainPubKey)
	poolAddrMgr.ObservedNextPoolAddrPubKey = poolAddrMgr.ObservedNextPoolAddrPubKey.TryRemoveKey(chainPubKey)

	nominatedNode := validatorMgr.Meta.Nominated
	queuedNode := validatorMgr.Meta.Queued
	for _, item := range nominatedNode {
		item.TryAddSignerPubKey(chainPubKey.PubKey)
		keeper.SetNodeAccount(ctx, item)
	}
	activeNodes, err := keeper.ListActiveNodeAccounts(ctx)
	if nil != err {
		ctx.Logger().Error("fail to get all active node accounts", "error", err)
		return sdk.ErrInternal("fail to get all active node accounts").Result()
	}

	for _, item := range activeNodes {
		if queuedNode.Contains(item) {
			// queued node doesn't join the signing committee
			continue
		}
		item.TryAddSignerPubKey(chainPubKey.PubKey)
		keeper.SetNodeAccount(ctx, item)
	}

	AddGasFees(ctx, keeper, msg.Tx.Gas)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeNexePoolPubKeyConfirmed,
			sdk.NewAttribute("pubkey", poolAddrMgr.currentPoolAddresses.Next.String()),
			sdk.NewAttribute("address", msg.Sender.String()),
			sdk.NewAttribute("chain", msg.Chain.String())))
	// THORNode have a pool address confirmed by a chain
	keeper.SetPoolAddresses(ctx, poolAddrMgr.currentPoolAddresses)

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
		na, err := keeper.GetNodeAccount(ctx, msg.Signer)
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

		if preConsensus == false && postConsensus == true && voter.Height == 0 {
			voter.Height = ctx.BlockHeight()
			keeper.SetTxInVoter(ctx, voter)
			txIn := voter.GetTx(activeNodeAccounts)
			var chain common.Chain
			if len(txIn.Coins) > 0 {
				chain = txIn.Coins[0].Asset.Chain
			}

			currentPoolAddress := poolAddressMgr.GetCurrentPoolAddresses().Current.GetByChain(chain)
			yggExists := keeper.YggdrasilExists(ctx, txIn.ObservePoolAddress)
			if !currentPoolAddress.PubKey.Equals(txIn.ObservePoolAddress) && !yggExists {
				ctx.Logger().Error("wrong pool address,refund without deduct fee", "pubkey", currentPoolAddress.PubKey.String(), "observe pool addr", txIn.ObservePoolAddress)
				refundTx(ctx, voter.TxID, txIn, txOutStore, keeper, txIn.ObservePoolAddress, chain, false)
				continue
			}

			m, err := processOneTxIn(ctx, keeper, tx.TxID, txIn, msg.Signer)
			if nil != err || chain.IsEmpty() {
				ctx.Logger().Error("fail to process txIn", "error", err, "txhash", tx.TxID.String())
				// Detect if the txIn is to the thorchain network or from the
				// thorchain network
				addr, err := txIn.ObservePoolAddress.GetAddress(chain)
				if err != nil {
					ctx.Logger().Error("fail to get address", "error", err, "txhash", tx.TxID.String())
					continue
				}

				if addr.Equals(txIn.Sender) {

					if keeper.YggdrasilExists(ctx, txIn.ObservePoolAddress) {
						ygg, err := keeper.GetYggdrasil(ctx, txIn.ObservePoolAddress)
						if nil != err {
							ctx.Logger().Error("fail to get yggdrasil", err)
							return sdk.ErrInternal("fail to get yggdrasil").Result()
						}
						var expectedCoins common.Coins
						memo, _ := ParseMemo(txIn.Memo)
						switch memo.GetType() {
						case txYggdrasilReturn:
							expectedCoins = ygg.Coins
						case txOutbound:
							txID := memo.GetTxID()
							inVoter := keeper.GetTxInVoter(ctx, txID)
							origTx := inVoter.GetTx(activeNodeAccounts)
							expectedCoins = origTx.Coins
						}

						na, err := keeper.GetNodeAccountByPubKey(ctx, ygg.PubKey)
						if err != nil {
							ctx.Logger().Error("fail to get node account", "error", err, "txhash", tx.TxID.String())
							return sdk.ErrInternal("fail to get node account").Result()
						}

						// Slash the node account, since THORNode are unable to
						// process the tx (ie unscheduled tx)
						var minusCoins common.Coins       // track funds to subtract from ygg pool
						minusRune := sdk.ZeroUint()       // track amt of rune to slash from bond
						for _, coin := range txIn.Coins { // assumes coins are asset uniq
							expectedCoin := expectedCoins.GetCoin(coin.Asset)
							if expectedCoin.Amount.LT(coin.Amount) {
								// take an additional 25% to ensure a penalty
								// is made by multiplying by 5, then divide by
								// 4
								diff := common.SafeSub(coin.Amount, expectedCoin.Amount).MulUint64(5).QuoUint64(4)
								if coin.Asset.IsRune() {
									minusRune = common.SafeSub(coin.Amount, diff)
									minusCoins = append(minusCoins, common.NewCoin(coin.Asset, diff))
								} else {
									pool := keeper.GetPool(ctx, coin.Asset)
									if !pool.Empty() {
										minusRune = pool.AssetValueInRune(diff)
										minusCoins = append(minusCoins, common.NewCoin(coin.Asset, diff))
										// Update pool balances
										pool.BalanceRune = pool.BalanceRune.Add(minusRune)
										pool.BalanceAsset = common.SafeSub(pool.BalanceAsset, diff)
										keeper.SetPool(ctx, pool)
									}
								}
							}
						}
						na.SubBond(minusRune)
						keeper.SetNodeAccount(ctx, na)
						ygg.SubFunds(minusCoins)
						if err := keeper.SetYggdrasil(ctx, ygg); nil != err {
							ctx.Logger().Error("fail to save yggdrasil", err)
							return sdk.ErrInternal("fail to save yggdrasil").Result()
						}
					}
				} else {
					// To thorchain network
					refundTx(ctx, voter.TxID, txIn, txOutStore, keeper, currentPoolAddress.PubKey, currentPoolAddress.Chain, true)
					ee := NewEmptyRefundEvent()
					buf, err := json.Marshal(ee)
					if nil != err {
						return sdk.ErrInternal("fail to marshal EmptyRefund event to json").Result()
					}
					event := NewEvent(
						ee.Type(),
						ctx.BlockHeight(),
						txIn.GetCommonTx(tx.TxID),
						buf,
						EventRefund,
					)

					if err := keeper.AddIncompleteEvents(ctx, event); err != nil {
						return sdk.ErrInternal(err.Error()).Result()
					}
				}
				continue
			}

			// ignoring the error
			_ = keeper.AddToTxInIndex(ctx, uint64(ctx.BlockHeight()), tx.TxID)
			if err := keeper.SetLastChainHeight(ctx, chain, txIn.BlockHeight); nil != err {
				return sdk.ErrInternal("fail to save last height to data store err:" + err.Error()).Result()
			}

			// add this chain to our list of supported chains
			chains, err := keeper.GetChains(ctx)
			if err != nil {
				return sdk.ErrInternal("fail to get chains:" + err.Error()).Result()
			}
			chains = append(chains, chain)
			keeper.SetChains(ctx, chains)

			// add addresses to observing addresses. This is used to detect
			// active/inactive observing node accounts
			keeper.AddObservingAddresses(ctx, txIn.Signers)

			result := handler(ctx, m)
			if !result.IsOK() {
				refundTx(ctx, voter.TxID, txIn, txOutStore, keeper, currentPoolAddress.PubKey, currentPoolAddress.Chain, true)
			}
		}
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func processOneTxIn(ctx sdk.Context, keeper Keeper, txID common.TxID, tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	if len(tx.Coins) == 0 {
		return nil, fmt.Errorf("no coin found")
	}
	memo, err := ParseMemo(tx.Memo)
	if err != nil {
		return nil, errors.Wrap(err, "fail to parse memo")
	}
	// THORNode should not have one tx across chain, if it is cross chain it should be separate tx
	chain := tx.Coins[0].Asset.Chain
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
		tx := tx.GetCommonTx(txID)
		newMsg, err = getMsgOutboundFromMemo(m, tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgOutbound from memo")
		}
	case BondMemo:
		newMsg, err = getMsgBondFromMemo(m, txID, tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgBond from memo")
		}
	case NextPoolMemo:
		txIn := tx.GetCommonTx(txID)
		newMsg = NewMsgNextPoolAddress(txIn, m.NextPoolAddr, tx.Sender, chain, signer)
	case AckMemo:
		txIn := tx.GetCommonTx(txID)
		newMsg = types.NewMsgAck(txIn, tx.Sender, chain, signer)
	case LeaveMemo:
		tx := tx.GetCommonTx(txID)
		newMsg = NewMsgLeave(tx, signer)
	case YggdrasilFundMemo:
		pk, err := keeper.FindPubKeyOfAddress(ctx, tx.To, tx.Coins[0].Asset.Chain)
		if err != nil {
			return nil, errors.Wrap(err, "fail to find Yggdrasil pubkey")
		}
		newMsg = NewMsgYggdrasil(pk, true, tx.Coins, txID, signer)
	case YggdrasilReturnMemo:
		pk, err := keeper.FindPubKeyOfAddress(ctx, tx.Sender, tx.Coins[0].Asset.Chain)
		if err != nil {
			return nil, errors.Wrap(err, "fail to find Yggdrasil pubkey")
		}
		newMsg = NewMsgYggdrasil(pk, false, tx.Coins, txID, signer)
	case ReserveMemo:
		res := NewReserveContributor(tx.Sender, tx.Coins[0].Amount)
		newMsg = NewMsgReserveContributor(res, signer)
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
		if !coin.Asset.IsBNB() {
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

	// Looks like at the moment THORNode can only process ont ty
	return NewMsgSwap(tx.GetCommonTx(txID), memo.GetAsset(), memo.Destination, memo.SlipLimit, signer), nil
}

func getMsgUnstakeFromMemo(memo WithdrawMemo, txID common.TxID, txIn TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	withdrawAmount := sdk.NewUint(MaxWithdrawBasisPoints)
	if len(memo.GetAmount()) > 0 {
		withdrawAmount = sdk.NewUintFromString(memo.GetAmount())
	}
	tx := txIn.GetCommonTx(txID)
	return NewMsgSetUnStake(tx, txIn.Sender, withdrawAmount, memo.GetAsset(), signer), nil
}

func getMsgStakeFromMemo(ctx sdk.Context, memo StakeMemo, txID common.TxID, txIn *TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	if len(txIn.Coins) > 2 {
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
	for _, coin := range txIn.Coins {
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
	if assetAmount.IsZero() && len(txIn.Coins) == 2 {
		return nil, errors.Errorf("did not find %s ", asset)
	}

	runeAddr := txIn.Sender
	assetAddr := memo.GetDestination()
	if !runeAddr.IsChain(common.BNBChain) {
		runeAddr = memo.GetDestination()
		assetAddr = txIn.Sender
	} else {
		// if it is on BNB chain , while the asset addr is empty, then the asset addr is runeAddr
		if assetAddr.IsEmpty() {
			assetAddr = runeAddr
		}
	}

	tx := txIn.GetCommonTx(txID)

	return NewMsgSetStakeData(
		tx,
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

func getMsgAddFromMemo(memo AddMemo, txID common.TxID, txIn TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	runeAmount := sdk.ZeroUint()
	assetAmount := sdk.ZeroUint()
	for _, coin := range txIn.Coins {
		if coin.Asset.IsRune() {
			runeAmount = coin.Amount
		} else if memo.GetAsset().Equals(coin.Asset) {
			assetAmount = coin.Amount
		}
	}
	tx := txIn.GetCommonTx(txID)
	return NewMsgAdd(
		tx,
		memo.GetAsset(),
		runeAmount,
		assetAmount,
		signer,
	), nil
}

func getMsgOutboundFromMemo(memo OutboundMemo, tx common.Tx, signer sdk.AccAddress) (sdk.Msg, error) {
	return NewMsgOutboundTx(
		tx,
		memo.GetTxID(),
		signer,
	), nil
}
func getMsgBondFromMemo(memo BondMemo, txID common.TxID, tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	runeAmount := sdk.ZeroUint()
	for _, coin := range tx.Coins {
		if coin.Asset.IsRune() {
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
	ctx.Logger().Info(fmt.Sprintf("receive MsgAdd %s", msg.Tx.ID))
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

	// emit event
	addEvt := NewEventAdd(
		pool.Asset,
	)
	stakeBytes, err := json.Marshal(addEvt)
	if err != nil {
		ctx.Logger().Error("fail to marshal add event", err)
		err = errors.Wrap(err, "fail to marshal add event to json")
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	evt := NewEvent(
		addEvt.Type(),
		ctx.BlockHeight(),
		msg.Tx,
		stakeBytes,
		EventSuccess,
	)
	keeper.SetCompletedEvent(ctx, evt)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgNoOp doesn't do anything, its a no op
func handleMsgNoOp(ctx sdk.Context) sdk.Result {
	ctx.Logger().Info("receive no op msg")
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgOutboundTx processes outbound tx from our pool
func handleMsgOutboundTx(ctx sdk.Context, keeper Keeper, poolAddressMgr *PoolAddressManager, msg MsgOutboundTx) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgOutboundTx %s", msg.Tx.ID))
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "signer", msg.GetSigners())
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgOutboundTx", "error", err)
		return err.Result()
	}
	currentChainPoolAddr := poolAddressMgr.GetCurrentPoolAddresses().Current.GetByChain(msg.Tx.Chain)
	if nil == currentChainPoolAddr {
		msg := fmt.Sprintf("THORNode don't have pool for chain %s", msg.Tx.Chain)
		ctx.Logger().Error(msg)
		return sdk.ErrUnknownRequest(msg).Result()
	}

	currentPoolAddr, err := currentChainPoolAddr.GetAddress()
	if nil != err {
		ctx.Logger().Error("fail to get current pool address", "error", err)
		return sdk.ErrUnknownRequest("fail to get current pool address").Result()
	}
	previousChainPoolAddr := poolAddressMgr.GetCurrentPoolAddresses().Previous.GetByChain(msg.Tx.Chain)
	previousPoolAddr := common.NoAddress
	if nil != previousChainPoolAddr {
		previousPoolAddr, err = previousChainPoolAddr.GetAddress()
		if nil != err {
			ctx.Logger().Error("fail to get previous pool address", "error", err)
			return sdk.ErrUnknownRequest("fail to get previous pool address").Result()
		}
	}

	if !currentPoolAddr.Equals(msg.Tx.FromAddress) && !previousPoolAddr.Equals(msg.Tx.FromAddress) {
		ctx.Logger().Error("message sent by unauthorized account", "sender", msg.Tx.FromAddress.String(), "current pool addr", currentPoolAddr.String())
		return sdk.ErrUnauthorized("Not authorized").Result()
	}

	voter := keeper.GetTxInVoter(ctx, msg.InTxID)
	voter.AddOutTx(msg.Tx)
	keeper.SetTxInVoter(ctx, voter)

	// complete events
	if voter.IsDone() {
		err := completeEvents(ctx, keeper, msg.InTxID, voter.OutTxs)
		if err != nil {
			ctx.Logger().Error("unable to complete events", "error", err)
			return sdk.ErrInternal(err.Error()).Result()
		}
	}

	// Apply Gas fees
	activeNodeAccounts, err := keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		ctx.Logger().Error("unable to get active node accounts", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	inTx := voter.GetTx(activeNodeAccounts)
	tx := inTx.GetCommonTx(msg.InTxID)
	tx.Gas = msg.Tx.Gas // get gas from outbound tx, and replace the inbound gas for applying gas
	AddGasFees(ctx, keeper, tx.Gas)

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
				msg.Tx.Coins.Contains(tx.Coin) {
				txOut.TxArray[i].OutHash = msg.Tx.ID
			}
		}
		keeper.SetTxOut(ctx, txOut)
	}
	keeper.SetLastSignedHeight(ctx, sdk.NewUint(uint64(voter.Height)))

	// If THORNode are sending from a yggdrasil pool, decrement coins on record
	pk, err := keeper.FindPubKeyOfAddress(ctx, msg.Tx.FromAddress, msg.Tx.Chain)
	if err != nil {
		ctx.Logger().Error("unable to find Yggdrasil pubkey", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	if !pk.IsEmpty() {
		ygg, err := keeper.GetYggdrasil(ctx, pk)
		if nil != err {
			ctx.Logger().Error("fail to get yggdrasil", err)
			return sdk.ErrInternal("fail to get yggdrasil").Result()
		}
		ygg.SubFunds(msg.Tx.Coins)
		if err := keeper.SetYggdrasil(ctx, ygg); nil != err {
			ctx.Logger().Error("fail to save yggdrasil", err)
			return sdk.ErrInternal("fail to save yggdrasil").Result()
		}
	}

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

	prevVal, err := keeper.GetAdminConfigValue(ctx, msg.AdminConfig.Key, nil)
	if err != nil {
		ctx.Logger().Error("unable to get admin config", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	keeper.SetAdminConfig(ctx, msg.AdminConfig)

	newVal, err := keeper.GetAdminConfigValue(ctx, msg.AdminConfig.Key, nil)
	if err != nil {
		ctx.Logger().Error("unable to get admin config", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	if newVal != "" && prevVal != newVal {
		adminEvt := NewEventAdminConfig(
			msg.AdminConfig.Key.String(),
			msg.AdminConfig.Value,
		)
		stakeBytes, err := json.Marshal(adminEvt)
		if err != nil {
			ctx.Logger().Error("fail to unmarshal admin config event", err)
			err = errors.Wrap(err, "fail to marshal admin config event to json")
			return sdk.ErrUnknownRequest(err.Error()).Result()
		}

		evt := NewEvent(
			adminEvt.Type(),
			ctx.BlockHeight(),
			msg.Tx,
			stakeBytes,
			EventSuccess,
		)
		keeper.SetCompletedEvent(ctx, evt)
	}

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

	if nodeAccount.Version.LT(msg.Version) {
		nodeAccount.Version = msg.Version
	}

	keeper.SetNodeAccount(ctx, nodeAccount)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent("set_version",
			sdk.NewAttribute("bep_address", msg.Signer.String()),
			sdk.NewAttribute("version", msg.Version.String())))
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgSetTrustAccount Update node account
func handleMsgSetTrustAccount(ctx sdk.Context, keeper Keeper, msg MsgSetTrustAccount) sdk.Result {
	ctx.Logger().Info("receive MsgSetTrustAccount", "validator consensus pub key", msg.ValidatorConsPubKey, "pubkey", msg.NodePubKeys.String())
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
	if err := keeper.EnsureTrustAccountUnique(ctx, msg.ValidatorConsPubKey, msg.NodePubKeys); nil != err {
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	// Here make sure THORNode don't change the node account's bond

	nodeAccount.UpdateStatus(NodeStandby, ctx.BlockHeight())
	keeper.SetNodeAccount(ctx, nodeAccount)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent("set_trust_account",
			sdk.NewAttribute("node_address", msg.Signer.String()),
			sdk.NewAttribute("node_secp256k1_pubkey", msg.NodePubKeys.Secp256k1.String()),
			sdk.NewAttribute("node_ed25519_pubkey", msg.NodePubKeys.Ed25519.String()),
			sdk.NewAttribute("validator_consensus_pub_key", msg.ValidatorConsPubKey)))
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
	// THORNode will not have pub keys at the moment, so have to leave it empty
	emptyPubKeys := common.PubKeys{
		Secp256k1: common.EmptyPubKey,
		Ed25519:   common.EmptyPubKey,
	}
	// white list the given bep address
	nodeAccount = NewNodeAccount(msg.NodeAddress, NodeWhiteListed, emptyPubKeys, "", msg.Bond, msg.BondAddress, ctx.BlockHeight())
	keeper.SetNodeAccount(ctx, nodeAccount)
	ctx.EventManager().EmitEvent(sdk.NewEvent("new_node", sdk.NewAttribute("address", msg.NodeAddress.String())))
	coinsToMint := keeper.GetAdminConfigWhiteListGasAsset(ctx, sdk.AccAddress{})
	// mint some gas asset
	err = keeper.Supply().MintCoins(ctx, ModuleName, coinsToMint)
	if nil != err {
		ctx.Logger().Error("fail to mint gas assets", "err", err)
	}
	if err := keeper.Supply().SendCoinsFromModuleToAccount(ctx, ModuleName, msg.NodeAddress, coinsToMint); nil != err {
		ctx.Logger().Error("fail to send newly minted gas asset to node address")
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgYggdrasil
func handleMsgYggdrasil(ctx sdk.Context, keeper Keeper, txOut *TxOutStore, poolAddrMgr *PoolAddressManager, validatorMgr *ValidatorManager, msg MsgYggdrasil) sdk.Result {
	ctx.Logger().Info("receive MsgYggdrasil", "pubkey", msg.PubKey.String(), "add_funds", msg.AddFunds, "coins", msg.Coins)

	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "signer", msg.GetSigners())
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgYggdrasil", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	ygg, err := keeper.GetYggdrasil(ctx, msg.PubKey)
	if nil != err && !stdErrors.Is(err, ErrYggdrasilNotFound) {
		ctx.Logger().Error("fail to get yggdrasil", err)
		return sdk.ErrInternal("fail to get yggdrasil").Result()
	}
	if msg.AddFunds {
		ygg.AddFunds(msg.Coins)
	} else {
		ygg.SubFunds(msg.Coins)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent("yggdrasil_return",
				sdk.NewAttribute("pubkey", ygg.PubKey.String()),
				sdk.NewAttribute("coins", msg.Coins.String()),
				sdk.NewAttribute("tx", msg.RequestTxHash.String())))

		na, err := keeper.GetNodeAccountByPubKey(ctx, msg.PubKey)
		if err != nil {
			ctx.Logger().Error("unable to get node account", "error", err)
			return sdk.ErrUnknownRequest(err.Error()).Result()
		}
		// TODO: slash their bond for any Yggdrasil funds that are unaccounted
		// for before sending their bond back. Keep in mind that THORNode won't get
		// back 100% of the funds (due to gas).
		RefundBond(ctx, msg.RequestTxHash, na, keeper, txOut)
	}
	if err := keeper.SetYggdrasil(ctx, ygg); nil != err {
		ctx.Logger().Error("fail to save yggdrasil", err)
		return sdk.ErrInternal("fail to save yggdrasil").Result()
	}

	// Ragnarok protocol get triggered, if all the Yggdrasil pool returned funds already, THORNode will continue Ragnarok
	if validatorMgr.Meta.Ragnarok {
		hasYggdrasilPool, err := keeper.HasValidYggdrasilPools(ctx)
		if nil != err {
			return sdk.ErrInternal(fmt.Errorf("fail to check yggdrasil pools: %w", err).Error()).Result()
		}
		if !hasYggdrasilPool {
			return handleRagnarokProtocolStep2(ctx, keeper, txOut, poolAddrMgr, validatorMgr)
		}
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func handleMsgLeave(ctx sdk.Context, keeper Keeper, txOut *TxOutStore, validatorManager *ValidatorManager, msg MsgLeave) sdk.Result {
	ctx.Logger().Info("receive MsgLeave", "sender", msg.Tx.FromAddress.String(), "request tx hash", msg.Tx.ID)
	if !isSignedByActiveObserver(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "signer", msg.GetSigners())
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgLeave", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	nodeAcc, err := keeper.GetNodeAccountByBondAddress(ctx, msg.Tx.FromAddress)
	if nil != err {
		ctx.Logger().Error("fail to get node account", "error", err)
		return sdk.ErrInternal("fail to get node account by bond address").Result()
	}
	if nodeAcc.IsEmpty() {
		return sdk.ErrUnknownRequest("node account doesn't exist").Result()
	}

	if nodeAcc.Status == NodeActive {
		// THORNode add the node to leave queue
		validatorManager.Meta.LeaveQueue = append(validatorManager.Meta.LeaveQueue, nodeAcc)
	} else {
		// node is not active , they are free to leave , refund them
		// given the node is not active, they should not have Yggdrasil pool either
		RefundBond(ctx, msg.Tx.ID, nodeAcc, keeper, txOut)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent("validator_request_leave",
			sdk.NewAttribute("signer bnb address", msg.Tx.FromAddress.String()),
			sdk.NewAttribute("destination", nodeAcc.BondAddress.String()),
			sdk.NewAttribute("tx", msg.Tx.ID.String())))

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func handleRagnarokProtocolStep2(ctx sdk.Context, keeper Keeper, txOut *TxOutStore, poolAddrMgr *PoolAddressManager, validatorManager *ValidatorManager) sdk.Result {
	// Ragnarok Protocol
	// If THORNode can no longer be BFT, do a graceful shutdown of the entire network.
	// 1) THORNode will request all yggdrasil pool to return fund , if THORNode don't have yggdrasil pool THORNode will go to step 3 directly
	// 2) upon receiving the yggdrasil fund,  THORNode will refund the validator's bond
	// 3) once all yggdrasil fund get returned, return all fund to stakes
	if !validatorManager.Meta.Ragnarok {
		// Ragnarok protocol didn't triggered , don't call this one
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}
	// get the first observer
	nas, err := keeper.ListActiveNodeAccounts(ctx)
	if nil != err {
		ctx.Logger().Error("can't get active nodes", err)
		return sdk.ErrInternal("can't get active nodes").Result()
	}
	if len(nas) == 0 {
		return sdk.ErrInternal("can't find any active nodes").Result()
	}
	poolIndexes, err := keeper.GetPoolIndex(ctx)
	if nil != err {
		ctx.Logger().Error("fail to get pool index", "err", err)
		return sdk.ErrInternal("fail to get pool index").Result()
	}
	// go through all the pooles
	for _, pi := range poolIndexes {
		poolStaker, err := keeper.GetPoolStaker(ctx, pi)
		if nil != err {
			ctx.Logger().Error("fail to get pool staker", err)
			return sdk.ErrInternal(err.Error()).Result()
		}

		// everyone withdraw
		for _, item := range poolStaker.Stakers {
			unstakeMsg := NewMsgSetUnStake(
				common.GetRagnarokTx(pi.Chain),
				item.RuneAddress,
				sdk.NewUint(10000),
				pi,
				nas[0].NodeAddress,
			)

			result := handleMsgSetUnstake(ctx, keeper, txOut, poolAddrMgr, unstakeMsg)
			if !result.IsOK() {
				ctx.Logger().Error("fail to unstake", "staker", item.RuneAddress)
				return result
			}
		}
		keeper.SetPoolData(
			ctx,
			pi,
			PoolSuspended)
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}

}
