package thorchain

import (
	. "gopkg.in/check.v1"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

func FundModule(c *C, ctx sdk.Context, k Keeper, name string, amt uint64) {
	coin, err := common.NewCoin(common.RuneNative, sdk.NewUint(amt*common.One)).Native()
	c.Assert(err, IsNil)
	err = k.Supply().MintCoins(ctx, ModuleName, sdk.NewCoins(coin))
	c.Assert(err, IsNil)
	err = k.Supply().SendCoinsFromModuleToModule(ctx, ModuleName, name, sdk.NewCoins(coin))
	c.Assert(err, IsNil)
}

func FundAccount(c *C, ctx sdk.Context, k Keeper, addr sdk.AccAddress, amt uint64) {
	coin, err := common.NewCoin(common.RuneNative, sdk.NewUint(amt*common.One)).Native()
	c.Assert(err, IsNil)
	err = k.Supply().MintCoins(ctx, ModuleName, sdk.NewCoins(coin))
	c.Assert(err, IsNil)
	err = k.Supply().SendCoinsFromModuleToAccount(ctx, ModuleName, addr, sdk.NewCoins(coin))
	c.Assert(err, IsNil)
}
