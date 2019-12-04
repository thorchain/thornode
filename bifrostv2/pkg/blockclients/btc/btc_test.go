package btc

import (
	"testing"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
)

type BTCSuite struct {
	Client *Client
}

var cfg = config.BTCConfiguration{
	CommonBlockChainConfigurations: config.CommonBlockChainConfigurations{
		ChainHost: "localhost:8332", // TODO change this to our mock server... Once it exists
		UserName:  "bitcoin",
		Password:  "password",
	},
	HTTPostMode: true,
	DisableTLS:  true,
}

var client, _ = NewClient(cfg)
var _ = Suite(&BTCSuite{Client: client})

func Test(t *testing.T) {
	TestingT(t)
}

func (s *BTCSuite) TestGetBlock(c *C) {
	block, err := s.Client.getBlock(0)
	if err != nil {
		c.Error(err)
	}

	c.Assert(block.Header.PrevBlock.String(), Equals, "0000000000000000000000000000000000000000000000000000000000000000")
	c.Assert(block.Header.MerkleRoot.String(), Equals, "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b")
}

func (s *BTCSuite) TestGetBlockHash(c *C) {
	hash, err := s.Client.getBlockHash(0)
	if err != nil {
		c.Error(err)
	}

	c.Assert(hash.String(), Equals, "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")
}
