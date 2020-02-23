package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type VaultSuite struct{}

var _ = Suite(&VaultSuite{})

func (s *VaultSuite) TestVault(c *C) {
	pk := GetRandomPubKey()

	vault := Vault{}
	c.Check(vault.IsEmpty(), Equals, true)
	c.Check(vault.IsValid(), NotNil)

	vault = NewVault(12, ActiveVault, YggdrasilVault, pk)
	c.Check(vault.PubKey.Equals(pk), Equals, true)
	c.Check(vault.HasFunds(), Equals, false)
	c.Check(vault.IsEmpty(), Equals, false)
	c.Check(vault.IsValid(), IsNil)

	coins := common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(500*common.One)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(400*common.One)),
	}

	vault.AddFunds(coins)
	c.Check(vault.HasFunds(), Equals, true)
	c.Check(vault.GetCoin(common.BNBAsset).Amount.Equal(sdk.NewUint(500*common.One)), Equals, true)
	c.Check(vault.GetCoin(common.BTCAsset).Amount.Equal(sdk.NewUint(400*common.One)), Equals, true)
	vault.AddFunds(coins)
	c.Check(vault.GetCoin(common.BNBAsset).Amount.Equal(sdk.NewUint(1000*common.One)), Equals, true)
	c.Check(vault.GetCoin(common.BTCAsset).Amount.Equal(sdk.NewUint(800*common.One)), Equals, true)
	vault.SubFunds(coins)
	c.Check(vault.GetCoin(common.BNBAsset).Amount.Equal(sdk.NewUint(500*common.One)), Equals, true)
	c.Check(vault.GetCoin(common.BTCAsset).Amount.Equal(sdk.NewUint(400*common.One)), Equals, true)
	vault.SubFunds(coins)
	c.Check(vault.GetCoin(common.BNBAsset).Amount.Equal(sdk.ZeroUint()), Equals, true)
	c.Check(vault.GetCoin(common.BTCAsset).Amount.Equal(sdk.ZeroUint()), Equals, true)
	c.Check(vault.HasFunds(), Equals, false)
	vault.SubFunds(coins)
	c.Check(vault.GetCoin(common.BNBAsset).Amount.Equal(sdk.ZeroUint()), Equals, true)
	c.Check(vault.GetCoin(common.BTCAsset).Amount.Equal(sdk.ZeroUint()), Equals, true)
	c.Check(vault.HasFunds(), Equals, false)
}

func (s *VaultSuite) TestGetTssSigners(c *C) {
	vault := NewVault(12, ActiveVault, AsgardVault, GetRandomPubKey())
	nodeAccounts := NodeAccounts{}
	memberShip := common.PubKeys{}
	for i := 0; i < 10; i++ {
		na := GetRandomNodeAccount(Active)
		nodeAccounts = append(nodeAccounts, na)
		memberShip = append(memberShip, na.PubKeySet.Secp256k1)
	}
	vault.Membership = memberShip
	addrs := []sdk.AccAddress{
		nodeAccounts[0].NodeAddress,
		nodeAccounts[1].NodeAddress,
	}
	keys, err := vault.GetMembers(addrs)
	c.Assert(err, IsNil)
	c.Assert(keys, HasLen, 2)
	c.Assert(keys[0].Equals(nodeAccounts[0].PubKeySet.Secp256k1), Equals, true)
	c.Assert(keys[1].Equals(nodeAccounts[1].PubKeySet.Secp256k1), Equals, true)
}
