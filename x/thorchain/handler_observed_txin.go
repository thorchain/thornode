package thorchain

import (
	"encoding/json"
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/common"
)

type ObservedTxInHandler struct {
	keeper       Keeper
	txOutStore   *TxOutStore
	poolAddrMgr  *PoolAddressManager
	validatorMgr *ValidatorManager
}

func NewObservedTxInHandler(keeper Keeper, txOutStore *TxOutStore, poolAddrMgr *PoolAddressManager, validatorMgr *ValidatorManager) ObservedTxInHandler {
	return ObservedTxInHandler{
		keeper:       keeper,
		txOutStore:   txOutStore,
		poolAddrMgr:  poolAddrMgr,
		validatorMgr: validatorMgr,
	}
}

func (h ObservedTxInHandler) Run(ctx sdk.Context, msg MsgObservedTxIn, version semver.Version) sdk.Result {
	if err := h.Validate(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	if err := h.Handle(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (h ObservedTxInHandler) Validate(ctx sdk.Context, msg MsgObservedTxIn, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.ValidateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h ObservedTxInHandler) ValidateV1(ctx sdk.Context, msg MsgObservedTxIn) error {
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

func (h ObservedTxInHandler) Handle(ctx sdk.Context, msg MsgObservedTxIn, version semver.Version) error {
	ctx.Logger().Info("handleMsgObservedTxIn request", "Tx:", msg.Txs[0].String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.HandleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

// Handle a message to observe inbound tx
func (h ObservedTxInHandler) HandleV1(ctx sdk.Context, msg MsgObservedTxIn) error {
	activeNodeAccounts, err := h.keeper.ListActiveNodeAccounts(ctx)
	if nil != err {
		return wrapError(ctx, err, "fail to get list of active node accounts")
	}

	handler := NewHandler(h.keeper, h.poolAddrMgr, h.txOutStore, h.validatorMgr)

	for _, tx := range msg.Txs {
		voter := h.keeper.GetTxInVoter(ctx, tx.ID)
		preConsensus := voter.HasConensus(activeNodeAccounts)
		voter.Add(tx, msg.Signer)
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
				err := refundTx(ctx, voter.TxID, txIn, txOutStore, keeper, txIn.ObservePoolAddress, chain, false)
				if err != nil {
					err = errors.Wrap(err, "Fail to refund")
					ctx.Logger().Error(err.Error())
					return sdk.ErrInternal(err.Error()).Result()
				}

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
									pool, err := keeper.GetPool(ctx, coin.Asset)
									if err != nil {
										return sdk.ErrInternal(err.Error()).Result()
									}

									if !pool.Empty() {
										minusRune = pool.AssetValueInRune(diff)
										minusCoins = append(minusCoins, common.NewCoin(coin.Asset, diff))
										// Update pool balances
										pool.BalanceRune = pool.BalanceRune.Add(minusRune)
										pool.BalanceAsset = common.SafeSub(pool.BalanceAsset, diff)
										if err := keeper.SetPool(ctx, pool); err != nil {
											err = errors.Wrap(err, "fail to set pool")
											ctx.Logger().Error(err.Error())
											return sdk.ErrInternal(err.Error()).Result()
										}
									}
								}
							}
						}
						na.SubBond(minusRune)
						if err := keeper.SetNodeAccount(ctx, na); nil != err {
							ctx.Logger().Error(fmt.Sprintf("fail to save node account(%s)", na), err)
							return sdk.ErrInternal("fail to save node account").Result()
						}
						ygg.SubFunds(minusCoins)
						if err := keeper.SetYggdrasil(ctx, ygg); nil != err {
							ctx.Logger().Error("fail to save yggdrasil", err)
							return sdk.ErrInternal("fail to save yggdrasil").Result()
						}
					}
				} else {
					// To thorchain network
					err := refundTx(ctx, voter.TxID, txIn, txOutStore, keeper, currentPoolAddress.PubKey, currentPoolAddress.Chain, true)
					if err != nil {
						err = errors.Wrap(err, "Fail to refund")
						ctx.Logger().Error(err.Error())
						return sdk.ErrInternal(err.Error()).Result()
					}
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
			if err := keeper.AddObservingAddresses(ctx, txIn.Signers); err != nil {
				err = errors.Wrap(err, "fail to add observer address")
				ctx.Logger().Error(err.Error())
				return sdk.ErrInternal(err.Error()).Result()
			}

			result := handler(ctx, m)
			if !result.IsOK() {
				err := refundTx(ctx, voter.TxID, txIn, txOutStore, keeper, currentPoolAddress.PubKey, currentPoolAddress.Chain, true)
				if err != nil {
					err = errors.Wrap(err, "Fail to refund")
					ctx.Logger().Error(err.Error())
					return sdk.ErrInternal(err.Error()).Result()
				}
			}
		}
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
