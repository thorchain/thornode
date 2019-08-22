package swapservice

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "swapservice" type messages.
func NewHandler(keeper Keeper, txOutStore *TxOutStore) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch m := msg.(type) {
		case MsgSetPoolData:
			return handleMsgSetPoolData(ctx, keeper, m)
		case MsgSetStakeData:
			result := handleMsgSetStakeData(ctx, keeper, m)
			processRefund(ctx, &result, txOutStore, keeper, m)
			return result
		case MsgSwap:
			result := handleMsgSwap(ctx, keeper, txOutStore, m)
			processRefund(ctx, &result, txOutStore, keeper, m)
			return result
		case MsgAdd:
			return handleMsgAdd(ctx, keeper, m)
		case MsgSetUnStake:
			return handleMsgSetUnstake(ctx, keeper, txOutStore, m)
		case MsgSetTxIn:
			return handleMsgSetTxIn(ctx, keeper, txOutStore, m)
		case MsgSetAdminConfig:
			return handleMsgSetAdminConfig(ctx, keeper, m)
		case MsgOutboundTx:
			return handleMsgOutboundTx(ctx, keeper, m)
		case MsgNoOp:
			return handleMsgNoOp(ctx, keeper, m)
		default:
			errMsg := fmt.Sprintf("Unrecognized swapservice Msg type: %v", m)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func isSignedByTrustAccounts(ctx sdk.Context, keeper Keeper, signers []sdk.AccAddress) bool {
	if len(signers) == 0 {
		return false
	}
	for _, signer := range signers {
		if !keeper.IsTrustAccount(ctx, signer) {
			ctx.Logger().Error("unauthorized account", "address", signer.String())
			return false
		}
	}
	return true
}

// Handle a message to set pooldata
func handleMsgSetPoolData(ctx sdk.Context, keeper Keeper, msg MsgSetPoolData) sdk.Result {
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
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

// Handle a message to set stake data
func handleMsgSetStakeData(ctx sdk.Context, keeper Keeper, msg MsgSetStakeData) sdk.Result {
	ctx.Logger().Info("handleMsgSetStakeData request", "stakerid:"+msg.Ticker)
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "ticker", msg.Ticker, "request tx hash", msg.RequestTxHash, "public address", msg.PublicAddress)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := keeper.GetPool(ctx, msg.Ticker).EnsureValidPoolStatus(msg); nil != err {
		ctx.Logger().Error("check pool status", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
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

	stakeEvt := NewEventStake(
		msg.RuneAmount,
		msg.TokenAmount,
		stakeUnits,
	)
	stakeBytes, err := json.Marshal(stakeEvt)
	if err != nil {
		ctx.Logger().Error("fail to save event", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	evt := NewEvent(
		stakeEvt.Type(),
		msg.RequestTxHash,
		msg.Ticker,
		stakeBytes,
	)
	keeper.AddIncompleteEvents(ctx, evt)
	// since there is no outbound tx for staking, we'll complete the event now
	blankTxID, _ := common.NewTxID(
		"0000000000000000000000000000000000000000000000000000000000000000",
	)
	keeper.CompleteEvents(ctx, []common.TxID{msg.RequestTxHash}, blankTxID)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// Handle a message to set stake data
func handleMsgSwap(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore, msg MsgSwap) sdk.Result {
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account", "request tx hash", msg.RequestTxHash, "source ticker", msg.SourceTicker, "target ticker", msg.TargetTicker)
		return sdk.ErrUnauthorized("Not authorized").Result()
	}

	tsl := keeper.GetAdminConfigTSL(ctx, common.NoBnbAddress)
	gsl := keeper.GetAdminConfigGSL(ctx, common.NoBnbAddress)

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
		tsl,
		gsl,
	) // If so, set the stake data to the value specified in the msg.
	if err != nil {
		ctx.Logger().Error("fail to process swap message", "error", err)

		return sdk.ErrInternal(err.Error()).Result()
	}

	res, err := keeper.cdc.MarshalBinaryLengthPrefixed(struct {
		Token common.Amount `json:"token"`
	}{
		Token: amount,
	})
	if nil != err {
		ctx.Logger().Error("fail to encode result to json", "error", err)
		return sdk.ErrInternal("fail to encode result to json").Result()
	}

	toi := &TxOutItem{
		ToAddress: msg.Destination,
	}
	toi.Coins = append(toi.Coins, common.Coin{
		Denom:  msg.TargetTicker,
		Amount: amount,
	})
	txOutStore.AddTxOutItem(toi)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      res,
		Codespace: DefaultCodespace,
	}
}

// handleMsgSetUnstake process unstake
func handleMsgSetUnstake(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore, msg MsgSetUnStake) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgSetUnstake from : %s(%s) unstake (%s)", msg, msg.PublicAddress, msg.WithdrawBasisPoints))
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
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
		Rune  common.Amount `json:"rune"`
		Token common.Amount `json:"token"`
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
	)
	keeper.AddIncompleteEvents(ctx, evt)

	toi := &TxOutItem{
		ToAddress: msg.PublicAddress,
	}
	toi.Coins = append(toi.Coins, common.Coin{
		Denom:  common.RuneTicker,
		Amount: runeAmt,
	})
	toi.Coins = append(toi.Coins, common.Coin{
		Denom:  msg.Ticker,
		Amount: tokenAmount,
	})
	txOutStore.AddTxOutItem(toi)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      res,
		Codespace: DefaultCodespace,
	}
}

func refundTx(ctx sdk.Context, tx TxIn, store *TxOutStore, keeper RefundStoreAccessor) {
	toi := &TxOutItem{
		ToAddress: tx.Sender,
	}

	for _, item := range tx.Coins {
		c := getRefundCoin(ctx, item.Denom, item.Amount, keeper)
		if c.Amount.GreaterThen(0) {
			toi.Coins = append(toi.Coins, c)
		}
	}
	store.AddTxOutItem(toi)
}

// handleMsgSetTxIn gets a binance tx hash, gets the tx/memo, and triggers
// another handler to process the request
func handleMsgSetTxIn(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore, msg MsgSetTxIn) sdk.Result {
	conflicts := make(common.TxIDs, 0)
	todo := make([]TxInVoter, 0)
	for _, tx := range msg.TxIns {
		// validate there are not conflicts first
		if keeper.CheckTxHash(ctx, tx.Key()) {
			conflicts = append(conflicts, tx.Key())
		} else {
			todo = append(todo, tx)
		}
	}

	totalTrusted := keeper.TotalTrustAccounts(ctx)
	handler := NewHandler(keeper, txOutStore)
	for _, tx := range todo {
		voter := keeper.GetTxInVoter(ctx, tx.TxID)
		preConsensus := voter.HasConensus(totalTrusted)
		voter.Adds(tx.Txs, msg.Signer)
		postConsensus := voter.HasConensus(totalTrusted)
		keeper.SetTxInVoter(ctx, voter)

		// TODO: if we change the number of trusted accounts, this will double
		// spend
		if preConsensus == false && postConsensus == true {
			msg, err := processOneTxIn(ctx, keeper, tx.TxID, voter.GetTx(totalTrusted), msg.Signer)
			if nil != err {
				ctx.Logger().Error("fail to process txHash", "error", err)
				refundTx(ctx, voter.GetTx(totalTrusted), txOutStore, keeper)
				continue
			}

			handler(ctx, msg)
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
	case AdminMemo:
		newMsg, err = getMsgAdminConfigFromMemo(ctx, keeper, m, tx, signer)
		if nil != err {
			return nil, errors.Wrap(err, "fail to get MsgAdminConfig from memo")
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
	return NewMsgSwap(txID, coin.Denom, memo.GetTicker(), coin.Amount, tx.Sender, memo.Destination, common.NewAmountFromFloat(memo.SlipLimit), signer), nil
}

func getMsgUnstakeFromMemo(memo WithdrawMemo, txID common.TxID, tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	withdrawAmount := common.NewAmountFromFloat(MaxWithdrawBasisPoints)
	var err error
	if len(memo.GetAmount()) > 0 {
		withdrawAmount, err = common.NewAmount(memo.GetAmount())
		if nil != err {
			return nil, err
		}
	}
	return NewMsgSetUnStake(tx.Sender, withdrawAmount, memo.GetTicker(), txID, signer), nil

}

func getMsgAdminConfigFromMemo(ctx sdk.Context, keeper Keeper, memo AdminMemo, tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	switch memo.GetAdminType() {
	case adminPoolStatus:
		ticker, err := common.NewTicker(memo.GetKey())
		if err != nil {
			return nil, errors.Wrapf(err, "Memo: %+v", memo)
		}
		pool := keeper.GetPool(ctx, ticker)
		if pool.Empty() {
			return nil, fmt.Errorf("pool doesn't exist: %s", ticker.String())
		}
		if !keeper.IsTrustAccountBnb(ctx, tx.Sender) {
			return nil, errors.New("Not authorized")
		}
		status := GetPoolStatus(memo.GetValue())
		return NewMsgSetPoolData(
			ticker,
			status,
			signer,
		), nil
	case adminKey:
		key := GetAdminConfigKey(memo.GetKey())
		return NewMsgSetAdminConfig(key, memo.GetValue(), tx.Sender, signer), nil
	}
	return nil, errors.New("invalid admin command type")
}

func getMsgStakeFromMemo(ctx sdk.Context, memo StakeMemo, txID common.TxID, tx *TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	runeAmount := common.ZeroAmount
	tokenAmount := common.ZeroAmount
	for _, coin := range tx.Coins {
		ctx.Logger().Info("coin", "denom", coin.Denom.String(), "amount", coin.Amount.String())
		if common.IsRune(coin.Denom) {
			runeAmount = coin.Amount
		} else {
			tokenAmount = coin.Amount
		}
	}
	return NewMsgSetStakeData(
		memo.GetTicker(),
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
		PoolBootstrap, // new pools start in a Bootstrap state
		signer,
	), nil
}

func getMsgAddFromMemo(memo AddMemo, txID common.TxID, tx TxIn, signer sdk.AccAddress) (sdk.Msg, error) {
	runeAmount := common.ZeroAmount
	tokenAmount := common.ZeroAmount
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

// handleMsgAdd
func handleMsgAdd(ctx sdk.Context, keeper Keeper, msg MsgAdd) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgAdd %s", msg.TxID))
	fmt.Printf("Add: %+v\n", msg)
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account")
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgOutboundTx", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	pool := keeper.GetPool(ctx, msg.Ticker)
	if pool.Ticker.IsEmpty() {
		return sdk.ErrUnknownRequest(fmt.Sprintf("pool %s not exist", msg.Ticker)).Result()
	}
	if msg.TokenAmount.GreaterThen(0) {
		pool.BalanceToken = pool.BalanceToken.Plus(msg.TokenAmount)
	}
	if msg.RuneAmount.GreaterThen(0) {
		pool.BalanceRune = pool.BalanceRune.Plus(msg.RuneAmount)
	}

	keeper.SetPool(ctx, pool)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgNoOp doesn't do anything, its a no op
func handleMsgNoOp(ctx sdk.Context, keeper Keeper, msg MsgNoOp) sdk.Result {
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgOutboundTx processes outbound tx from our pool
func handleMsgOutboundTx(ctx sdk.Context, keeper Keeper, msg MsgOutboundTx) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgOutboundTx %s at height %d", msg.TxID, msg.Height))
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account")
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgOutboundTx", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	// ensure the bnb address this tx was sent from is from our pool
	poolAddress := keeper.GetAdminConfigPoolAddress(ctx, common.NoBnbAddress)
	if !poolAddress.Equals(msg.Sender) {
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
	txOut.Hash = msg.TxID
	keeper.SetTxOut(ctx, txOut)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// handleMsgSetAdminConfig process admin config
func handleMsgSetAdminConfig(ctx sdk.Context, keeper Keeper, msg MsgSetAdminConfig) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgSetAdminConfig %s --> %s", msg.AdminConfig.Key, msg.AdminConfig.Value))
	if !isSignedByTrustAccounts(ctx, keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account")
		return sdk.ErrUnauthorized("Not authorized").Result()
	}
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("invalid MsgSetAdminConfig", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	if !keeper.IsTrustAccountBnb(ctx, msg.From) {
		ctx.Logger().Error("message signed by unauthorized account")
		return sdk.ErrUnauthorized("Not authorized").Result()
	}

	keeper.SetAdminConfig(ctx, msg.AdminConfig)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
