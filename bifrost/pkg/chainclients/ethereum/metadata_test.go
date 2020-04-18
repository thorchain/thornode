package ethereum

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

type MetadataSuite struct{}

var _ = Suite(&MetadataSuite{})

func (s *MetadataSuite) TestMetaData(c *C) {
	store := NewEthereumMetaDataStore()

	pk := types.GetRandomPubKey()

	store.Set(pk, EthereumMetadata{"lol", 2, 3})

	meta := store.Get(types.GetRandomPubKey())
	c.Check(meta.Address, Equals, "")
	c.Check(meta.Nonce, Equals, uint64(0))

	meta = store.Get(pk)
	c.Check(meta.Address, Equals, "lol")
	c.Check(meta.Nonce, Equals, uint64(2))

	meta = store.GetByAccount("lol")
	c.Check(meta.Address, Equals, "lol")
	c.Check(meta.Nonce, Equals, uint64(2))

	meta = store.GetByAccount("why")
	c.Check(meta.Address, Equals, "")
	c.Check(meta.Nonce, Equals, uint64(0))

	store.NonceInc(pk)
	meta = store.Get(pk)
	c.Check(meta.Address, Equals, "lol")
	c.Check(meta.Nonce, Equals, uint64(3))
}
