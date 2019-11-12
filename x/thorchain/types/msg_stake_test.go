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
	assetAddress := GetRandomBNBAddress()
	txID := GetRandomTxHash()
	c.Check(txID.IsEmpty(), Equals, false)
	tx := common.NewTx(
		txID,
		bnbAddress,
		GetRandomBNBAddress(),
		common.Coins{
			common.NewCoin(common.BTCAsset, sdk.NewUint(100000000)),
		},
		"",
	)
	m := NewMsgSetStakeData(tx, common.BNBAsset, sdk.NewUint(100000000), sdk.NewUint(100000000), bnbAddress, assetAddress, addr)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_stakedata")

	inputs := []struct {
		asset     common.Asset
		r         sdk.Uint
		amt       sdk.Uint
		runeAddr  common.Address
		assetAddr common.Address
		txHash    common.TxID
		signer    sdk.AccAddress
	}{
		{
			asset:     common.Asset{},
			r:         sdk.NewUint(100000000),
			amt:       sdk.NewUint(100000000),
			runeAddr:  bnbAddress,
			assetAddr: assetAddress,
			txHash:    txID,
			signer:    addr,
		},
		{
			asset:     common.BNBAsset,
			r:         sdk.NewUint(100000000),
			amt:       sdk.NewUint(100000000),
			runeAddr:  common.NoAddress,
			assetAddr: common.NoAddress,
			txHash:    txID,
			signer:    addr,
		},
		{
			asset:     common.BNBAsset,
			r:         sdk.NewUint(100000000),
			amt:       sdk.NewUint(100000000),
			runeAddr:  bnbAddress,
			assetAddr: assetAddress,
			txHash:    common.TxID(""),
			signer:    addr,
		},
		{
			asset:     common.BNBAsset,
			r:         sdk.NewUint(100000000),
			amt:       sdk.NewUint(100000000),
			runeAddr:  bnbAddress,
			assetAddr: assetAddress,
			txHash:    txID,
			signer:    sdk.AccAddress{},
		},
	}
	for i, item := range inputs {
		tx := common.NewTx(
			item.txHash,
			item.runeAddr,
			GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(item.asset, item.r),
			},
			"",
		)
		m := NewMsgSetStakeData(tx, item.asset, item.r, item.amt, item.runeAddr, item.assetAddr, item.signer)
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
