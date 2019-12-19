package thorchain

import (
	"github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperVaultSuite struct{}

var _ = Suite(&KeeperVaultSuite{})

func (s *KeeperVaultSuite) TestVault(c *C) {
	ctx, k := setupKeeperForTest(c)
	pubKey := GetRandomPubKey()
	yggdrasil := NewVault(YggdrasilVault, pubKey)
	err := k.SetVault(ctx, yggdrasil)
	c.Assert(err, IsNil)
	c.Assert(k.VaultExists(ctx, pubKey), Equals, true)
	pubKey1 := GetRandomPubKey()
	yggdrasil1 := NewVault(YggdrasilVault, pubKey1)
	yggdrasil1.Coins = common.Coins{
		common.NewCoin(common.BNBAsset, types.NewUint(100)),
	}
	c.Assert(k.SetVault(ctx, yggdrasil1), IsNil)
	ygg, err := k.GetVault(ctx, pubKey1)
	c.Assert(err, IsNil)
	c.Assert(ygg.IsEmpty(), Equals, false)
	hasYgg, err := k.HasValidVaultPools(ctx)
	c.Assert(err, IsNil)
	c.Assert(hasYgg, Equals, true)
}
