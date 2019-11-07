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
	txID := GetRandomTxHash()
	msgNextPool := NewMsgNextPoolAddress(txID, bnb, sender, common.BNBChain, addr)
	c.Assert(msgNextPool.Type(), Equals, "set_next_pooladdress")
	EnsureMsgBasicCorrect(msgNextPool, c)
	msgNextPool1 := NewMsgNextPoolAddress("", bnb, sender, common.BNBChain, addr)
	c.Assert(msgNextPool1.ValidateBasic(), NotNil)
	msgNextPool2 := NewMsgNextPoolAddress(txID, common.EmptyPubKey, sender, common.BNBChain, addr)
	c.Assert(msgNextPool2.ValidateBasic(), NotNil)
	msgNextPool3 := NewMsgNextPoolAddress(txID, bnb, "", common.BNBChain, addr)
	c.Assert(msgNextPool3.ValidateBasic(), NotNil)
	msgEmptyChain := NewMsgNextPoolAddress(txID, bnb, sender, common.EmptyChain, addr)
	c.Assert(msgEmptyChain.ValidateBasic(), NotNil)
}
