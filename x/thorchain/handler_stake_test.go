package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type HandlerStakeSuite struct{}

var _ = Suite(&HandlerStakeSuite{})

func (HandlerStakeSuite) TestStakeHandler(c *C) {
	w := getHandlerTestWrapper(c, 1, true, true)
	// happy path
	stakeHandler := NewStakeHandler(w.keeper)
	preStakePool, err := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	bnbAddr := GetRandomBNBAddress()
	stakeTxHash := GetRandomTxHash()
	tx := common.NewTx(
		stakeTxHash,
		bnbAddr,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(common.One*5))},
		common.BNBGasFeeSingleton,
		"stake:BNB",
	)
	ver := semver.MustParse("0.1.0")
	msgSetStake := NewMsgSetStakeData(
		tx,
		common.BNBAsset,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		bnbAddr,
		w.activeNodeAccount.NodeAddress)
	result := stakeHandler.Run(w.ctx, msgSetStake, ver)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	postStakePool, err := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(postStakePool.BalanceAsset.String(), Equals, preStakePool.BalanceAsset.Add(msgSetStake.AssetAmount).String())
	c.Assert(postStakePool.BalanceRune.String(), Equals, preStakePool.BalanceRune.Add(msgSetStake.RuneAmount).String())
}

func (HandlerStakeSuite) TestStakeHandler_NoPool_ShouldCreateNewPool(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	// happy path
	stakeHandler := NewStakeHandler(w.keeper)
	preStakePool, err := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(preStakePool.Empty(), Equals, true)
	bnbAddr := GetRandomBNBAddress()
	stakeTxHash := GetRandomTxHash()
	tx := common.NewTx(
		stakeTxHash,
		bnbAddr,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(common.One*5))},
		common.BNBGasFeeSingleton,
		"stake:BNB",
	)
	ver := semver.MustParse("0.1.0")
	msgSetStake := NewMsgSetStakeData(
		tx,
		common.BNBAsset,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		bnbAddr,
		w.activeNodeAccount.NodeAddress)
	result := stakeHandler.Run(w.ctx, msgSetStake, ver)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	postStakePool, err := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(postStakePool.BalanceAsset.String(), Equals, preStakePool.BalanceAsset.Add(msgSetStake.AssetAmount).String())
	c.Assert(postStakePool.BalanceRune.String(), Equals, preStakePool.BalanceRune.Add(msgSetStake.RuneAmount).String())

	// bad version
	result = stakeHandler.Run(w.ctx, msgSetStake, semver.Version{})
	c.Assert(result.Code, Equals, CodeBadVersion)
}
func (HandlerStakeSuite) TestStakeHandlerValidation(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	testCases := []struct {
		name           string
		msg            MsgSetStakeData
		expectedResult sdk.CodeType
	}{
		{
			name:           "not signed by an active node account should fail",
			msg:            NewMsgSetStakeData(GetRandomTx(), common.BNBAsset, sdk.NewUint(common.One*5), sdk.NewUint(common.One*5), GetRandomBNBAddress(), GetRandomBNBAddress(), GetRandomNodeAccount(NodeActive).NodeAddress),
			expectedResult: sdk.CodeUnauthorized,
		},
		{
			name:           "empty signer should fail",
			msg:            NewMsgSetStakeData(GetRandomTx(), common.BNBAsset, sdk.NewUint(common.One*5), sdk.NewUint(common.One*5), GetRandomBNBAddress(), GetRandomBNBAddress(), sdk.AccAddress{}),
			expectedResult: sdk.CodeInvalidAddress,
		},
		{
			name:           "empty asset should fail",
			msg:            NewMsgSetStakeData(GetRandomTx(), common.Asset{}, sdk.NewUint(common.One*5), sdk.NewUint(common.One*5), GetRandomBNBAddress(), GetRandomBNBAddress(), GetRandomNodeAccount(NodeActive).NodeAddress),
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name:           "empty RUNE address should fail",
			msg:            NewMsgSetStakeData(GetRandomTx(), common.BNBAsset, sdk.NewUint(common.One*5), sdk.NewUint(common.One*5), common.NoAddress, GetRandomBNBAddress(), GetRandomNodeAccount(NodeActive).NodeAddress),
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name:           "empty ASSET address should fail",
			msg:            NewMsgSetStakeData(GetRandomTx(), common.BTCAsset, sdk.NewUint(common.One*5), sdk.NewUint(common.One*5), GetRandomBNBAddress(), common.NoAddress, GetRandomNodeAccount(NodeActive).NodeAddress),
			expectedResult: sdk.CodeUnknownRequest,
		},
	}

	for _, item := range testCases {
		ver := semver.MustParse("0.1.0")
		stakeHandler := NewStakeHandler(w.keeper)
		result := stakeHandler.Run(w.ctx, item.msg, ver)
		c.Assert(result.Code, Equals, item.expectedResult, Commentf(item.name))
	}
}
