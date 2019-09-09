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
		rune   common.Amount
		token  common.Amount
		status PoolStatus
	}{
		{
			ticker: common.Ticker(""),
			rune:   common.NewAmountFromFloat(1),
			token:  common.NewAmountFromFloat(1),
			status: Enabled,
		},
		{
			ticker: common.BNBTicker,
			rune:   common.Amount(""),
			token:  common.NewAmountFromFloat(1),
			status: Enabled,
		},
		{
			ticker: common.BNBTicker,
			rune:   common.NewAmountFromFloat(1),
			token:  common.Amount(""),
			status: Enabled,
		},
		{
			ticker: common.BNBTicker,
			rune:   common.NewAmountFromFloat(1),
			token:  common.NewAmountFromFloat(1),
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
