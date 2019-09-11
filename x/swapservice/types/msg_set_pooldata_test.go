package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type MsgSetPoolDataSuite struct{}

var _ = Suite(&MsgSetPoolDataSuite{})

func (MsgSetPoolDataSuite) TestMsgSetPoolData(c *C) {
	addr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	c.Check(addr.Empty(), Equals, false)
	m := NewMsgSetPoolData(common.BNBTicker, Enabled, addr)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_pooldata")

	inputs := []struct {
		ticker common.Ticker
		rune   sdk.Uint
		token  sdk.Uint
		status PoolStatus
	}{
		{
			ticker: common.Ticker(""),
			rune:   sdk.NewUint(100000000),
			token:  sdk.NewUint(100000000),
			status: Enabled,
		},

		{
			ticker: common.BNBTicker,
			rune:   sdk.NewUint(100000000),
			token:  sdk.NewUint(100000000),
			status: PoolStatus(-1),
		},
	}

	for _, item := range inputs {
		m := NewMsgSetPoolData(item.ticker, item.status, addr)
		m.BalanceRune = item.rune
		m.BalanceToken = item.token
		c.Assert(m.ValidateBasic(), NotNil)
	}
}
