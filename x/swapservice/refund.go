package swapservice

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
)

// RefundStoreAccessor define the method the is required for Refund operation
// We need this interface thus we can mock the behaviour and write unit test
type RefundStoreAccessor interface {
	GetAdminConfigMRRA(ctx sdk.Context, addr sdk.AccAddress) sdk.Uint
	GetPool(ctx sdk.Context, ticker common.Ticker) Pool
}

// getRefundCoin
func getRefundCoin(ctx sdk.Context, ticker common.Ticker, amount sdk.Uint, keeper RefundStoreAccessor) common.Coin {
	minimumRefundRune := keeper.GetAdminConfigMRRA(ctx, EmptyAccAddress)
	if common.IsRune(ticker) {
		if amount.GT(minimumRefundRune) {
			// refund the difference
			return common.NewCoin(ticker, amount.Sub(minimumRefundRune))
		} else {
			return common.NewCoin(ticker, sdk.ZeroUint())
		}
	}
	ctx.Logger().Debug("refund coin", "minimumRefundRune", minimumRefundRune)
	pool := keeper.GetPool(ctx, ticker)
	poolTokenPrice := pool.TokenPriceInRune()
	totalRuneAmt := sdk.NewUint(uint64(math.Round(float64(amount.Uint64()) * poolTokenPrice))) //amount.Mul(poolTokenPrice)
	ctx.Logger().Debug("refund coin", "pool price", poolTokenPrice, "total rune amount", totalRuneAmt)
	if totalRuneAmt.GT(minimumRefundRune) {
		tokenToRefund := common.UintToFloat64(totalRuneAmt.Sub(minimumRefundRune)) / poolTokenPrice

		return common.NewCoin(ticker, common.FloatToUint(tokenToRefund))
	}
	return common.NewCoin(ticker, sdk.ZeroUint())
}
