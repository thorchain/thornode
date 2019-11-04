package types

import (
	"encoding/json"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type PoolAddressesSuite struct{}

var _ = Suite(&PoolAddressesSuite{})

func (PoolAddressesSuite) TestNewPoolAddresses(c *C) {
	previous := common.PoolPubKeys{
		common.NewPoolPubKey(common.BNBChain, 1024, GetRandomPubKey()),
	}
	current := common.PoolPubKeys{
		common.NewPoolPubKey(common.BNBChain, 0, GetRandomPubKey()),
		common.NewPoolPubKey(common.BTCChain, 0, GetRandomPubKey()),
		common.NewPoolPubKey(common.ETHChain, 0, GetRandomPubKey()),
	}
	next := common.PoolPubKeys{}
	poolAddr := NewPoolAddresses(previous, current, next, 28800, 27800)
	result, err := json.MarshalIndent(poolAddr, "", "	")
	c.Assert(err, IsNil)
	c.Log(string(result))
	bnbChainPoolAddr := current.GetByChain(common.BNBChain)
	c.Assert(bnbChainPoolAddr, NotNil)
	c.Assert(bnbChainPoolAddr.Chain.Equals(common.BNBChain), Equals, true)
}
