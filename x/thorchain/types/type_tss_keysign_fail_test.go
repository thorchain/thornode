package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/tss/go-tss/blame"
)

type TypeTssKeysignFailTestSuite struct{}

var _ = Suite(&TypeTssKeysignFailTestSuite{})

func (s *TypeTssKeysignFailTestSuite) TestVoter(c *C) {
	nodes := []blame.Node{
		blame.Node{Pubkey: GetRandomPubKey().String()},
		blame.Node{Pubkey: GetRandomPubKey().String()},
		blame.Node{Pubkey: GetRandomPubKey().String()},
	}
	b := blame.Blame{BlameNodes: nodes, FailReason: "fail to keysign"}
	m := NewMsgTssKeysignFail(1, b, "hello", common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100))}, GetRandomBech32Addr())
	tss := NewTssKeysignFailVoter(m.ID, 1)
	c.Check(tss.Empty(), Equals, false)

	addr := GetRandomBech32Addr()
	c.Check(tss.HasSigned(addr), Equals, false)
	tss.Sign(addr)
	c.Check(tss.Signers, HasLen, 1)
	c.Check(tss.HasSigned(addr), Equals, true)
	tss.Sign(addr) // ensure signing twice doesn't duplicate
	c.Check(tss.Signers, HasLen, 1)

	c.Check(tss.HasConsensus(nil), Equals, false)
	nas := NodeAccounts{
		NodeAccount{NodeAddress: addr, Status: Active},
	}
	c.Check(tss.HasConsensus(nas), Equals, true)
}
