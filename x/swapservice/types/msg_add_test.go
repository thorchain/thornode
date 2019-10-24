package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgAddSuite struct{}

var _ = Suite(&MsgAddSuite{})

func (mas *MsgAddSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}
func (mas *MsgAddSuite) TestMsgAdd(c *C) {
	txId := GetRandomTxHash()
	c.Check(txId.IsEmpty(), Equals, false)
	addr := GetRandomBech32Addr()
	c.Check(addr.Empty(), Equals, false)
	ma := NewMsgAdd(common.BNBAsset, sdk.NewUint(100000000), sdk.NewUint(100000000), txId, addr)
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
		token  sdk.Uint
		txHash common.TxID
		signer sdk.AccAddress
	}{
		{
			ticker: common.Asset{},
			rune:   sdk.NewUint(100000000),
			token:  sdk.NewUint(100000000),
			txHash: txId,
			signer: addr,
		},
		{
			ticker: common.BNBAsset,
			rune:   sdk.NewUint(100000000),
			token:  sdk.NewUint(100000000),
			txHash: common.TxID(""),
			signer: addr,
		},
		{
			ticker: common.BNBAsset,
			rune:   sdk.NewUint(100000000),
			token:  sdk.NewUint(100000000),
			txHash: txId,
			signer: sdk.AccAddress{},
		},
	}
	for _, item := range inputs {
		msgAdd := NewMsgAdd(item.ticker, item.rune, item.token, item.txHash, item.signer)
		err := msgAdd.ValidateBasic()
		c.Assert(err, NotNil)
	}
}
