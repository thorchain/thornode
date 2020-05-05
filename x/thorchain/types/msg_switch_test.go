package types

import (
	. "gopkg.in/check.v1"

	sdk "github.com/cosmos/cosmos-sdk/types"

	common "gitlab.com/thorchain/thornode/common"
)

type MsgSwitchSuite struct{}

var _ = Suite(&MsgSwitchSuite{})

func (MsgSwitchSuite) TestMsgSwitchSuite(c *C) {
	tx := GetRandomTx()
	tx.Coins = common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
	}

	acc1 := GetRandomBNBAddress()
	acc2 := GetRandomBech32Addr()

	c.Assert(acc1.IsEmpty(), Equals, false)
	msg := NewMsgSwitch(tx, acc1, acc2)
	c.Assert(msg.Route(), Equals, RouterKey)
	c.Assert(msg.Type(), Equals, "switch")
	c.Assert(msg.ValidateBasic(), IsNil)
	c.Assert(len(msg.GetSignBytes()) > 0, Equals, true)
	c.Assert(msg.GetSigners(), NotNil)
	c.Assert(msg.GetSigners()[0].String(), Equals, acc2.String())

	// test too many coins
	tx.Coins = common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(100*common.One)),
	}
	msg = NewMsgSwitch(tx, acc1, acc2)
	c.Assert(msg.ValidateBasic(), NotNil)

	// test too little coins
	tx.Coins = common.Coins{}
	msg = NewMsgSwitch(tx, acc1, acc2)
	c.Assert(msg.ValidateBasic(), NotNil)

	// test non rune token
	tx.Coins = common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
	}
	msg = NewMsgSwitch(tx, acc1, acc2)
	c.Assert(msg.ValidateBasic(), NotNil)
}
