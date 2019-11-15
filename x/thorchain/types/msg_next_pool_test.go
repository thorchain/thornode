package types

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type MsgNextPoolTestSuite struct{}

var _ = Suite(&MsgNextPoolTestSuite{})

func (MsgNextPoolTestSuite) TestMsgNextPool(c *C) {
	sender := GetRandomBNBAddress()
	bnb := GetRandomPubKey()
	addr := GetRandomBech32Addr()
	c.Check(addr.Empty(), Equals, false)
	tx := GetRandomTx()
	msgNextPool := NewMsgNextPoolAddress(tx, bnb, sender, common.BNBChain, addr)
	c.Assert(msgNextPool.Type(), Equals, "set_next_pooladdress")
	EnsureMsgBasicCorrect(msgNextPool, c)
	msgNextPool2 := NewMsgNextPoolAddress(tx, common.EmptyPubKey, sender, common.BNBChain, addr)
	c.Assert(msgNextPool2.ValidateBasic(), NotNil)
	msgNextPool3 := NewMsgNextPoolAddress(tx, bnb, "", common.BNBChain, addr)
	c.Assert(msgNextPool3.ValidateBasic(), NotNil)
	msgEmptyChain := NewMsgNextPoolAddress(tx, bnb, sender, common.EmptyChain, addr)
	c.Assert(msgEmptyChain.ValidateBasic(), NotNil)
	tx.ID = ""
	msgNextPool1 := NewMsgNextPoolAddress(tx, bnb, sender, common.BNBChain, addr)
	c.Assert(msgNextPool1.ValidateBasic(), NotNil)
}
