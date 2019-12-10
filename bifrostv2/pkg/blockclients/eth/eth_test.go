package eth

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
)

type EthSuite struct {
	Client *Client
}

var cfg = config.ChainConfigurations{
	ChainHost: "https://mainnet.infura.io", // TODO change this to our mock server.. Once it exists
}

var client, _ = NewClient(cfg)
var _ = Suite(&EthSuite{Client: client})

func Test(t *testing.T) {
	TestingT(t)
}

func (s *EthSuite) TestGetBlock(c *C) {
	block, err := s.Client.getBlock(0)
	if err != nil {
		c.Error(err)
	}

	c.Check(block.Number().Int64(), Equals, int64(0))
	c.Assert(block.ParentHash().String(), Equals, "0x0000000000000000000000000000000000000000000000000000000000000000")
	c.Assert(block.Hash().String(), Equals, "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3")
}

func (s *EthSuite) TestGetCurrentBlock(c *C) {
	block, err := s.Client.getCurrentBlock()
	if err != nil {
		c.Error(err)
	}
	spew.Dump(block)
	c.Log(block.NumberU64())
	c.Check(block.NumberU64() > 9048204, Equals, true)
}

func (s *EthSuite) TestOutboundBlockHeight(c *C) {
	block, err := s.Client.getCurrentBlock()
	if err != nil {
		c.Error(err)
	}
	c.Log("block1:", block.NumberU64())
	c.Check(block.NumberU64() > 9048204, Equals, true)

	block2, err := s.Client.getBlock(block.NumberU64() * 2)
	c.Check(err, NotNil)
	c.Check(block2, IsNil)
}
