package signer

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

type StorageSuite struct{}

var _ = Suite(&StorageSuite{})

func (s *StorageSuite) TestStorage(c *C) {
	store, err := NewSignerStore("", "my secret passphrase")
	c.Assert(err, IsNil)

	item := NewTxOutStoreItem(12, types.TxOutItem{Memo: "foo"})

	c.Assert(store.Set(item), IsNil)
	c.Check(store.Has(item.Key()), Equals, true)

	getItem, err := store.Get(item.Key())
	c.Assert(err, IsNil)
	c.Check(getItem.TxOutItem.Memo, Equals, item.TxOutItem.Memo)

	items := store.List()
	c.Assert(items, HasLen, 1, Commentf("%d", len(items)))

	c.Assert(store.Remove(item), IsNil)
	c.Check(store.Has(item.Key()), Equals, false)

	spent := NewTxOutStoreItem(10, types.TxOutItem{Memo: "spent"})
	spent.Status = TxSpent
	items = []TxOutStoreItem{
		NewTxOutStoreItem(12, types.TxOutItem{Memo: "foo"}),
		NewTxOutStoreItem(12, types.TxOutItem{Memo: "bar"}),
		NewTxOutStoreItem(13, types.TxOutItem{Memo: "baz"}),
		NewTxOutStoreItem(10, types.TxOutItem{Memo: "boo"}),
		spent,
	}

	c.Assert(store.Batch(items), IsNil)

	items = store.List()
	c.Assert(items, HasLen, 4)
	c.Check(items[0].TxOutItem.Memo, Equals, "boo")
	c.Check(items[1].TxOutItem.Memo, Equals, "bar")
	c.Check(items[2].TxOutItem.Memo, Equals, "foo")
	c.Check(items[3].TxOutItem.Memo, Equals, "baz")

	c.Check(store.Close(), IsNil)
}

func (s *StorageSuite) TestKey(c *C) {
	item1 := NewTxOutStoreItem(12, types.TxOutItem{Memo: "foo"})
	item2 := NewTxOutStoreItem(12, types.TxOutItem{Memo: "foo"})
	item3 := NewTxOutStoreItem(1222, types.TxOutItem{Memo: "foo"})
	item4 := NewTxOutStoreItem(12, types.TxOutItem{Memo: "bar"})
	c.Check(item1.Key(), Equals, item2.Key())
	c.Check(item1.Key(), Not(Equals), item3.Key())
	c.Check(item1.Key(), Not(Equals), item4.Key())

	item1.Status = TxSpent
	c.Check(item1.Key(), Equals, item2.Key())
}
