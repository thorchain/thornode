package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgStakeSuite struct{}

var _ = Suite(&MsgStakeSuite{})

func (MsgStakeSuite) TestMsgStake(c *C) {
	addr := GetRandomBech32Addr()
	c.Check(addr.Empty(), Equals, false)
	bnbAddress := GetRandomBNBAddress()
	txID := GetRandomTxHash()
	c.Check(txID.IsEmpty(), Equals, false)
	m := NewMsgSetStakeData(common.BNBAsset, sdk.NewUint(100000000), sdk.NewUint(100000000), bnbAddress, txID, addr)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_stakedata")

	inputs := []struct {
		asset         common.Asset
		r             sdk.Uint
		token         sdk.Uint
		publicAddress common.Address
		txHash        common.TxID
		signer        sdk.AccAddress
	}{
		{
			asset:         common.Asset{},
			r:             sdk.NewUint(100000000),
			token:         sdk.NewUint(100000000),
			publicAddress: bnbAddress,
			txHash:        txID,
			signer:        addr,
		},
		{
			asset:         common.BNBAsset,
			r:             sdk.NewUint(100000000),
			token:         sdk.NewUint(100000000),
			publicAddress: common.NoAddress,
			txHash:        txID,
			signer:        addr,
		},
		{
			asset:         common.BNBAsset,
			r:             sdk.NewUint(100000000),
			token:         sdk.NewUint(100000000),
			publicAddress: bnbAddress,
			txHash:        common.TxID(""),
			signer:        addr,
		},
		{
			asset:         common.BNBAsset,
			r:             sdk.NewUint(100000000),
			token:         sdk.NewUint(100000000),
			publicAddress: bnbAddress,
			txHash:        txID,
			signer:        sdk.AccAddress{},
		},
	}
	for i, item := range inputs {
		m := NewMsgSetStakeData(item.asset, item.r, item.token, item.publicAddress, item.txHash, item.signer)
		c.Assert(m.ValidateBasic(), NotNil, Commentf("%d) %s\n", i, m))
	}
}

func EnsureMsgBasicCorrect(m sdk.Msg, c *C) {
	signers := m.GetSigners()
	c.Check(signers, NotNil)
	c.Check(len(signers), Equals, 1)
	c.Check(m.ValidateBasic(), IsNil)
	c.Check(m.Route(), Equals, RouterKey)
	c.Check(m.GetSignBytes(), NotNil)
}
