package binance

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

type MetadataSuite struct{}

var _ = Suite(&MetadataSuite{})

func (s *MetadataSuite) TestMetaData(c *C) {
	store := NewBinanceMetaDataStore()

	pk := types.GetRandomPubKey()

	store.Set(pk, BinanceMetadata{1, 2, 3})

	meta := store.Get(types.GetRandomPubKey())
	c.Check(meta.AccountNumber, Equals, int64(0))
	c.Check(meta.SeqNumber, Equals, int64(0))

	meta = store.Get(pk)
	c.Check(meta.AccountNumber, Equals, int64(1))
	c.Check(meta.SeqNumber, Equals, int64(2))

	meta = store.GetByAccount(1)
	c.Check(meta.AccountNumber, Equals, int64(1))
	c.Check(meta.SeqNumber, Equals, int64(2))

	meta = store.GetByAccount(10000)
	c.Check(meta.AccountNumber, Equals, int64(0))
	c.Check(meta.SeqNumber, Equals, int64(0))

	store.SeqInc(pk)
	meta = store.Get(pk)
	c.Check(meta.AccountNumber, Equals, int64(1))
	c.Check(meta.SeqNumber, Equals, int64(3))
}
