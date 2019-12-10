package bnb

import (
	"github.com/binance-chain/go-sdk/common/types"

	"testing"

	"gitlab.com/thorchain/thornode/bifrostv2/config"

	. "gopkg.in/check.v1"
)

type BNBSuite struct {
	Client *Client
}

var cfg = config.ChainConfigurations{
	Name:         "BNB",
	ChainHost:    "http://45.76.119.188:27147", // TODO change to mock server...
	ChainNetwork: "mainnet",
}

var _ = Suite(&BNBSuite{})

func Test(t *testing.T) {
	TestingT(t)
}

func (s *BNBSuite) SetUpSuite(c *C) {
	var err error
	s.Client, err = NewClient(cfg)
	c.Assert(err, IsNil)
}

func (s *BNBSuite) TestSetNetwork(c *C) {
	network := setNetwork(config.ChainConfigurations{})
	c.Assert(network, Equals, types.TestNetwork)

	network = setNetwork(config.ChainConfigurations{
		ChainNetwork: "mainnet",
	})
	c.Assert(network, Equals, types.ProdNetwork)

	network = setNetwork(config.ChainConfigurations{
		ChainNetwork: "testnet",
	})
	c.Assert(network, Equals, types.TestNetwork)
}

func (s *BNBSuite) TestGetBlock(c *C) {
	ResultBlock, err := s.Client.getBlock(1)
	c.Assert(err, IsNil)
	c.Assert(ResultBlock, NotNil)

}
