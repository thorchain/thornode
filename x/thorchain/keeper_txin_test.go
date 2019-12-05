package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type KeeperTxInSuite struct{}

var _ = Suite(&KeeperTxInSuite{})

func (s *KeeperTxInSuite) TestTxInVoter(c *C) {
	ctx, k := setupKeeperForTest(c)

	tx := GetRandomTx()
	voter := NewObservedTxVoter(
		tx.ID,
		ObservedTxs{NewObservedTx(tx, sdk.NewUint(12), GetRandomPubKey())},
	)

	k.SetObservedTxVoter(ctx, voter)
	voter, err := k.GetObservedTxVoter(ctx, voter.TxID)
	c.Assert(err, IsNil)
	c.Check(voter.TxID.Equals(tx.ID), Equals, true)
}

func (s *KeeperTxInSuite) TestTxInInex(c *C) {
	ctx, k := setupKeeperForTest(c)

	txID := GetRandomTxHash()
	k.AddToObservedTxIndex(ctx, 10, txID)
	k.AddToObservedTxIndex(ctx, 10, txID) // check it dedups appropriately
	index, err := k.GetObservedTxIndex(ctx, 10)
	c.Assert(err, IsNil)
	c.Assert(index, HasLen, 1)
	c.Check(index[0].Equals(txID), Equals, true)
}
