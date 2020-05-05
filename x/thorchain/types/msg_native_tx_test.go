package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	common "gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgNativeTxSuite struct{}

var _ = Suite(&MsgNativeTxSuite{})

func (MsgNativeTxSuite) TestMsgNativeTxSuite(c *C) {
	acc1 := GetRandomBech32Addr()
	c.Assert(acc1.Empty(), Equals, false)

	coins := common.Coins{
		common.NewCoin(common.RuneNative, sdk.NewUint(12*common.One)),
	}
	memo := "hello"
	msg := NewMsgNativeTx(coins, memo, acc1)
	c.Assert(msg.Route(), Equals, RouterKey)
	c.Assert(msg.Type(), Equals, "native_tx")
	c.Assert(msg.ValidateBasic(), IsNil)
	c.Assert(len(msg.GetSignBytes()) > 0, Equals, true)
	c.Assert(msg.GetSigners(), NotNil)
	c.Assert(msg.GetSigners()[0].String(), Equals, acc1.String())

	// ensure non-native assets are blocked
	coins = common.Coins{
		common.NewCoin(common.BTCAsset, sdk.NewUint(12*common.One)),
	}
	msg = NewMsgNativeTx(coins, memo, acc1)
	c.Assert(msg.ValidateBasic(), NotNil)
}
