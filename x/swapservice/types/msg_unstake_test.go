package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type MsgUnstakeSuite struct{}

var _ = Suite(&MsgUnstakeSuite{})

func (MsgUnstakeSuite) TestMsgUnstake(c *C) {
	txID := GetRandomTxHash()
	bnb := GetRandomBNBAddress()
	acc1 := GetRandomBech32Addr()
	m := NewMsgSetUnStake(bnb, sdk.NewUint(10000), common.BNBTicker, txID, acc1)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_unstake")

	inputs := []struct {
		publicAddress       common.Address
		withdrawBasisPoints sdk.Uint
		ticker              common.Ticker
		requestTxHash       common.TxID
		signer              sdk.AccAddress
	}{
		{
			publicAddress:       common.NoAddress,
			withdrawBasisPoints: sdk.NewUint(10000),
			ticker:              common.BNBTicker,
			requestTxHash:       txID,
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.NewUint(12000),
			ticker:              common.BNBTicker,
			requestTxHash:       txID,
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.ZeroUint(),
			ticker:              common.BNBTicker,
			requestTxHash:       txID,
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.NewUint(10000),
			ticker:              common.Ticker(""),
			requestTxHash:       txID,
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.NewUint(10000),
			ticker:              common.BNBTicker,
			requestTxHash:       common.TxID(""),
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.NewUint(10000),
			ticker:              common.BNBTicker,
			requestTxHash:       txID,
			signer:              sdk.AccAddress{},
		},
	}
	for _, item := range inputs {
		m := NewMsgSetUnStake(item.publicAddress, item.withdrawBasisPoints, item.ticker, item.requestTxHash, item.signer)
		c.Assert(m.ValidateBasic(), NotNil)
	}
}
