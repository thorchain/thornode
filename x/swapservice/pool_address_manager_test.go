package swapservice

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/statechain/cmd"
)

type PoolAddressManagerSuite struct{}

var _ = Suite(&PoolAddressManagerSuite{})

func (ps *PoolAddressManagerSuite) SetUpSuite(c *C) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	config.SetBech32PrefixForConsensusNode(cmd.Bech32PrefixConsAddr, cmd.Bech32PrefixConsPub)
}

func (PoolAddressManagerSuite) TestSetupInitialPoolAddresses(c *C) {
	ctx, k := setupKeeperForTest(c)
	poolAddrMgr := NewPoolAddressManager(k)
	c.Assert(poolAddrMgr, NotNil)
	// incorrect block height
	pa, err := poolAddrMgr.setupInitialPoolAddresses(ctx, 0)
	c.Assert(err, NotNil)
	c.Assert(pa.IsEmpty(), Equals, true)

	pa1, err := poolAddrMgr.setupInitialPoolAddresses(ctx, 1)
	c.Assert(err, NotNil)
	c.Assert(pa1.IsEmpty(), Equals, true)

	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	addr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	c.Check(addr.Empty(), Equals, false)
	bepConsPubKey := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdrg`
	trustAccount := NewTrustAccount(bnb, addr, bepConsPubKey)
	err = trustAccount.IsValid()
	c.Assert(err, IsNil)
	nodeAddress, err := sdk.AccAddressFromBech32("bep1rtgz3lcaw8vw0yfsc8ga0rdgwa3qh9ju7vfsnk")
	c.Assert(err, IsNil)
	na := NewNodeAccount(nodeAddress, NodeActive, trustAccount)
	k.SetNodeAccount(ctx, na)

	pa2, err := poolAddrMgr.setupInitialPoolAddresses(ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(pa2.IsEmpty(), Equals, false)
	c.Assert(pa2.Current.String(), Equals, bnb.String())
	c.Assert(pa2.Next.String(), Equals, bnb.String())

	// Two nodes
	na1 := GetRandomNodeAccount(NodeActive)
	k.SetNodeAccount(ctx, na1)

	// with two active nodes
	pa3, err := poolAddrMgr.setupInitialPoolAddresses(ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(pa3.IsEmpty(), Equals, false)
	c.Assert(pa3.Current.String(), Equals, bnb.String())
	c.Assert(pa3.Next.String(), Equals, na1.Accounts.SignerBNBAddress.String())

	nodeAccounts := NodeAccounts{na1}
	// with more than two  active nodes
	for i := 0; i < 10; i++ {
		na2 := GetRandomNodeAccount(NodeActive)
		k.SetNodeAccount(ctx, na2)
		nodeAccounts = append(nodeAccounts, na2)
	}

	sort.Sort(nodeAccounts)

	pa4, err := poolAddrMgr.setupInitialPoolAddresses(ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(pa4.IsEmpty(), Equals, false)
	c.Assert(pa4.Current.String(), Equals, bnb.String())
	c.Assert(pa4.Next.String(), Equals, nodeAccounts[0].Accounts.SignerBNBAddress.String())
	c.Logf("%+v", pa4)

	newPa := poolAddrMgr.rotatePoolAddress(ctx, 101, pa4)
	c.Assert(newPa.IsEmpty(), Equals, false)
	c.Assert(newPa.Previous.String(), Equals, pa4.Current.String())
	c.Assert(newPa.Current.String(), Equals, pa4.Next.String())
	c.Assert(newPa.Next.String(), Equals, nodeAccounts[1].Accounts.SignerBNBAddress.String())
	c.Assert(newPa.RotateAt, Equals, int64(201))
}
