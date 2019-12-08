package thorchain

import (
	"errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type HandlerStakeSuite struct{}

var _ = Suite(&HandlerStakeSuite{})

type MockStackKeeper struct {
	KVStoreDummy
	currentPool       Pool
	activeNodeAccount NodeAccount
	failGetPool       bool
	failAddEvent      bool
	failStakeEvent    bool
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
func (m *MockStackKeeper) AddIncompleteEvents(_ sdk.Context, _ Event) error {
	if m.failAddEvent {
		return errors.New("fail to add incomplete events")
	}
	return nil
}
func (m *MockStackKeeper) GetIncompleteEvents(_ sdk.Context) (Events, error) {
	if m.failStakeEvent {
		return nil, errors.New("fail to get incomplete events")
	}
	return nil, nil
}
func (m *MockStackKeeper) GetLastEventID(_ sdk.Context) (int64, error) {
	return 0, nil
}
func (HandlerStakeSuite) TestStakeHandler(c *C) {
	ctx, _ := setupKeeperForTest(c)
	activeNodeAccount := GetRandomNodeAccount(NodeActive)
	k := &MockStackKeeper{
		activeNodeAccount: activeNodeAccount,
		currentPool: Pool{
			BalanceRune:         sdk.ZeroUint(),
			BalanceAsset:        sdk.ZeroUint(),
			Asset:               common.BNBAsset,
			PoolUnits:           sdk.ZeroUint(),
			PoolAddress:         "",
			Status:              PoolEnabled,
			ExpiryInBlockHeight: 0,
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
	msgSetStake := NewMsgSetStakeData(
		tx,
		common.BNBAsset,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		bnbAddr,
		activeNodeAccount.NodeAddress)
	result := stakeHandler.Run(ctx, msgSetStake, ver)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	postStakePool, err := k.GetPool(ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(postStakePool.BalanceAsset.String(), Equals, preStakePool.BalanceAsset.Add(msgSetStake.AssetAmount).String())
	c.Assert(postStakePool.BalanceRune.String(), Equals, preStakePool.BalanceRune.Add(msgSetStake.RuneAmount).String())
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
	msgSetStake := NewMsgSetStakeData(
		tx,
		common.BNBAsset,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		bnbAddr,
		activeNodeAccount.NodeAddress)
	result := stakeHandler.Run(ctx, msgSetStake, ver)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	postStakePool, err := k.GetPool(ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(postStakePool.BalanceAsset.String(), Equals, preStakePool.BalanceAsset.Add(msgSetStake.AssetAmount).String())
	c.Assert(postStakePool.BalanceRune.String(), Equals, preStakePool.BalanceRune.Add(msgSetStake.RuneAmount).String())

	// bad version
	result = stakeHandler.Run(ctx, msgSetStake, semver.Version{})
	c.Assert(result.Code, Equals, CodeBadVersion)
}
func (HandlerStakeSuite) TestStakeHandlerValidation(c *C) {
	ctx, _ := setupKeeperForTest(c)
	activeNodeAccount := GetRandomNodeAccount(NodeActive)
	k := &MockStackKeeper{
		activeNodeAccount: activeNodeAccount,
		currentPool: Pool{
			BalanceRune:         sdk.ZeroUint(),
			BalanceAsset:        sdk.ZeroUint(),
			Asset:               common.BNBAsset,
			PoolUnits:           sdk.ZeroUint(),
			PoolAddress:         "",
			Status:              PoolEnabled,
			ExpiryInBlockHeight: 0,
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
		stakeHandler := NewStakeHandler(k)
		result := stakeHandler.Run(ctx, item.msg, ver)
		c.Assert(result.Code, Equals, item.expectedResult, Commentf(item.name))
	}
}
func (HandlerStakeSuite) TestHandlerStakeFailScenario(c *C) {
	ctx, _ := setupKeeperForTest(c)
	activeNodeAccount := GetRandomNodeAccount(NodeActive)
	emptyPool := Pool{
		BalanceRune:         sdk.ZeroUint(),
		BalanceAsset:        sdk.ZeroUint(),
		Asset:               common.BNBAsset,
		PoolUnits:           sdk.ZeroUint(),
		PoolAddress:         "",
		Status:              PoolEnabled,
		ExpiryInBlockHeight: 0,
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
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name: "fail to get add incomplete event should fail stake",
			k: &MockStackKeeper{
				activeNodeAccount: activeNodeAccount,
				currentPool:       emptyPool,
				failAddEvent:      true,
			},
			expectedResult: sdk.CodeInternal,
		},
		{
			name: "fail to get  incomplete event should fail stake",
			k: &MockStackKeeper{
				activeNodeAccount: activeNodeAccount,
				currentPool:       emptyPool,
				failStakeEvent:    true,
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
		msgSetStake := NewMsgSetStakeData(
			tx,
			common.BNBAsset,
			sdk.NewUint(100*common.One),
			sdk.NewUint(100*common.One),
			bnbAddr,
			bnbAddr,
			activeNodeAccount.NodeAddress)
		stakeHandler := NewStakeHandler(tc.k)
		result := stakeHandler.Run(ctx, msgSetStake, ver)
		c.Assert(result.Code, Equals, tc.expectedResult, Commentf(tc.name))
	}
}
