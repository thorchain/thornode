package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/common"
)

type ObservedTxOutHandler struct {
	keeper       Keeper
	txOutStore   TxOutStore
	poolAddrMgr  PoolAddressManager
	validatorMgr ValidatorManager
}

func NewObservedTxOutHandler(keeper Keeper, txOutStore TxOutStore, poolAddrMgr PoolAddressManager, validatorMgr ValidatorManager) ObservedTxOutHandler {
	return ObservedTxOutHandler{
		keeper:       keeper,
		txOutStore:   txOutStore,
		poolAddrMgr:  poolAddrMgr,
		validatorMgr: validatorMgr,
	}
}

func (h ObservedTxOutHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version) sdk.Result {
	msg, ok := m.(MsgObservedTxOut)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.Validate(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	if err := h.handle(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (h ObservedTxOutHandler) validate(ctx sdk.Context, msg MsgObservedTxOut, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h ObservedTxOutHandler) validateV1(ctx sdk.Context, msg MsgObservedTxOut) error {
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

func (h ObservedTxOutHandler) handle(ctx sdk.Context, msg MsgObservedTxOut, version semver.Version) error {
	ctx.Logger().Info("handleMsgObservedTxOut request", "Tx:", msg.Txs[0].String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h ObservedTxOutHandler) outboundFailure(ctx sdk.Context, tx ObservedTx, activeNodeAccounts NodeAccounts) error {
	if h.keeper.YggdrasilExists(ctx, tx.ObservedPubKey) {
		ygg, err := h.keeper.GetYggdrasil(ctx, tx.ObservedPubKey)
		if nil != err {
			ctx.Logger().Error("fail to get yggdrasil", err)
			return err
		}
		var expectedCoins common.Coins
		memo, _ := ParseMemo(tx.Tx.Memo)
		switch memo.GetType() {
		case txYggdrasilReturn:
			expectedCoins = ygg.Coins
		case txOutbound:
			txID := memo.GetTxID()
			inVoter, err := h.keeper.GetObservedTxVoter(ctx, txID)
			if err != nil {
				return err
			}
			origTx := inVoter.GetTx(activeNodeAccounts)
			expectedCoins = origTx.Tx.Coins
		}

		na, err := h.keeper.GetNodeAccountByPubKey(ctx, ygg.PubKey)
		if err != nil {
			ctx.Logger().Error("fail to get node account", "error", err, "txhash", tx.Tx.ID.String())
			return err
		}

		// Slash the node account, since THORNode are unable to
		// process the tx (ie unscheduled tx)
		var minusCoins common.Coins        // track funds to subtract from ygg pool
		minusRune := sdk.ZeroUint()        // track amt of rune to slash from bond
		for _, coin := range tx.Tx.Coins { // assumes coins are asset uniq
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
					pool, err := h.keeper.GetPool(ctx, coin.Asset)
					if err != nil {
						return err
					}

					if !pool.Empty() {
						minusRune = pool.AssetValueInRune(diff)
						minusCoins = append(minusCoins, common.NewCoin(coin.Asset, diff))
						// Update pool balances
						pool.BalanceRune = pool.BalanceRune.Add(minusRune)
						pool.BalanceAsset = common.SafeSub(pool.BalanceAsset, diff)
						if err := h.keeper.SetPool(ctx, pool); err != nil {
							err = errors.Wrap(err, "fail to set pool")
							ctx.Logger().Error(err.Error())
							return err
						}
					}
				}
			}
		}
		na.SubBond(minusRune)
		if err := h.keeper.SetNodeAccount(ctx, na); nil != err {
			ctx.Logger().Error(fmt.Sprintf("fail to save node account(%s)", na), err)
			return err
		}
		ygg.SubFunds(minusCoins)
		if err := h.keeper.SetYggdrasil(ctx, ygg); nil != err {
			ctx.Logger().Error("fail to save yggdrasil", err)
			return err
		}
	}

	fmt.Println("DONE.")
	return nil
}

func (h ObservedTxOutHandler) preflight(ctx sdk.Context, voter ObservedTxVoter, nas NodeAccounts, tx ObservedTx, signer sdk.AccAddress) (ObservedTxVoter, bool) {
	voter.Add(tx, signer)

	ok := false
	if voter.HasConensus(nas) && voter.Height == 0 {
		ok = true
		voter.Height = ctx.BlockHeight()
	}
	h.keeper.SetObservedTxVoter(ctx, voter)

	// Check to see if we have enough identical observations to process the transaction
	return voter, ok
}

// Handle a message to observe inbound tx
func (h ObservedTxOutHandler) handleV1(ctx sdk.Context, msg MsgObservedTxOut) error {
	activeNodeAccounts, err := h.keeper.ListActiveNodeAccounts(ctx)
	if nil != err {
		return wrapError(ctx, err, "fail to get list of active node accounts")
	}

	handler := NewHandler(h.keeper, h.poolAddrMgr, h.txOutStore, h.validatorMgr)

	for _, tx := range msg.Txs {
		voter, err := h.keeper.GetObservedTxVoter(ctx, tx.Tx.ID)
		if err != nil {
			fmt.Printf("Err1 %s\n", err)
			return err
		}

		if voter, ok := h.preflight(ctx, voter, activeNodeAccounts, tx, msg.Signer); ok {
			txOut := voter.GetTx(activeNodeAccounts) // get consensus tx, in case our for loop is incorrect
			if ok := isCurrentVaultPubKey(ctx, h.keeper, h.poolAddrMgr, txOut); !ok {
				if err := refundTx(ctx, txOut, h.txOutStore, h.keeper, false); err != nil {
					fmt.Printf("Err2 %s\n", err)
					return err
				}
				fmt.Println("continue 1")
				continue
			}

			m, err := processOneTxIn(ctx, h.keeper, txOut, msg.Signer)
			if nil != err || tx.Tx.Chain.IsEmpty() {
				ctx.Logger().Error("fail to process txOut", "error", err, "txhash", tx.Tx.ID.String())
				// Detect if the txOut is to the thorchain network or from the
				// thorchain network
				if err := h.outboundFailure(ctx, txOut, activeNodeAccounts); err != nil {
					fmt.Printf("Err3 %s\n", err)
					return err
				}
				fmt.Println("continue 2")
				continue
			}

			// add addresses to observing addresses. This is used to detect
			// active/inactive observing node accounts
			if err := h.keeper.AddObservingAddresses(ctx, txOut.Signers); err != nil {
				fmt.Printf("Err4 %s\n", err)
				return err
			}

			result := handler(ctx, m)
			if !result.IsOK() {
				if err := refundTx(ctx, txOut, h.txOutStore, h.keeper, true); err != nil {
					fmt.Printf("Err5 %s\n", err)
					return err
				}
				fmt.Printf("Non-zero result: %+v\n", result)
			}
		}
	}

	fmt.Printf("DONEEE>")
	return nil
}
