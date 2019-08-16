package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	common "gitlab.com/thorchain/bepswap/common"
)

// FromSdkCoins convert the sdk.Coins type to our own coin type
func FromSdkCoins(c sdk.Coins) (common.Coins, error) {
	var cs common.Coins
	for _, item := range c {
		t, err := common.NewTicker(item.Denom)
		if nil != err {
			return nil, errors.Wrapf(err, "fail to convert sdk.Coin to statechain Coin type,ticker:%s invalid", item.Denom)
		}
		a, err := common.NewAmount(item.Amount.String())
		if nil != err {
			return nil, errors.Wrapf(err, "fail to convert amount %s ", item.Amount)
		}
		cs = append(cs, common.NewCoin(t, a))
	}
	return cs, nil
}
