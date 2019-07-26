package swapservice

import (
	"math"
	"testing"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

func TestSwapCalculation(t *testing.T) {
	inputs := []struct {
		name              string
		source            string
		runeBalance       float64
		tokenBalance      float64
		amountToSwap      float64
		runeBalanceAfter  float64
		tokenBalanceAfter float64
		amountToReturn    float64
	}{
		{
			name:              "normal",
			source:            types.RuneTicker,
			runeBalance:       100.0,
			tokenBalance:      100.0,
			amountToSwap:      5.0,
			runeBalanceAfter:  105.0,
			tokenBalanceAfter: 95.46,
			amountToReturn:    4.54,
		},
		{
			name:              "normal-token",
			source:            "BNB",
			runeBalance:       100.0,
			tokenBalance:      100.0,
			amountToSwap:      5.0,
			runeBalanceAfter:  95.46,
			tokenBalanceAfter: 105.0,
			amountToReturn:    4.54,
		},
	}
	for _, item := range inputs {
		t.Run(item.name, func(st *testing.T) {
			r, t, a := calculateSwap(item.source, item.runeBalance, item.tokenBalance, item.amountToSwap)
			if round(r) != item.runeBalanceAfter || round(t) != item.tokenBalanceAfter || round(a) != item.amountToReturn {
				st.Errorf("expected rune balance after: %f,token balance:%f ,amount:%f, however we got rune balance:%f,token balance:%f,amount:%f", item.runeBalanceAfter, item.tokenBalanceAfter, item.amountToReturn, r, t, a)
			}
		})
	}
}
func round(input float64) float64 {
	return math.Round(input*100) / 100
}
