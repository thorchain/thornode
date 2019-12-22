package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type TypeObservedTxSuite struct{}

var _ = Suite(&TypeObservedTxSuite{})

func (s TypeObservedTxSuite) TestVoter(c *C) {
	txID := GetRandomTxHash()

	bnb := GetRandomBNBAddress()
	acc1 := GetRandomBech32Addr()
	acc2 := GetRandomBech32Addr()
	acc3 := GetRandomBech32Addr()
	acc4 := GetRandomBech32Addr()

	accConsPub1 := GetRandomBech32ConsensusPubKey()
	accConsPub2 := GetRandomBech32ConsensusPubKey()
	accConsPub3 := GetRandomBech32ConsensusPubKey()
	accConsPub4 := GetRandomBech32ConsensusPubKey()

	accPubKeySet1 := GetRandomPubKeySet()
	accPubKeySet2 := GetRandomPubKeySet()
	accPubKeySet3 := GetRandomPubKeySet()
	accPubKeySet4 := GetRandomPubKeySet()

	tx1 := GetRandomTx()
	tx1.Memo = "hello"
	tx2 := GetRandomTx()
	observePoolAddr := GetRandomPubKey()
	voter := NewObservedTxVoter(txID, nil)

	obTx1 := NewObservedTx(tx1, sdk.ZeroUint(), observePoolAddr)
	obTx2 := NewObservedTx(tx2, sdk.ZeroUint(), observePoolAddr)

	voter.Add(obTx1, acc1)
	c.Assert(voter.Txs, HasLen, 1)

	voter.Add(obTx1, acc1) // check THORNode don't duplicate the same signer
	c.Assert(voter.Txs, HasLen, 1)
	c.Assert(voter.Txs[0].Signers, HasLen, 1)

	voter.Add(obTx1, acc2) // append a signature
	c.Assert(voter.Txs, HasLen, 1)
	c.Assert(voter.Txs[0].Signers, HasLen, 2)

	voter.Add(obTx2, acc1) // same validator seeing a different version of tx
	c.Assert(voter.Txs, HasLen, 1)
	c.Assert(voter.Txs[0].Signers, HasLen, 2)

	voter.Add(obTx2, acc3) // second version
	c.Assert(voter.Txs, HasLen, 2)
	c.Assert(voter.Txs[0].Signers, HasLen, 2)
	c.Assert(voter.Txs[1].Signers, HasLen, 1)

	trusts3 := NodeAccounts{
		NodeAccount{
			NodeAddress:         acc1,
			Status:              Active,
			PubKeySet:           accPubKeySet1,
			ValidatorConsPubKey: accConsPub1,
		},
		NodeAccount{
			NodeAddress:         acc2,
			Status:              Active,
			PubKeySet:           accPubKeySet2,
			ValidatorConsPubKey: accConsPub2,
		},
		NodeAccount{
			NodeAddress:         acc3,
			Status:              Active,
			PubKeySet:           accPubKeySet3,
			ValidatorConsPubKey: accConsPub3,
		},
	}
	trusts4 := NodeAccounts{
		NodeAccount{
			NodeAddress:         acc1,
			Status:              Active,
			PubKeySet:           accPubKeySet1,
			ValidatorConsPubKey: accConsPub1,
		},
		NodeAccount{
			NodeAddress:         acc2,
			Status:              Active,
			PubKeySet:           accPubKeySet2,
			ValidatorConsPubKey: accConsPub2,
		},
		NodeAccount{
			NodeAddress:         acc3,
			Status:              Active,
			PubKeySet:           accPubKeySet3,
			ValidatorConsPubKey: accConsPub3,
		},
		NodeAccount{
			NodeAddress:         acc4,
			Status:              Active,
			PubKeySet:           accPubKeySet4,
			ValidatorConsPubKey: accConsPub4,
		},
	}

	tx := voter.GetTx(trusts3)
	c.Check(tx.Tx.Memo, Equals, "hello")
	tx = voter.GetTx(trusts4)
	c.Check(tx.Empty(), Equals, true)
	c.Check(voter.HasConensus(trusts3), Equals, true)
	c.Check(voter.HasConensus(trusts4), Equals, false)
	c.Check(voter.Key().Equals(txID), Equals, true)
	c.Check(voter.String() == txID.String(), Equals, true)

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
		tx := common.Tx{
			FromAddress: item.sender,
			ToAddress:   GetRandomBNBAddress(),
			Coins:       item.coins,
			Gas:         common.BNBGasFeeSingleton,
			Memo:        item.memo,
		}
		txIn := NewObservedTx(tx, sdk.ZeroUint(), item.observePoolAddr)
		c.Assert(txIn.Valid(), NotNil)
	}
}

func (TypeObservedTxSuite) TestObservedTxEquals(c *C) {
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
		tx    ObservedTx
		tx1   ObservedTx
		equal bool
	}{
		{
			tx:    NewObservedTx(common.Tx{FromAddress: bnb, ToAddress: GetRandomBNBAddress(), Coins: coins1, Memo: "memo", Gas: common.BNBGasFeeSingleton}, sdk.ZeroUint(), observePoolAddr),
			tx1:   NewObservedTx(common.Tx{FromAddress: bnb, ToAddress: GetRandomBNBAddress(), Coins: coins1, Memo: "memo1", Gas: common.BNBGasFeeSingleton}, sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewObservedTx(common.Tx{FromAddress: bnb, ToAddress: GetRandomBNBAddress(), Coins: coins1, Memo: "memo", Gas: common.BNBGasFeeSingleton}, sdk.ZeroUint(), observePoolAddr),
			tx1:   NewObservedTx(common.Tx{FromAddress: bnb1, ToAddress: GetRandomBNBAddress(), Coins: coins1, Memo: "memo", Gas: common.BNBGasFeeSingleton}, sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewObservedTx(common.Tx{Coins: coins2, Memo: "memo", FromAddress: bnb, ToAddress: GetRandomBNBAddress(), Gas: common.BNBGasFeeSingleton}, sdk.ZeroUint(), observePoolAddr),
			tx1:   NewObservedTx(common.Tx{Coins: coins1, Memo: "memo", FromAddress: bnb, ToAddress: GetRandomBNBAddress(), Gas: common.BNBGasFeeSingleton}, sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewObservedTx(common.Tx{Coins: coins3, Memo: "memo", FromAddress: bnb, ToAddress: GetRandomBNBAddress(), Gas: common.BNBGasFeeSingleton}, sdk.ZeroUint(), observePoolAddr),
			tx1:   NewObservedTx(common.Tx{Coins: coins1, Memo: "memo", FromAddress: bnb, ToAddress: GetRandomBNBAddress(), Gas: common.BNBGasFeeSingleton}, sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewObservedTx(common.Tx{Coins: coins4, Memo: "memo", FromAddress: bnb, ToAddress: GetRandomBNBAddress(), Gas: common.BNBGasFeeSingleton}, sdk.ZeroUint(), observePoolAddr),
			tx1:   NewObservedTx(common.Tx{Coins: coins1, Memo: "memo", FromAddress: bnb, ToAddress: GetRandomBNBAddress(), Gas: common.BNBGasFeeSingleton}, sdk.ZeroUint(), observePoolAddr),
			equal: false,
		},
		{
			tx:    NewObservedTx(common.Tx{Coins: coins1, Memo: "memo", FromAddress: bnb, ToAddress: GetRandomBNBAddress(), Gas: common.BNBGasFeeSingleton}, sdk.ZeroUint(), observePoolAddr),
			tx1:   NewObservedTx(common.Tx{Coins: coins1, Memo: "memo", FromAddress: bnb, ToAddress: GetRandomBNBAddress(), Gas: common.BNBGasFeeSingleton}, sdk.ZeroUint(), observePoolAddr1),
			equal: false,
		},
	}
	for _, item := range inputs {
		c.Assert(item.tx.Equals(item.tx1), Equals, item.equal)
	}
}
