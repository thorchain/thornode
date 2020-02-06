package signer

import (
	"os"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

type StorageSuite struct{}

var _ = Suite(&StorageSuite{})

func (s *StorageSuite) TestStorage(c *C) {
	dir := os.TempDir()
	defer os.Remove(dir)

	store, err := NewSignerStore(dir)
	c.Assert(err, IsNil)

	item := NewTxOutStoreItem(12, types.TxOutItem{Memo: "foo"})

	c.Assert(store.Set(item), IsNil)
	c.Check(store.Has(item.Key()), Equals, true)

	getItem, err := store.Get(item.Key())
	c.Assert(err, IsNil)
	c.Check(getItem.TxOutItem.Memo, Equals, item.TxOutItem.Memo)

	items, err := store.List()
	c.Assert(err, IsNil)
	c.Assert(items, HasLen, 1, Commentf("%d", len(items)))

	c.Assert(store.Remove(item), IsNil)
	c.Check(store.Has(item.Key()), Equals, false)

	items = []TxOutStoreItem{
		NewTxOutStoreItem(12, types.TxOutItem{Memo: "foo"}),
		NewTxOutStoreItem(12, types.TxOutItem{Memo: "bar"}),
		NewTxOutStoreItem(13, types.TxOutItem{Memo: "baz"}),
		NewTxOutStoreItem(10, types.TxOutItem{Memo: "boo"}),
	}

	c.Assert(store.Batch(items), IsNil)

	items, err = store.List()
	c.Assert(err, IsNil)
	c.Assert(items, HasLen, 4)
	c.Check(items[0].TxOutItem.Memo, Equals, "boo")
	c.Check(items[1].TxOutItem.Memo, Equals, "foo")
	c.Check(items[2].TxOutItem.Memo, Equals, "bar")
	c.Check(items[3].TxOutItem.Memo, Equals, "baz")

	c.Check(store.Close(), IsNil)
}
