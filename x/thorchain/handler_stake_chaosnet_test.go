// +build chaosnet

package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

func (HandlerStakeSuite) TestStakeRUNEOverLimit(c *C) {
	ctx, _ := setupKeeperForTest(c)
	activeNodeAccount := GetRandomNodeAccount(NodeActive)
	k := &MockStackKeeper{
		activeNodeAccount: activeNodeAccount,
		currentPool: Pool{
			BalanceRune:  sdk.ZeroUint(),
			BalanceAsset: sdk.ZeroUint(),
			Asset:        common.BNBAsset,
			PoolUnits:    sdk.ZeroUint(),
			PoolAddress:  "",
			Status:       PoolEnabled,
		},
	}
	// happy path
	stakeHandler := NewStakeHandler(k)
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
	ver := constants.SWVersion
	msgSetStake := NewMsgSetStakeData(
		tx,
		common.BNBAsset,
		sdk.NewUint(1000_000*common.One),
		sdk.NewUint(100_000*common.One),
		bnbAddr,
		bnbAddr,
		activeNodeAccount.NodeAddress)
	constAccessor := constants.NewConstantValue010()
	result := stakeHandler.Run(ctx, msgSetStake, ver, constAccessor)
	c.Assert(result.Code, Equals, CodeStakeRUNEOverLimit)
}
