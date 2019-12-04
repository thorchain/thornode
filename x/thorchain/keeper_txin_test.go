package thorchain

import (
	. "gopkg.in/check.v1"
)

type KeeperTxInSuite struct{}

var _ = Suite(&KeeperTxInSuite{})

func (s *KeeperTxInSuite) TestTxInVoter(c *C) {
	ctx, k := setupKeeperForTest(c)

	txID := GetRandomTxHash()
	voter := NewTxInVoter(txID, nil)

	k.SetTxInVoter(ctx, voter)
	voter, err := k.GetTxInVoter(ctx, voter.TxID)
	c.Assert(err, IsNil)
	c.Check(voter.TxID.Equals(txID), Equals, true)
}

func (s *KeeperTxInSuite) TestTxInInex(c *C) {
	ctx, k := setupKeeperForTest(c)

	txID := GetRandomTxHash()
	k.AddToTxInIndex(ctx, 10, txID)
	k.AddToTxInIndex(ctx, 10, txID) // check it dedups appropriately
	index, err := k.GetTxInIndex(ctx, 10)
	c.Assert(err, IsNil)
	c.Assert(index, HasLen, 1)
	c.Check(index[0].Equals(txID), Equals, true)
}
