package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgSetPoolDataSuite struct{}

var _ = Suite(&MsgSetPoolDataSuite{})

func (MsgSetPoolDataSuite) TestMsgSetPoolData(c *C) {
	addr := GetRandomBech32Addr()
	c.Check(addr.Empty(), Equals, false)
	m := NewMsgSetPoolData(common.BNBAsset, Enabled, addr)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_pooldata")

	inputs := []struct {
		asset  common.Asset
		rAmt   sdk.Uint
		aAmt   sdk.Uint
		status PoolStatus
	}{
		{
			asset:  common.Asset{},
			rAmt:   sdk.NewUint(100000000),
			aAmt:   sdk.NewUint(100000000),
			status: Enabled,
		},

		{
			asset:  common.BNBAsset,
			rAmt:   sdk.NewUint(100000000),
			aAmt:   sdk.NewUint(100000000),
			status: PoolStatus(-1),
		},
	}

	for _, item := range inputs {
		m := NewMsgSetPoolData(item.asset, item.status, addr)
		m.BalanceRune = item.rAmt
		m.BalanceAsset = item.aAmt
		c.Assert(m.ValidateBasic(), NotNil)
	}
}
