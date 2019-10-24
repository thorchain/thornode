package types

import (
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgEndPoolTestSuite struct{}

var _ = Suite(&MsgEndPoolTestSuite{})

func (MsgEndPoolTestSuite) TestMsgEndPool(c *C) {
	asset := common.BNBAsset
	bnb := GetRandomBNBAddress()
	addr := GetRandomBech32Addr()
	c.Check(addr.Empty(), Equals, false)
	txID := GetRandomTxHash()
	msgEndPool := NewMsgEndPool(asset, bnb, txID, addr)
	c.Assert(msgEndPool.Route(), Equals, RouterKey)
	c.Assert(msgEndPool.Type(), Equals, "set_poolend")
	c.Assert(msgEndPool.ValidateBasic(), IsNil)
	c.Assert(len(msgEndPool.GetSignBytes()) > 0, Equals, true)
	c.Assert(msgEndPool.GetSigners(), NotNil)
	c.Assert(msgEndPool.GetSigners()[0].String(), Equals, addr.String())

	errEndPool := NewMsgEndPool(common.Asset{}, bnb, txID, addr)
	c.Assert(errEndPool.ValidateBasic(), NotNil)
	errEndPool1 := NewMsgEndPool(common.RuneA1FAsset, bnb, txID, addr)
	c.Assert(errEndPool1.ValidateBasic(), NotNil)
	errEndPool2 := NewMsgEndPool(common.BNBAsset, bnb, "", addr)
	c.Assert(errEndPool2.ValidateBasic(), NotNil)
	errEndPool3 := NewMsgEndPool(common.BNBAsset, "", txID, addr)
	c.Assert(errEndPool3.ValidateBasic(), NotNil)

}
