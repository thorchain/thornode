package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgAddSuite struct{}

var _ = Suite(&MsgAddSuite{})

func (mas *MsgAddSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}
func (mas *MsgAddSuite) TestMsgAdd(c *C) {
	tx := GetRandomTx()
	addr := GetRandomBech32Addr()
	c.Check(addr.Empty(), Equals, false)
	ma := NewMsgAdd(tx, common.BNBAsset, sdk.NewUint(100000000), sdk.NewUint(100000000), addr)
	c.Check(ma.Route(), Equals, RouterKey)
	c.Check(ma.Type(), Equals, "set_add")
	err := ma.ValidateBasic()
	c.Assert(err, IsNil)
	buf := ma.GetSignBytes()
	c.Assert(buf, NotNil)
	c.Check(len(buf) > 0, Equals, true)
	signer := ma.GetSigners()
	c.Assert(signer, NotNil)
	c.Check(len(signer) > 0, Equals, true)

	inputs := []struct {
		ticker common.Asset
		rune   sdk.Uint
		asset  sdk.Uint
		txHash common.TxID
		signer sdk.AccAddress
	}{
		{
			ticker: common.Asset{},
			rune:   sdk.NewUint(100000000),
			asset:  sdk.NewUint(100000000),
			txHash: tx.ID,
			signer: addr,
		},
		{
			ticker: common.BNBAsset,
			rune:   sdk.NewUint(100000000),
			asset:  sdk.NewUint(100000000),
			txHash: common.TxID(""),
			signer: addr,
		},
		{
			ticker: common.BNBAsset,
			rune:   sdk.NewUint(100000000),
			asset:  sdk.NewUint(100000000),
			txHash: tx.ID,
			signer: sdk.AccAddress{},
		},
	}
	for _, item := range inputs {
		tx := GetRandomTx()
		tx.ID = item.txHash
		msgAdd := NewMsgAdd(tx, item.ticker, item.rune, item.asset, item.signer)
		err := msgAdd.ValidateBasic()
		c.Assert(err, NotNil)
	}
}
