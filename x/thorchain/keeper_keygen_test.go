package thorchain

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperKeygenSuite struct{}

var _ = Suite(&KeeperKeygenSuite{})

func (s *KeeperKeygenSuite) TestKeeperKeygen(c *C) {
	var err error
	ctx, k := setupKeeperForTest(c)

	keygenBlock := NewKeygenBlock(1)
	keygenMembers := common.PubKeys{GetRandomPubKey(), GetRandomPubKey(), GetRandomPubKey()}
	keygen, err := NewKeygen(ctx.BlockHeight(), keygenMembers, AsgardKeygen)
	c.Assert(err, IsNil)
	c.Assert(keygen.IsEmpty(), Equals, false)
	keygenBlock.Keygens = append(keygenBlock.Keygens, keygen)
	c.Assert(k.SetKeygenBlock(ctx, keygenBlock), IsNil)

	keygenBlock, err = k.GetKeygenBlock(ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(keygenBlock, NotNil)
	c.Assert(keygenBlock.Height, Equals, int64(1))

	keygenBlock, err = k.GetKeygenBlock(ctx, 100)
	c.Assert(err, IsNil)
	c.Assert(keygenBlock, NotNil)

	iter := k.GetKeygenBlockIterator(ctx)
	defer iter.Close()
}
