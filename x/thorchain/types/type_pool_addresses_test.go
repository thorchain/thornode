package types

import (
	"encoding/json"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type PoolAddressesSuite struct{}

var _ = Suite(&PoolAddressesSuite{})

func (PoolAddressesSuite) TestNewPoolAddresses(c *C) {
	pk1, err := common.NewPoolPubKey(common.BNBChain, 1024, GetRandomPubKey())
	c.Assert(err, IsNil)
	previous := common.PoolPubKeys{pk1}

	pk2, err := common.NewPoolPubKey(common.BNBChain, 0, GetRandomPubKey())
	c.Assert(err, IsNil)
	pk3, err := common.NewPoolPubKey(common.BTCChain, 0, GetRandomPubKey())
	c.Assert(err, IsNil)
	pk4, err := common.NewPoolPubKey(common.ETHChain, 0, GetRandomPubKey())
	c.Assert(err, IsNil)
	current := common.PoolPubKeys{pk2, pk3, pk4}
	c.Assert(err, IsNil)
	next := common.PoolPubKeys{}
	poolAddr := NewPoolAddresses(previous, current, next)
	result, err := json.MarshalIndent(poolAddr, "", "	")
	c.Assert(err, IsNil)
	c.Log(string(result))
	bnbChainPoolAddr := current.GetByChain(common.BNBChain)
	c.Assert(bnbChainPoolAddr, NotNil)
	c.Assert(bnbChainPoolAddr.Chain.Equals(common.BNBChain), Equals, true)
}
