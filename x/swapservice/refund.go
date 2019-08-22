package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
)

// RefundStoreAccessor define the method the is required for Refund operation
// We need this interface thus we can mock the behaviour and write unit test
type RefundStoreAccessor interface {
	GetAdminConfigMRRA(ctx sdk.Context, bnb common.BnbAddress) common.Amount
	GetPool(ctx sdk.Context, ticker common.Ticker) Pool
}

// processRefund take in the sdk.Result and decide whether we should refund customer
func processRefund(ctx sdk.Context, result *sdk.Result, store *TxOutStore, keeper RefundStoreAccessor, msg sdk.Msg) {
	if result.IsOK() {
		return
	}
	switch m := msg.(type) {
	case MsgSetStakeData:
		toi := &TxOutItem{
			ToAddress: m.PublicAddress,
		}
		c := getRefundCoin(ctx, common.RuneTicker, m.RuneAmount, keeper)
		c1 := getRefundCoin(ctx, m.Ticker, m.TokenAmount, keeper)
		if !c.Amount.GreaterThen(0) && !c1.Amount.GreaterThen(0) {
			reason := fmt.Sprintf("rune:%s,coin:%s both less than the minimum refund value", m.RuneAmount, m.TokenAmount)
			result.Events = result.Events.AppendEvent(
				sdk.NewEvent("no refund", sdk.NewAttribute("reason", reason)))
			// nothing to refund
			return
		}
		if c.Amount.GreaterThen(0) {
			toi.Coins = append(toi.Coins, c)
		}
		if c1.Amount.GreaterThen(0) {
			toi.Coins = append(toi.Coins, c1)
		}
		store.AddTxOutItem(toi)
	case MsgSwap:
		toi := &TxOutItem{
			ToAddress: m.Requester,
		}
		c := getRefundCoin(ctx, m.SourceTicker, m.Amount, keeper)
		if c.Amount.Equals(common.ZeroAmount) {
			reason := fmt.Sprintf("%s less than the minimum refund value", m.Amount)
			result.Events = result.Events.AppendEvent(
				sdk.NewEvent("no refund", sdk.NewAttribute("reason", reason)))
			// nothing to refund
			return
		}
		toi.Coins = append(toi.Coins, c)
		store.AddTxOutItem(toi)
	default:
		return
	}
}

// getRefundCoin
func getRefundCoin(ctx sdk.Context, ticker common.Ticker, amount common.Amount, keeper RefundStoreAccessor) common.Coin {
	minimumRefundRune := keeper.GetAdminConfigMRRA(ctx, common.NoBnbAddress)
	if common.IsRune(ticker) {
		if amount.Float64() > minimumRefundRune.Float64() {
			// refund the difference
			return common.NewCoin(ticker, common.NewAmountFromFloat(amount.Float64()-minimumRefundRune.Float64()))
		} else {
			return common.NewCoin(ticker, common.ZeroAmount)
		}
	}
	pool := keeper.GetPool(ctx, ticker)
	poolTokenPrice := pool.TokenPriceInRune()
	totalRuneAmt := amount.Float64() * poolTokenPrice
	if totalRuneAmt > minimumRefundRune.Float64() {
		return common.NewCoin(ticker, common.NewAmountFromFloat(amount.Float64()-poolTokenPrice*minimumRefundRune.Float64()))
	}
	return common.NewCoin(ticker, common.ZeroAmount)
}
