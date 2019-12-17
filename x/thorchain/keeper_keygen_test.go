package thorchain

import (
	. "gopkg.in/check.v1"
)

type KeeperKeygensSuite struct{}

var _ = Suite(&KeeperKeygensSuite{})

func (s *KeeperKeygensSuite) TestKeeperKeygens(c *C) {
	var err error
	ctx, k := setupKeeperForTest(c)

	keygens := NewKeygens(1)
	keygen := Keygen{GetRandomPubKey(), GetRandomPubKey(), GetRandomPubKey()}
	keygens.Keygens = append(keygens.Keygens, keygen)
	c.Assert(k.SetKeygens(ctx, keygens), IsNil)

	keygens, err = k.GetKeygens(ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(keygens, NotNil)
	c.Assert(keygens.Height, Equals, uint64(1))

	keygens, err = k.GetKeygens(ctx, 100)
	c.Assert(err, IsNil)
	c.Assert(keygens, NotNil)

	iter := k.GetKeygensIterator(ctx)
	defer iter.Close()
}
