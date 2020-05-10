package types

import (
	blame "gitlab.com/thorchain/tss/go-tss/blame"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type MsgTssPoolSuite struct{}

var _ = Suite(&MsgTssPoolSuite{})

func (s *MsgTssPoolSuite) TestMsgTssPool(c *C) {
	pk := GetRandomPubKey()
	pks := common.PubKeys{
		GetRandomPubKey(), GetRandomPubKey(), GetRandomPubKey(),
	}
	addr := GetRandomBech32Addr()
	msg := NewMsgTssPool(pks, pk, AsgardKeygen, 1, blame.Blame{}, common.Chains{common.RuneAsset().Chain}, addr)
	c.Check(msg.Type(), Equals, "set_tss_pool")
	c.Assert(msg.ValidateBasic(), IsNil)

	c.Check(NewMsgTssPool(pks, pk, AsgardKeygen, 1, blame.Blame{}, common.Chains{common.RuneAsset().Chain}, nil).ValidateBasic(), NotNil)
	c.Check(NewMsgTssPool(nil, pk, AsgardKeygen, 1, blame.Blame{}, common.Chains{common.RuneAsset().Chain}, addr).ValidateBasic(), NotNil)
	c.Check(NewMsgTssPool(pks, "", AsgardKeygen, 1, blame.Blame{}, common.Chains{common.RuneAsset().Chain}, addr).ValidateBasic(), NotNil)
	c.Check(NewMsgTssPool(pks, "bogusPubkey", AsgardKeygen, 1, blame.Blame{}, common.Chains{common.RuneAsset().Chain}, addr).ValidateBasic(), NotNil)

	// fails on empty chain list
	msg = NewMsgTssPool(pks, pk, AsgardKeygen, 1, blame.Blame{}, common.Chains{}, addr)
	c.Check(msg.ValidateBasic(), NotNil)
	// fails on duplicates in chain list
	msg = NewMsgTssPool(pks, pk, AsgardKeygen, 1, blame.Blame{}, common.Chains{common.RuneAsset().Chain, common.RuneAsset().Chain}, addr)
	c.Check(msg.ValidateBasic(), NotNil)
}
