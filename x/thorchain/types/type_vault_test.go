package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type VaultSuite struct{}

var _ = Suite(&VaultSuite{})

func (s *VaultSuite) TestVault(c *C) {
	pk := GetRandomPubKey()

	vault := Vault{}
	c.Check(vault.IsEmpty(), Equals, true)
	c.Check(vault.IsValid(), NotNil)

	vault = NewVault(YggdrasilVault, pk)
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
