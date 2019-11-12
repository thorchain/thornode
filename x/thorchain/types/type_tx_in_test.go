package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type TypeTxInSuite struct{}

var _ = Suite(&TypeTxInSuite{})

func (s TypeTxInSuite) TestVoter(c *C) {
	txID := GetRandomTxHash()
	txID2 := GetRandomTxHash()

	bnb := GetRandomBNBAddress()
	acc1 := GetRandomBech32Addr()
	acc2 := GetRandomBech32Addr()
	acc3 := GetRandomBech32Addr()
	acc4 := GetRandomBech32Addr()

	accConsPub1 := GetRandomBech32ConsensusPubKey()
	accConsPub2 := GetRandomBech32ConsensusPubKey()
	accConsPub3 := GetRandomBech32ConsensusPubKey()
	accConsPub4 := GetRandomBech32ConsensusPubKey()

	accPubKeys1 := GetRandomPubKeys()
	accPubKeys2 := GetRandomPubKeys()
	accPubKeys3 := GetRandomPubKeys()
	accPubKeys4 := GetRandomPubKeys()

	observePoolAddr := GetRandomPubKey()
	voter := NewTxInVoter(txID, nil)

	txIn := NewTxIn(nil, "hello", bnb, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr)
	txIn2 := NewTxIn(nil, "goodbye", bnb, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr)

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
			NodeAddress:         acc1,
			Status:              Active,
			NodePubKey:          accPubKeys1,
			ValidatorConsPubKey: accConsPub1,
		},
		NodeAccount{
			NodeAddress:         acc2,
			Status:              Active,
			NodePubKey:          accPubKeys2,
			ValidatorConsPubKey: accConsPub2,
		},
		NodeAccount{
			NodeAddress:         acc3,
			Status:              Active,
			NodePubKey:          accPubKeys3,
			ValidatorConsPubKey: accConsPub3,
		},
	}
	trusts4 := NodeAccounts{
		NodeAccount{
			NodeAddress:         acc1,
			Status:              Active,
			NodePubKey:          accPubKeys1,
			ValidatorConsPubKey: accConsPub1,
		},
		NodeAccount{
			NodeAddress:         acc2,
			Status:              Active,
			NodePubKey:          accPubKeys2,
			ValidatorConsPubKey: accConsPub2,
		},
		NodeAccount{
			NodeAddress:         acc3,
			Status:              Active,
			NodePubKey:          accPubKeys3,
			ValidatorConsPubKey: accConsPub3,
		},
		NodeAccount{
			NodeAddress:         acc4,
			Status:              Active,
			NodePubKey:          accPubKeys4,
			ValidatorConsPubKey: accConsPub4,
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
		c.Check(transaction.OutHashes[0].Equals(txID2), Equals, true)
	}

	txIn.SetReverted(txID2)
	c.Check(txIn.OutHashes[0].Equals(txID2), Equals, true)
	c.Check(len(txIn.String()) > 0, Equals, true)
	statechainCoins := common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(100)),
		common.NewCoin(common.BNBAsset, sdk.NewUint(100)),
	}
	inputs := []struct {
		coins           common.Coins
		memo            string
		sender          common.Address
		observePoolAddr common.PubKey
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
			observePoolAddr: common.EmptyPubKey,
		},
	}

	for _, item := range inputs {
		txIn := NewTxIn(item.coins, item.memo, item.sender, GetRandomBNBAddress(), sdk.ZeroUint(), item.observePoolAddr)
		c.Assert(txIn.Valid(), NotNil)
	}
}

func (TypeTxInSuite) TestTxInEquals(c *C) {
	coins1 := common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
		common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
	}
	coins2 := common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
	}
	coins3 := common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(200*common.One)),
		common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
	}
	coins4 := common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
		common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
	}
	bnb, err := common.NewAddress("bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38")
	c.Assert(err, IsNil)
	bnb1, err := common.NewAddress("bnb1yk882gllgv3rt2rqrsudf6kn2agr94etnxu9a7")
	c.Assert(err, IsNil)
	observePoolAddr := GetRandomPubKey()
	observePoolAddr1 := GetRandomPubKey()
	inputs := []struct {
		tx    TxIn
		tx1   TxIn
		equal bool
	}{
		{
			tx:    NewTxIn(coins1, "memo", bnb, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr),
			tx1:   NewTxIn(coins1, "memo1", bnb, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewTxIn(coins1, "memo", bnb, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr),
			tx1:   NewTxIn(coins1, "memo", bnb1, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewTxIn(coins2, "memo", bnb, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr),
			tx1:   NewTxIn(coins1, "memo", bnb, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewTxIn(coins3, "memo", bnb, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr),
			tx1:   NewTxIn(coins1, "memo", bnb, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewTxIn(coins4, "memo", bnb, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr),
			tx1:   NewTxIn(coins1, "memo", bnb, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewTxIn(coins1, "memo", bnb, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr),
			tx1:   NewTxIn(coins1, "memo", bnb, GetRandomBNBAddress(), sdk.ZeroUint(), observePoolAddr1),
			equal: false,
		},
	}
	for _, item := range inputs {
		c.Assert(item.tx.Equals(item.tx1), Equals, item.equal)
	}
}
