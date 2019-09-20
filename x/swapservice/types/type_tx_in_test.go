package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type TypeTxInSuite struct{}

var _ = Suite(&TypeTxInSuite{})

func (s TypeTxInSuite) TestVoter(c *C) {
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	txID2, err := common.NewTxID("47B4FE474A63DDF79DF2790C1C5162F4C213484750AB8292CFE7342E4B0B40E2")
	c.Assert(err, IsNil)
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	acc1, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	acc2, err := sdk.AccAddressFromBech32("bep1p4tyz2d243j539n6hrx3q8p77uyqks6ks4ju9u")
	c.Assert(err, IsNil)
	acc3, err := sdk.AccAddressFromBech32("bep1lkasqgxc3k65fqqw9zeurxwxmayfjej3ygwpzl")
	c.Assert(err, IsNil)
	acc4, err := sdk.AccAddressFromBech32("bep1gnaghgzcpd73hcxeturml96maa0fajg9zrtrez")
	c.Assert(err, IsNil)

	voter := NewTxInVoter(txID, nil)

	txIn := NewTxIn(nil, "hello", bnb, sdk.ZeroUint())
	txIn2 := NewTxIn(nil, "goodbye", bnb, sdk.ZeroUint())

	voter.Adds([]TxIn{txIn}, acc1)
	c.Assert(voter.Txs, HasLen, 1)

	voter.Adds([]TxIn{txIn}, acc1) // check we don't duplicate the same signer
	c.Assert(voter.Txs, HasLen, 1)
	c.Assert(voter.Txs[0].Signers, HasLen, 1)

	voter.Add(txIn, acc2) // append a signature
	c.Assert(voter.Txs, HasLen, 1)
	c.Assert(voter.Txs[0].Signers, HasLen, 2)

	voter.Add(txIn2, acc1) // same validator seeing a different version of tx
	c.Assert(voter.Txs, HasLen, 1)
	c.Assert(voter.Txs[0].Signers, HasLen, 2)

	voter.Add(txIn2, acc3) // second version
	c.Assert(voter.Txs, HasLen, 2)
	c.Assert(voter.Txs[0].Signers, HasLen, 2)
	c.Assert(voter.Txs[1].Signers, HasLen, 1)

	trusts3 := TrustAccounts{
		TrustAccount{ObserverAddress: acc1},
		TrustAccount{ObserverAddress: acc2},
		TrustAccount{ObserverAddress: acc3},
	}
	trusts4 := TrustAccounts{
		TrustAccount{ObserverAddress: acc1},
		TrustAccount{ObserverAddress: acc2},
		TrustAccount{ObserverAddress: acc3},
		TrustAccount{ObserverAddress: acc4},
	}

	tx := voter.GetTx(trusts3)
	c.Check(tx.Memo, Equals, "hello")
	tx = voter.GetTx(trusts4)
	c.Check(tx.Empty(), Equals, true)
	c.Check(voter.HasConensus(trusts3), Equals, true)
	c.Check(voter.HasConensus(trusts4), Equals, false)
	c.Check(voter.Key().Equals(txID), Equals, true)
	c.Check(voter.String() == txID.String(), Equals, true)
	voter.SetDone(txID2)
	for _, transaction := range voter.Txs {
		c.Check(transaction.Done.Equals(txID2), Equals, true)
	}

	txIn.SetReverted(txID2)
	c.Check(txIn.Done.Equals(txID2), Equals, true)
	c.Check(len(txIn.String()) > 0, Equals, true)
	coins := sdk.NewCoins(
		sdk.NewCoin("rune", sdk.NewInt(100)), sdk.NewCoin("bnb", sdk.NewInt(100)))
	statechainCoins, err := FromSdkCoins(coins)
	c.Assert(err, IsNil)
	c.Assert(statechainCoins, NotNil)
	inputs := []struct {
		coins  common.Coins
		memo   string
		sender common.BnbAddress
	}{
		{
			coins:  nil,
			memo:   "test",
			sender: bnb,
		},
		{
			coins:  common.Coins{},
			memo:   "test",
			sender: bnb,
		},
		{
			coins:  statechainCoins,
			memo:   "",
			sender: bnb,
		},
		{
			coins:  statechainCoins,
			memo:   "test",
			sender: common.NoBnbAddress,
		},
	}

	for _, item := range inputs {
		txIn := NewTxIn(item.coins, item.memo, item.sender, sdk.ZeroUint())
		c.Assert(txIn.Valid(), NotNil)
	}
}

func (TypeTxInSuite) TestTxInEquals(c *C) {
	coins1 := common.Coins{
		common.NewCoin(common.BNBTicker, sdk.NewUint(100*common.One)),
		common.NewCoin(common.RuneA1FTicker, sdk.NewUint(100*common.One)),
	}
	coins2 := common.Coins{
		common.NewCoin(common.BNBTicker, sdk.NewUint(100*common.One)),
	}
	coins3 := common.Coins{
		common.NewCoin(common.BNBTicker, sdk.NewUint(200*common.One)),
		common.NewCoin(common.RuneA1FTicker, sdk.NewUint(100*common.One)),
	}
	coins4 := common.Coins{
		common.NewCoin(common.RuneB1ATicker, sdk.NewUint(100*common.One)),
		common.NewCoin(common.RuneA1FTicker, sdk.NewUint(100*common.One)),
	}
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	bnb1, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqb")
	c.Assert(err, IsNil)
	inputs := []struct {
		tx    TxIn
		tx1   TxIn
		equal bool
	}{
		{
			tx:    NewTxIn(coins1, "memo", bnb, sdk.ZeroUint()),
			tx1:   NewTxIn(coins1, "memo1", bnb, sdk.ZeroUint()),
			equal: false,
		},
		{
			tx:    NewTxIn(coins1, "memo", bnb, sdk.ZeroUint()),
			tx1:   NewTxIn(coins1, "memo", bnb1, sdk.ZeroUint()),
			equal: false,
		},
		{
			tx:    NewTxIn(coins2, "memo", bnb, sdk.ZeroUint()),
			tx1:   NewTxIn(coins1, "memo", bnb, sdk.ZeroUint()),
			equal: false,
		},
		{
			tx:    NewTxIn(coins3, "memo", bnb, sdk.ZeroUint()),
			tx1:   NewTxIn(coins1, "memo", bnb, sdk.ZeroUint()),
			equal: false,
		},
		{
			tx:    NewTxIn(coins4, "memo", bnb, sdk.ZeroUint()),
			tx1:   NewTxIn(coins1, "memo", bnb, sdk.ZeroUint()),
			equal: false,
		},
	}
	for _, item := range inputs {
		c.Assert(item.tx.Equals(item.tx1), Equals, item.equal)
	}
}
