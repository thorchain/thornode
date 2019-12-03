package thorchain

import (
	"github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperYggdrasilSuite struct{}

var _ = Suite(&KeeperYggdrasilSuite{})

func (KeeperYggdrasilSuite) TestYggdrasil(c *C) {
	ctx, k := setupKeeperForTest(c)
	pubKey := GetRandomPubKey()
	yggdrasil := NewYggdrasil(pubKey)
	err := k.SetYggdrasil(ctx, yggdrasil)
	c.Assert(err, IsNil)
	c.Assert(k.YggdrasilExists(ctx, pubKey), Equals, true)
	pubKey1 := GetRandomPubKey()
	yggdrasil1 := NewYggdrasil(pubKey1)
	yggdrasil1.Coins = common.Coins{
		common.NewCoin(common.BNBAsset, types.NewUint(100)),
	}
	c.Assert(k.SetYggdrasil(ctx, yggdrasil1), IsNil)
	ygg, err := k.GetYggdrasil(ctx, pubKey1)
	c.Assert(err, IsNil)
	c.Assert(ygg.IsEmpty(), Equals, false)
	hasYgg, err := k.HasValidYggdrasilPools(ctx)
	c.Assert(err, IsNil)
	c.Assert(hasYgg, Equals, true)

	addr, err := pubKey1.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	pk, err := k.FindPubKeyOfAddress(ctx, addr, common.BNBChain)
	c.Assert(err, IsNil)
	c.Assert(pk.IsEmpty(), Equals, false)

}
