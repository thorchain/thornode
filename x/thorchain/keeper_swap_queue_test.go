package thorchain

import (
	. "gopkg.in/check.v1"
)

type KeeperSwapQueueSuite struct{}

var _ = Suite(&KeeperSwapQueueSuite{})

func (s *KeeperSwapQueueSuite) TestKeeperSwapQueue(c *C) {
	ctx, k := setupKeeperForTest(c)

	// not found
	_, err := k.GetSwapQueueItem(ctx, GetRandomTxHash())
	c.Assert(err, NotNil)

	msg := MsgSwap{
		Tx: GetRandomTx(),
	}

	c.Assert(k.SetSwapQueueItem(ctx, msg), IsNil)
	msg2, err := k.GetSwapQueueItem(ctx, msg.Tx.ID)
	c.Assert(err, IsNil)
	c.Check(msg2.Tx.ID.Equals(msg.Tx.ID), Equals, true)

	iter := k.GetSwapQueueIterator(ctx)
	defer iter.Close()

	// test remove
	k.RemoveSwapQueueItem(ctx, msg.Tx.ID)
	_, err = k.GetSwapQueueItem(ctx, msg.Tx.ID)
	c.Check(err, NotNil)
}
