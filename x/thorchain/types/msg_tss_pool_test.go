package types

import (
	common "gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgTssPoolSuite struct{}

var _ = Suite(&MsgTssPoolSuite{})

func (s *MsgTssPoolSuite) TestMsgTssPool(c *C) {
	pk := GetRandomPubKey()
	pks := common.PubKeys{
		GetRandomPubKey(), GetRandomPubKey(), GetRandomPubKey(),
	}
	addr := GetRandomBech32Addr()
	msg := NewMsgTssPool(pks, pk, 10, addr)
	c.Check(msg.Type(), Equals, "set_tss_pool")
	c.Assert(msg.ValidateBasic(), IsNil)

	c.Check(NewMsgTssPool(pks, pk, 0, addr).ValidateBasic(), NotNil)
	c.Check(NewMsgTssPool(pks, pk, 10, nil).ValidateBasic(), NotNil)
	c.Check(NewMsgTssPool(nil, pk, 10, addr).ValidateBasic(), NotNil)
	c.Check(NewMsgTssPool(pks, "", 10, addr).ValidateBasic(), NotNil)
	c.Check(NewMsgTssPool(pks, "bogusPubkey", 10, addr).ValidateBasic(), NotNil)
}
