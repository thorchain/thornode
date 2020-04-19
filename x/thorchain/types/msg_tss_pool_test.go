package types

import (
	tssCommon "gitlab.com/thorchain/tss/go-tss/common"
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
	msg := NewMsgTssPool(pks, pk, AsgardKeygen, 1, tssCommon.NoBlame, common.Chains{common.BNBChain}, addr)
	c.Check(msg.Type(), Equals, "set_tss_pool")
	c.Assert(msg.ValidateBasic(), IsNil)

	c.Check(NewMsgTssPool(pks, pk, AsgardKeygen, 1, tssCommon.NoBlame, common.Chains{common.BNBChain}, nil).ValidateBasic(), NotNil)
	c.Check(NewMsgTssPool(nil, pk, AsgardKeygen, 1, tssCommon.NoBlame, common.Chains{common.BNBChain}, addr).ValidateBasic(), NotNil)
	c.Check(NewMsgTssPool(pks, "", AsgardKeygen, 1, tssCommon.NoBlame, common.Chains{common.BNBChain}, addr).ValidateBasic(), NotNil)
	c.Check(NewMsgTssPool(pks, "bogusPubkey", AsgardKeygen, 1, tssCommon.NoBlame, common.Chains{common.BNBChain}, addr).ValidateBasic(), NotNil)
}
