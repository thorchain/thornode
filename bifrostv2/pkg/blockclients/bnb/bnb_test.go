package bnb

import (
	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/types/tx"

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
	c.Skip("Need a mock server!!!")
	ResultBlock, err := s.Client.getBlock(1)
	c.Assert(err, IsNil)
	c.Assert(ResultBlock, NotNil)
}

func (s *BNBSuite) TestProcessStdTx(c *C) {
	// happy path
	stdTx, err := getStdTx(
		"tbnb1yycn4mh6ffwpjf584t8lpp7c27ghu03gpvqkfj",
		"tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj",
		types.Coins{types.Coin{Denom: "BNB", Amount: 194765912}},
		"outbound:256",
	)
	c.Assert(err, IsNil)
	c.Assert(stdTx, NotNil)

	items, err := s.Client.processStdTx(stdTx)
	c.Assert(err, IsNil)
	c.Assert(items, NotNil)
	c.Assert(len(items), Equals, 1)
	c.Assert(items[0].Memo, Equals, "outbound:256")
	c.Assert(items[0].To.String(), Equals, "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj")
	c.Assert(items[0].From.String(), Equals, "tbnb1yycn4mh6ffwpjf584t8lpp7c27ghu03gpvqkfj")

	// no memo (will be refunded)
	stdTx, err = getStdTx(
		"tbnb1yycn4mh6ffwpjf584t8lpp7c27ghu03gpvqkfj",
		"tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj",
		types.Coins{types.Coin{Denom: "BNB", Amount: 194765912}},
		"",
	)
	c.Assert(err, IsNil)
	c.Assert(stdTx, NotNil)

	items, err = s.Client.processStdTx(stdTx)
	c.Assert(err, IsNil)
	c.Assert(items, NotNil)
	c.Assert(len(items), Equals, 1)
	c.Assert(items[0].Memo, Equals, "")
	c.Assert(items[0].To.String(), Equals, "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj")
	c.Assert(items[0].From.String(), Equals, "tbnb1yycn4mh6ffwpjf584t8lpp7c27ghu03gpvqkfj")

	// TODO add test for other message types

	// TODO add test with multiple `SendMsg` tx's.
}

func getStdTx(f, t string, coins []types.Coin, memo string) (tx.StdTx, error) {
	types.Network = types.TestNetwork
	from, err := types.AccAddressFromBech32(f)
	if err != nil {
		return tx.StdTx{}, err
	}
	to, err := types.AccAddressFromBech32(t)
	if err != nil {
		return tx.StdTx{}, err
	}
	transfers := []msg.Transfer{msg.Transfer{to, coins}}
	m := msg.CreateSendMsg(from, coins, transfers)
	return tx.NewStdTx([]msg.Msg{m}, nil, memo, 0, nil), nil
}
