package signer

import (
	"fmt"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
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

	pk := common.PubKey("thorpub1addwnpepqfup3y8p0egd7ml7vrnlxgl3wvnp89mpn0tjpj0p2nm2gh0n9hlrv3mvmnz")

	spent := NewTxOutStoreItem(10, types.TxOutItem{Chain: common.BNBChain, VaultPubKey: pk, Memo: "spent"})
	spent.Status = TxSpent
	items = []TxOutStoreItem{
		NewTxOutStoreItem(12, types.TxOutItem{Chain: common.BTCChain, VaultPubKey: pk, Memo: "foo"}),
		NewTxOutStoreItem(12, types.TxOutItem{Chain: common.BNBChain, VaultPubKey: pk, Memo: "bar"}),
		NewTxOutStoreItem(13, types.TxOutItem{Chain: common.BNBChain, VaultPubKey: pk, Memo: "baz"}),
		NewTxOutStoreItem(10, types.TxOutItem{Chain: common.BTCChain, VaultPubKey: pk, Memo: "boo"}),
		spent,
	}

	c.Assert(store.Batch(items), IsNil)

	items = store.List()
	c.Assert(items, HasLen, 4)
	c.Check(items[0].TxOutItem.Memo, Equals, "boo")
	c.Check(items[1].TxOutItem.Memo, Equals, "foo", Commentf("%s", items[1].TxOutItem.Memo))
	c.Check(items[2].TxOutItem.Memo, Equals, "bar", Commentf("%s", items[2].TxOutItem.Memo))
	c.Check(items[3].TxOutItem.Memo, Equals, "baz")

	ordered := store.OrderedLists()
	c.Assert(ordered, HasLen, 2, Commentf("%+v", ordered))
	c.Check(ordered[fmt.Sprintf("BTC-%s", pk.String())][0].TxOutItem.Memo, Equals, "boo")
	c.Check(ordered[fmt.Sprintf("BTC-%s", pk.String())][1].TxOutItem.Memo, Equals, "foo", Commentf("%s", items[1].TxOutItem.Memo))
	c.Check(ordered[fmt.Sprintf("BNB-%s", pk.String())][0].TxOutItem.Memo, Equals, "bar", Commentf("%s", items[2].TxOutItem.Memo))
	c.Check(ordered[fmt.Sprintf("BNB-%s", pk.String())][1].TxOutItem.Memo, Equals, "baz")

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
