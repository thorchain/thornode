package types

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type TypeTssSuite struct{}

var _ = Suite(&TypeTssSuite{})

func (s *TypeTssSuite) TestVoter(c *C) {
	pk := GetRandomPubKey()
	pks := common.PubKeys{
		GetRandomPubKey(), GetRandomPubKey(), GetRandomPubKey(),
	}
	tss := NewTssVoter(
		"hello",
		pks,
		pk,
	)
	c.Check(tss.Empty(), Equals, false)

	chains := common.Chains{common.BNBChain, common.BTCChain}

	addr := GetRandomBech32Addr()
	c.Check(tss.HasSigned(addr), Equals, false)
	tss.Sign(addr, chains)
	c.Check(tss.Signers, HasLen, 1)
	c.Check(tss.HasSigned(addr), Equals, true)
	tss.Sign(addr, chains) // ensure signing twice doesn't duplicate
	c.Check(tss.Signers, HasLen, 1)
	c.Check(tss.Chains, HasLen, 2)

	c.Check(tss.HasConsensus(nil), Equals, false)
	nas := NodeAccounts{
		NodeAccount{NodeAddress: addr, Status: Active},
	}
	c.Check(tss.HasConsensus(nas), Equals, true)
}
