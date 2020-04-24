package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
	. "gopkg.in/check.v1"
)

var _ = Suite(&HandlerSwitchSuite{})

type HandlerSwitchSuite struct{}

func (s *HandlerSwitchSuite) TestValidate(c *C) {
	ctx, k := setupKeeperForTest(c)

	na := GetRandomNodeAccount(NodeActive)
	c.Assert(k.SetNodeAccount(ctx, na), IsNil)
	tx := GetRandomTx()
	tx.Coins = common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
	}
	destination := GetRandomBech32Addr()

	handler := NewSwitchHandler(k)
	// happy path
	msg := NewMsgSwitch(tx, destination, na.NodeAddress)
	err := handler.validate(ctx, msg, constants.SWVersion)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, errBadVersion)

	// invalid msg
	msg = MsgSwitch{}
	err = handler.validate(ctx, msg, constants.SWVersion)
	c.Assert(err, NotNil)
}

func (s *HandlerSwitchSuite) TestHandle(c *C) {
	ctx, k := setupKeeperForTest(c)

	na := GetRandomNodeAccount(NodeActive)
	c.Assert(k.SetNodeAccount(ctx, na), IsNil)
	tx := GetRandomTx()
	tx.Coins = common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
	}
	destination := GetRandomBech32Addr()

	handler := NewSwitchHandler(k)

	msg := NewMsgSwitch(tx, destination, na.NodeAddress)
	result := handler.handle(ctx, msg, constants.SWVersion)
	c.Assert(result.IsOK(), Equals, true, Commentf("%+v", result.Log))
	coin, err := common.NewCoin(common.RuneNative, sdk.NewUint(100*common.One)).Native()
	c.Assert(err, IsNil)
	c.Check(k.CoinKeeper().HasCoins(ctx, destination, sdk.NewCoins(coin)), Equals, true)

	// check that we can add more an account
	result = handler.handle(ctx, msg, constants.SWVersion)
	c.Assert(result.IsOK(), Equals, true, Commentf("%+v", result.Log))
	coin, err = common.NewCoin(common.RuneNative, sdk.NewUint(200*common.One)).Native()
	c.Assert(err, IsNil)
	c.Check(k.CoinKeeper().HasCoins(ctx, destination, sdk.NewCoins(coin)), Equals, true)
}
