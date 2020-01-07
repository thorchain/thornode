package thorchain

import (
	"errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type HandlerStakeSuite struct{}

var _ = Suite(&HandlerStakeSuite{})

type MockStackKeeper struct {
	KVStoreDummy
	currentPool        Pool
	activeNodeAccount  NodeAccount
	failGetPool        bool
	failGetNextEventID bool
	addedEvent         bool
}

func (m *MockStackKeeper) PoolExist(_ sdk.Context, asset common.Asset) bool {
	return m.currentPool.Asset.Equals(asset)
}

func (m *MockStackKeeper) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	if m.failGetPool {
		return Pool{}, errors.New("fail to get pool")
	}
	return m.currentPool, nil
}

func (m *MockStackKeeper) SetPool(_ sdk.Context, pool Pool) error {
	m.currentPool = pool
	return nil
}

func (m *MockStackKeeper) GetNodeAccount(_ sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	if m.activeNodeAccount.NodeAddress.Equals(addr) {
		return m.activeNodeAccount, nil
	}
	return NodeAccount{}, errors.New("not exist")
}

func (m *MockStackKeeper) GetPoolStaker(_ sdk.Context, asset common.Asset) (PoolStaker, error) {
	return PoolStaker{
		Asset:      asset,
		TotalUnits: sdk.ZeroUint(),
		Stakers:    nil,
	}, nil
}

func (m *MockStackKeeper) GetStakerPool(_ sdk.Context, addr common.Address) (StakerPool, error) {
	return StakerPool{
		RuneAddress:  addr,
		AssetAddress: addr,
		PoolUnits:    nil,
	}, nil
}

func (m *MockStackKeeper) UpsertEvent(_ sdk.Context, _ Event) error {
	if m.failGetNextEventID {
		return kaboom
	}
	m.addedEvent = true
	return nil
}

func (HandlerStakeSuite) TestStakeHandler(c *C) {
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
	preStakePool, err := k.GetPool(ctx, common.BNBAsset)
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
	constAccessor := constants.GetConstantValues(ver)
	msgSetStake := NewMsgSetStakeData(
		tx,
		common.BNBAsset,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		bnbAddr,
		activeNodeAccount.NodeAddress)
	result := stakeHandler.Run(ctx, msgSetStake, ver, constAccessor)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	postStakePool, err := k.GetPool(ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(postStakePool.BalanceAsset.String(), Equals, preStakePool.BalanceAsset.Add(msgSetStake.AssetAmount).String())
	c.Assert(postStakePool.BalanceRune.String(), Equals, preStakePool.BalanceRune.Add(msgSetStake.RuneAmount).String())
	c.Check(k.addedEvent, Equals, true)
}

func (HandlerStakeSuite) TestStakeHandler_NoPool_ShouldCreateNewPool(c *C) {
	ctx, _ := setupKeeperForTest(c)
	activeNodeAccount := GetRandomNodeAccount(NodeActive)
	k := &MockStackKeeper{
		activeNodeAccount: activeNodeAccount,
		currentPool: Pool{
			BalanceRune:  sdk.ZeroUint(),
			BalanceAsset: sdk.ZeroUint(),
			PoolUnits:    sdk.ZeroUint(),
		},
	}
	// happy path
	stakeHandler := NewStakeHandler(k)
	preStakePool, err := k.GetPool(ctx, common.BNBAsset)
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
	constAccessor := constants.GetConstantValues(ver)
	msgSetStake := NewMsgSetStakeData(
		tx,
		common.BNBAsset,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		bnbAddr,
		activeNodeAccount.NodeAddress)
	result := stakeHandler.Run(ctx, msgSetStake, ver, constAccessor)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	postStakePool, err := k.GetPool(ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(postStakePool.BalanceAsset.String(), Equals, preStakePool.BalanceAsset.Add(msgSetStake.AssetAmount).String())
	c.Assert(postStakePool.BalanceRune.String(), Equals, preStakePool.BalanceRune.Add(msgSetStake.RuneAmount).String())
	c.Check(k.addedEvent, Equals, true)

	// bad version
	result = stakeHandler.Run(ctx, msgSetStake, semver.Version{}, constAccessor)
	c.Assert(result.Code, Equals, CodeBadVersion)
}
func (HandlerStakeSuite) TestStakeHandlerValidation(c *C) {
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
			expectedResult: CodeStakeFailValidation,
		},
		{
			name:           "empty asset should fail",
			msg:            NewMsgSetStakeData(GetRandomTx(), common.Asset{}, sdk.NewUint(common.One*5), sdk.NewUint(common.One*5), GetRandomBNBAddress(), GetRandomBNBAddress(), GetRandomNodeAccount(NodeActive).NodeAddress),
			expectedResult: CodeStakeFailValidation,
		},
		{
			name:           "empty RUNE address should fail",
			msg:            NewMsgSetStakeData(GetRandomTx(), common.BNBAsset, sdk.NewUint(common.One*5), sdk.NewUint(common.One*5), common.NoAddress, GetRandomBNBAddress(), GetRandomNodeAccount(NodeActive).NodeAddress),
			expectedResult: CodeStakeFailValidation,
		},
		{
			name:           "empty ASSET address should fail",
			msg:            NewMsgSetStakeData(GetRandomTx(), common.BTCAsset, sdk.NewUint(common.One*5), sdk.NewUint(common.One*5), GetRandomBNBAddress(), common.NoAddress, GetRandomNodeAccount(NodeActive).NodeAddress),
			expectedResult: CodeStakeFailValidation,
		},
	}
	ver := semver.MustParse("0.1.0")
	constAccessor := constants.GetConstantValues(ver)
	for _, item := range testCases {
		stakeHandler := NewStakeHandler(k)
		result := stakeHandler.Run(ctx, item.msg, ver, constAccessor)
		c.Assert(result.Code, Equals, item.expectedResult, Commentf(item.name))
	}
}
func (HandlerStakeSuite) TestHandlerStakeFailScenario(c *C) {
	ctx, _ := setupKeeperForTest(c)
	activeNodeAccount := GetRandomNodeAccount(NodeActive)
	emptyPool := Pool{
		BalanceRune:  sdk.ZeroUint(),
		BalanceAsset: sdk.ZeroUint(),
		Asset:        common.BNBAsset,
		PoolUnits:    sdk.ZeroUint(),
		PoolAddress:  "",
		Status:       PoolEnabled,
	}

	testCases := []struct {
		name           string
		k              Keeper
		expectedResult sdk.CodeType
	}{
		{
			name: "fail to get pool should fail stake",
			k: &MockStackKeeper{
				activeNodeAccount: activeNodeAccount,
				currentPool:       emptyPool,
				failGetPool:       true,
			},
			expectedResult: sdk.CodeInternal,
		},
		{
			name: "suspended pool should fail stake",
			k: &MockStackKeeper{
				activeNodeAccount: activeNodeAccount,
				currentPool: Pool{
					BalanceRune:  sdk.ZeroUint(),
					BalanceAsset: sdk.ZeroUint(),
					Asset:        common.BNBAsset,
					PoolUnits:    sdk.ZeroUint(),
					Status:       PoolSuspended,
				},
			},
			expectedResult: CodeInvalidPoolStatus,
		},
		{
			name: "fail to get next event id should fail stake",
			k: &MockStackKeeper{
				activeNodeAccount:  activeNodeAccount,
				currentPool:        emptyPool,
				failGetNextEventID: true,
			},
			expectedResult: sdk.CodeInternal,
		},
	}
	for _, tc := range testCases {
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
		constAccessor := constants.GetConstantValues(ver)
		msgSetStake := NewMsgSetStakeData(
			tx,
			common.BNBAsset,
			sdk.NewUint(100*common.One),
			sdk.NewUint(100*common.One),
			bnbAddr,
			bnbAddr,
			activeNodeAccount.NodeAddress)
		stakeHandler := NewStakeHandler(tc.k)
		result := stakeHandler.Run(ctx, msgSetStake, ver, constAccessor)
		c.Assert(result.Code, Equals, tc.expectedResult, Commentf(tc.name))
	}
}
