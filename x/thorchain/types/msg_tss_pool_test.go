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
	msg := NewMsgTssPool(pks, pk, common.EmptyBlame, addr)
	c.Check(msg.Type(), Equals, "set_tss_pool")
	c.Assert(msg.ValidateBasic(), IsNil)

	c.Check(NewMsgTssPool(pks, pk, common.EmptyBlame, nil).ValidateBasic(), NotNil)
	c.Check(NewMsgTssPool(nil, pk, common.EmptyBlame, addr).ValidateBasic(), NotNil)
	c.Check(NewMsgTssPool(pks, "", common.EmptyBlame, addr).ValidateBasic(), NotNil)
	c.Check(NewMsgTssPool(pks, "bogusPubkey", common.EmptyBlame, addr).ValidateBasic(), NotNil)
}
