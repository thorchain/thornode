package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgEndPoolTestSuite struct{}

var _ = Suite(&MsgEndPoolTestSuite{})

func (MsgEndPoolTestSuite) TestMsgEndPool(c *C) {
	asset := common.BNBAsset
	bnb := GetRandomBNBAddress()
	addr := GetRandomBech32Addr()
	c.Check(addr.Empty(), Equals, false)
	txID := GetRandomTxHash()
	tx := common.NewTx(
		txID,
		bnb,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
		"",
	)
	msgEndPool := NewMsgEndPool(asset, tx, addr)
	c.Assert(msgEndPool.Route(), Equals, RouterKey)
	c.Assert(msgEndPool.Type(), Equals, "set_poolend")
	c.Assert(msgEndPool.ValidateBasic(), IsNil)
	c.Assert(len(msgEndPool.GetSignBytes()) > 0, Equals, true)
	c.Assert(msgEndPool.GetSigners(), NotNil)
	c.Assert(msgEndPool.GetSigners()[0].String(), Equals, addr.String())

	errEndPool := NewMsgEndPool(common.Asset{}, tx, addr)
	c.Assert(errEndPool.ValidateBasic(), NotNil)
	errEndPool1 := NewMsgEndPool(common.RuneAsset(), tx, addr)
	c.Assert(errEndPool1.ValidateBasic(), NotNil)
	tx.ID = ""
	errEndPool2 := NewMsgEndPool(common.BNBAsset, tx, addr)
	c.Assert(errEndPool2.ValidateBasic(), NotNil)
	tx.ID = txID
	tx.FromAddress = ""
	errEndPool3 := NewMsgEndPool(common.BNBAsset, tx, addr)
	c.Assert(errEndPool3.ValidateBasic(), NotNil)

}
