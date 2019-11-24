package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type PoolTestSuite struct{}

var _ = Suite(&PoolTestSuite{})

func (PoolTestSuite) TestPool(c *C) {
	p := NewPool()
	c.Check(p.Empty(), Equals, true)
	p.Asset = common.BNBAsset
	c.Check(p.Empty(), Equals, false)
	p.BalanceRune = sdk.NewUint(100 * common.One)
	p.BalanceAsset = sdk.NewUint(50 * common.One)
	c.Check(p.AssetValueInRune(sdk.NewUint(25*common.One)).Equal(sdk.NewUint(50*common.One)), Equals, true)
	c.Check(p.RuneValueInAsset(sdk.NewUint(50*common.One)).Equal(sdk.NewUint(25*common.One)), Equals, true)
	c.Log(p.String())

	signer := GetRandomBech32Addr()
	bnbAddress := GetRandomBNBAddress()
	txID := GetRandomTxHash()

	tx := common.NewTx(
		txID,
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(1)),
		},
		common.BNBGasFeeSingleton,
		"",
	)
	m := NewMsgSwap(tx, common.BNBAsset, bnbAddress, sdk.NewUint(2), signer)

	c.Check(p.EnsureValidPoolStatus(m), IsNil)
	msgNoop := NewMsgNoOp(signer)
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
