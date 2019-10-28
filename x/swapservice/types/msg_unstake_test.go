package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type MsgUnstakeSuite struct{}

var _ = Suite(&MsgUnstakeSuite{})

func (MsgUnstakeSuite) TestMsgUnstake(c *C) {
	txID := GetRandomTxHash()
	bnb := GetRandomBNBAddress()
	acc1 := GetRandomBech32Addr()
	m := NewMsgSetUnStake(bnb, sdk.NewUint(10000), common.BNBAsset, txID, acc1)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_unstake")

	inputs := []struct {
		publicAddress       common.Address
		withdrawBasisPoints sdk.Uint
		asset               common.Asset
		requestTxHash       common.TxID
		signer              sdk.AccAddress
	}{
		{
			publicAddress:       common.NoAddress,
			withdrawBasisPoints: sdk.NewUint(10000),
			asset:               common.BNBAsset,
			requestTxHash:       txID,
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.NewUint(12000),
			asset:               common.BNBAsset,
			requestTxHash:       txID,
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.ZeroUint(),
			asset:               common.BNBAsset,
			requestTxHash:       txID,
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.NewUint(10000),
			asset:               common.Asset{},
			requestTxHash:       txID,
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.NewUint(10000),
			asset:               common.BNBAsset,
			requestTxHash:       common.TxID(""),
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.NewUint(10000),
			asset:               common.BNBAsset,
			requestTxHash:       txID,
			signer:              sdk.AccAddress{},
		},
	}
	for _, item := range inputs {
		m := NewMsgSetUnStake(item.publicAddress, item.withdrawBasisPoints, item.asset, item.requestTxHash, item.signer)
		c.Assert(m.ValidateBasic(), NotNil)
	}
}
