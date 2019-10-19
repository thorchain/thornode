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
	bnb, err := common.NewAddress("bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38")
	c.Assert(err, IsNil)
	acc1, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	acc2, err := sdk.AccAddressFromBech32("bep1p4tyz2d243j539n6hrx3q8p77uyqks6ks4ju9u")
	c.Assert(err, IsNil)
	acc3, err := sdk.AccAddressFromBech32("bep1lkasqgxc3k65fqqw9zeurxwxmayfjej3ygwpzl")
	c.Assert(err, IsNil)
	acc4, err := sdk.AccAddressFromBech32("bep1gnaghgzcpd73hcxeturml96maa0fajg9zrtrez")
	c.Assert(err, IsNil)
	accConsPub1 := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdrg`
	accConsPub2 := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdrx`
	accConsPub3 := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdry`
	accConsPub4 := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdrz`

	observePoolAddr, err := common.NewAddress("tbnb1h792qpzue9dmdu8pfpdpyjqfd0s902dljvs4jz")
	c.Assert(err, IsNil)
	voter := NewTxInVoter(txID, nil)

	txIn := NewTxIn(nil, "hello", bnb, sdk.ZeroUint(), observePoolAddr)
	txIn2 := NewTxIn(nil, "goodbye", bnb, sdk.ZeroUint(), observePoolAddr)

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

	trusts3 := NodeAccounts{
		NodeAccount{
			NodeAddress: acc1,
			Status:      Active,
			Accounts: TrustAccount{
				SignerBNBAddress:       bnb,
				ObserverBEPAddress:     acc1,
				ValidatorBEPConsPubKey: accConsPub1,
			},
		},
		NodeAccount{
			NodeAddress: acc2,
			Status:      Active,
			Accounts: TrustAccount{
				SignerBNBAddress:       bnb,
				ObserverBEPAddress:     acc2,
				ValidatorBEPConsPubKey: accConsPub2,
			},
		},
		NodeAccount{
			NodeAddress: acc3,
			Status:      Active,
			Accounts: TrustAccount{
				SignerBNBAddress:       bnb,
				ObserverBEPAddress:     acc3,
				ValidatorBEPConsPubKey: accConsPub3,
			},
		},
	}
	trusts4 := NodeAccounts{
		NodeAccount{
			NodeAddress: acc1,
			Status:      Active,
			Accounts: TrustAccount{
				SignerBNBAddress:       bnb,
				ObserverBEPAddress:     acc1,
				ValidatorBEPConsPubKey: accConsPub1,
			},
		},
		NodeAccount{
			NodeAddress: acc2,
			Status:      Active,
			Accounts: TrustAccount{
				SignerBNBAddress:       bnb,
				ObserverBEPAddress:     acc2,
				ValidatorBEPConsPubKey: accConsPub2,
			},
		},
		NodeAccount{
			NodeAddress: acc3,
			Status:      Active,
			Accounts: TrustAccount{
				SignerBNBAddress:       bnb,
				ObserverBEPAddress:     acc3,
				ValidatorBEPConsPubKey: accConsPub3,
			},
		},
		NodeAccount{
			NodeAddress: acc4,
			Status:      Active,
			Accounts: TrustAccount{
				SignerBNBAddress:       bnb,
				ObserverBEPAddress:     acc4,
				ValidatorBEPConsPubKey: accConsPub4,
			},
		},
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
		coins           common.Coins
		memo            string
		sender          common.Address
		observePoolAddr common.Address
	}{
		{
			coins:           nil,
			memo:            "test",
			sender:          bnb,
			observePoolAddr: observePoolAddr,
		},
		{
			coins:           common.Coins{},
			memo:            "test",
			sender:          bnb,
			observePoolAddr: observePoolAddr,
		},
		{
			coins:           statechainCoins,
			memo:            "",
			sender:          bnb,
			observePoolAddr: observePoolAddr,
		},
		{
			coins:           statechainCoins,
			memo:            "test",
			sender:          common.NoAddress,
			observePoolAddr: observePoolAddr,
		},
		{
			coins:           statechainCoins,
			memo:            "test",
			sender:          bnb,
			observePoolAddr: common.NoAddress,
		},
	}

	for _, item := range inputs {
		txIn := NewTxIn(item.coins, item.memo, item.sender, sdk.ZeroUint(), item.observePoolAddr)
		c.Assert(txIn.Valid(), NotNil)
	}
}

func (TypeTxInSuite) TestTxInEquals(c *C) {
	coins1 := common.Coins{
		common.NewCoin(common.BNBChain, common.BNBTicker, sdk.NewUint(100*common.One)),
		common.NewCoin(common.BNBChain, common.RuneA1FTicker, sdk.NewUint(100*common.One)),
	}
	coins2 := common.Coins{
		common.NewCoin(common.BNBChain, common.BNBTicker, sdk.NewUint(100*common.One)),
	}
	coins3 := common.Coins{
		common.NewCoin(common.BNBChain, common.BNBTicker, sdk.NewUint(200*common.One)),
		common.NewCoin(common.BNBChain, common.RuneA1FTicker, sdk.NewUint(100*common.One)),
	}
	coins4 := common.Coins{
		common.NewCoin(common.BNBChain, common.RuneB1ATicker, sdk.NewUint(100*common.One)),
		common.NewCoin(common.BNBChain, common.RuneA1FTicker, sdk.NewUint(100*common.One)),
	}
	bnb, err := common.NewAddress("bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38")
	c.Assert(err, IsNil)
	bnb1, err := common.NewAddress("bnb1yk882gllgv3rt2rqrsudf6kn2agr94etnxu9a7")
	c.Assert(err, IsNil)
	observePoolAddr, err := common.NewAddress("bnb1g0xakzh03tpa54khxyvheeu92hwzypkdce77rm")
	c.Assert(err, IsNil)
	observePoolAddr1, err := common.NewAddress("bnb1zxseqkfm3en5cw6dh9xgmr85hw6jtwamnd2y2v")
	c.Assert(err, IsNil)
	inputs := []struct {
		tx    TxIn
		tx1   TxIn
		equal bool
	}{
		{
			tx:    NewTxIn(coins1, "memo", bnb, sdk.ZeroUint(), observePoolAddr),
			tx1:   NewTxIn(coins1, "memo1", bnb, sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewTxIn(coins1, "memo", bnb, sdk.ZeroUint(), observePoolAddr),
			tx1:   NewTxIn(coins1, "memo", bnb1, sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewTxIn(coins2, "memo", bnb, sdk.ZeroUint(), observePoolAddr),
			tx1:   NewTxIn(coins1, "memo", bnb, sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewTxIn(coins3, "memo", bnb, sdk.ZeroUint(), observePoolAddr),
			tx1:   NewTxIn(coins1, "memo", bnb, sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewTxIn(coins4, "memo", bnb, sdk.ZeroUint(), observePoolAddr),
			tx1:   NewTxIn(coins1, "memo", bnb, sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewTxIn(coins1, "memo", bnb, sdk.ZeroUint(), observePoolAddr),
			tx1:   NewTxIn(coins1, "memo", bnb, sdk.ZeroUint(), observePoolAddr1),
			equal: false,
		},
	}
	for _, item := range inputs {
		c.Assert(item.tx.Equals(item.tx1), Equals, item.equal)
	}
}
