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
		voter, err := h.keeper.GetObservedTxVoter(ctx, tx.Tx.ID)
		if err != nil {
			fmt.Printf("FOO1")
			return err
		}
		preConsensus := voter.HasConensus(activeNodeAccounts)
		voter.Add(tx, msg.Signer)
		postConsensus := voter.HasConensus(activeNodeAccounts)
		h.keeper.SetObservedTxVoter(ctx, voter)

		if preConsensus == false && postConsensus == true && voter.Height == 0 {
			voter.Height = ctx.BlockHeight()
			h.keeper.SetObservedTxVoter(ctx, voter)
			txIn := voter.GetTx(activeNodeAccounts)
			var chain common.Chain
			if len(txIn.Tx.Coins) > 0 {
				chain = txIn.Tx.Coins[0].Asset.Chain
			}

			currentPoolAddress := h.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(chain)
			yggExists := h.keeper.YggdrasilExists(ctx, txIn.ObservedPubKey)
			if !currentPoolAddress.PubKey.Equals(txIn.ObservedPubKey) && !yggExists {
				ctx.Logger().Error("wrong pool address,refund without deduct fee", "pubkey", currentPoolAddress.PubKey.String(), "observe pool addr", txIn.ObservedPubKey)
				err := refundTx(ctx, txIn, h.txOutStore, h.keeper, txIn.ObservedPubKey, chain, false)
				if err != nil {
					err = errors.Wrap(err, "Fail to refund")
					ctx.Logger().Error(err.Error())
					fmt.Printf("FOO2")
					return err
				}

				continue
			}

			m, err := processOneTxIn(ctx, h.keeper, txIn, msg.Signer)
			if nil != err || chain.IsEmpty() {
				ctx.Logger().Error("fail to process txIn", "error", err, "txhash", tx.Tx.ID.String())
				// Detect if the txIn is to the thorchain network or from the
				// thorchain network
				addr, err := txIn.ObservedPubKey.GetAddress(chain)
				if err != nil {
					ctx.Logger().Error("fail to get address", "error", err, "txhash", tx.Tx.ID.String())
					continue
				}

				if addr.Equals(txIn.Tx.FromAddress) {

					if h.keeper.YggdrasilExists(ctx, txIn.ObservedPubKey) {
						ygg, err := h.keeper.GetYggdrasil(ctx, txIn.ObservedPubKey)
						if nil != err {
							ctx.Logger().Error("fail to get yggdrasil", err)
							fmt.Printf("FOO3")
							return err
						}
						var expectedCoins common.Coins
						memo, _ := ParseMemo(txIn.Tx.Memo)
						switch memo.GetType() {
						case txYggdrasilReturn:
							expectedCoins = ygg.Coins
						case txOutbound:
							txID := memo.GetTxID()
							inVoter, err := h.keeper.GetObservedTxVoter(ctx, txID)
							if err != nil {
								fmt.Printf("FOO4")
								return err
							}
							origTx := inVoter.GetTx(activeNodeAccounts)
							expectedCoins = origTx.Tx.Coins
						}

						na, err := h.keeper.GetNodeAccountByPubKey(ctx, ygg.PubKey)
						if err != nil {
							ctx.Logger().Error("fail to get node account", "error", err, "txhash", tx.Tx.ID.String())
							fmt.Printf("FOO5")
							return err
						}

						// Slash the node account, since THORNode are unable to
						// process the tx (ie unscheduled tx)
						var minusCoins common.Coins          // track funds to subtract from ygg pool
						minusRune := sdk.ZeroUint()          // track amt of rune to slash from bond
						for _, coin := range txIn.Tx.Coins { // assumes coins are asset uniq
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
										fmt.Printf("FOO6")
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
											fmt.Printf("FOO7")
											return err
										}
									}
								}
							}
						}
						na.SubBond(minusRune)
						if err := h.keeper.SetNodeAccount(ctx, na); nil != err {
							ctx.Logger().Error(fmt.Sprintf("fail to save node account(%s)", na), err)
							fmt.Printf("FOO8")
							return err
						}
						ygg.SubFunds(minusCoins)
						if err := h.keeper.SetYggdrasil(ctx, ygg); nil != err {
							ctx.Logger().Error("fail to save yggdrasil", err)
							fmt.Printf("FOO9")
							return err
						}
					}
				} else {
					// To thorchain network
					err := refundTx(ctx, txIn, h.txOutStore, h.keeper, currentPoolAddress.PubKey, currentPoolAddress.Chain, true)
					if err != nil {
						err = errors.Wrap(err, "Fail to refund")
						ctx.Logger().Error(err.Error())
						fmt.Printf("FOO9")
						return err
					}
					ee := NewEmptyRefundEvent()
					buf, err := json.Marshal(ee)
					if nil != err {
						fmt.Printf("FOO10")
						return err
					}
					event := NewEvent(
						ee.Type(),
						ctx.BlockHeight(),
						tx.Tx,
						buf,
						EventRefund,
					)

					if err := h.keeper.AddIncompleteEvents(ctx, event); err != nil {
						fmt.Printf("FOO11")
						return err
					}
				}
				continue
			}

			// ignoring the error
			_ = h.keeper.AddToObservedTxIndex(ctx, uint64(ctx.BlockHeight()), tx.Tx.ID)
			if err := h.keeper.SetLastChainHeight(ctx, chain, txIn.BlockHeight); nil != err {
				fmt.Printf("FOO11")
				return err
			}

			// add this chain to our list of supported chains
			chains, err := h.keeper.GetChains(ctx)
			if err != nil {
				fmt.Printf("FOO12")
				return err
			}
			chains = append(chains, chain)
			h.keeper.SetChains(ctx, chains)

			// add addresses to observing addresses. This is used to detect
			// active/inactive observing node accounts
			if err := h.keeper.AddObservingAddresses(ctx, txIn.Signers); err != nil {
				err = errors.Wrap(err, "fail to add observer address")
				ctx.Logger().Error(err.Error())
				fmt.Printf("FOO13")
				return err
			}

			result := handler(ctx, m)
			if !result.IsOK() {
				err := refundTx(ctx, txIn, h.txOutStore, h.keeper, currentPoolAddress.PubKey, currentPoolAddress.Chain, true)
				if err != nil {
					err = errors.Wrap(err, "Fail to refund")
					ctx.Logger().Error(err.Error())
					fmt.Printf("FOO14")
					return err
				}
			}
		}
	}

	return nil
}
