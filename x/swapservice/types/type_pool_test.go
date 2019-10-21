package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type PoolTestSuite struct{}

var _ = Suite(&PoolTestSuite{})

func (PoolTestSuite) TestPool(c *C) {
	p := NewPool()
	c.Check(p.Empty(), Equals, true)
	c.Check(p.TokenPriceInRune(), Equals, float64(0))
	p.Ticker = common.BNBTicker
	c.Check(p.Empty(), Equals, false)
	p.BalanceRune = sdk.NewUint(100 * common.One)
	p.BalanceToken = sdk.NewUint(50 * common.One)
	c.Check(p.TokenPriceInRune(), Equals, 2.0)
	c.Log(p.String())

	addr := GetRandomBech32Addr()
	bnbAddress := GetRandomBNBAddress()
	txID := GetRandomTxHash()

	m := NewMsgSwap(txID, common.RuneA1FTicker, common.BNBTicker, sdk.NewUint(1), bnbAddress, bnbAddress, sdk.NewUint(2), addr)

	c.Check(p.EnsureValidPoolStatus(m), IsNil)
	msgNoop := NewMsgNoOp(addr)
	c.Check(p.EnsureValidPoolStatus(msgNoop), IsNil)
	p.Status = Enabled
	c.Check(p.EnsureValidPoolStatus(m), IsNil)
	p.Status = PoolStatus(100)
	c.Check(p.EnsureValidPoolStatus(msgNoop), NotNil)

	p.Status = Suspended
	c.Check(p.EnsureValidPoolStatus(msgNoop), NotNil)

}

func (PoolTestSuite) TestPoolStatus(c *C) {
	inputs := []string{
		"enabled", "bootstrap", "suspended", "whatever",
	}
	for _, item := range inputs {
		ps := GetPoolStatus(item)
		c.Assert(ps.Valid(), IsNil)
	}
	var ps PoolStatus
	err := json.Unmarshal([]byte(`"Enabled"`), &ps)
	c.Assert(err, IsNil)
	c.Check(ps == Enabled, Equals, true)
	err = json.Unmarshal([]byte(`{asdf}`), &ps)
	c.Assert(err, NotNil)
}
