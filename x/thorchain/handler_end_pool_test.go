package thorchain

import (
	"errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"

	. "gopkg.in/check.v1"
)

type HandlerEndPoolSuite struct{}

type TestEndPoolKeeper struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestEndPoolKeeper) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

var _ = Suite(&HandlerEndPoolSuite{})

func (s *HandlerEndPoolSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestEndPoolKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	txOutStore := NewTxStoreDummy()

	handler := NewEndPoolHandler(keeper, txOutStore)

	// happy path
	ver := semver.MustParse("0.1.0")
	bnbAddr := GetRandomBNBAddress()
	txHash := GetRandomTxHash()
	tx := common.NewTx(
		txHash,
		bnbAddr,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
		common.BNBGasFeeSingleton,
		"",
	)
	signer := GetRandomBech32Addr()
	msg := NewMsgEndPool(common.BNBAsset, tx, signer)
	err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// invalid msg
	msg = MsgEndPool{}
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)

	// not active node
	keeper = &TestEndPoolKeeper{
		na: GetRandomNodeAccount(NodeWhiteListed),
	}
	handler = NewEndPoolHandler(keeper, txOutStore)
	msg = NewMsgEndPool(common.BNBAsset, tx, signer)
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, Equals, notAuthorized)

}

type TestEndPoolHandleKeeper struct {
	KVStoreDummy
	currentPool       Pool
	activeNodeAccount NodeAccount
	failAddEvent      bool
	failStakeEvent    bool
	poolStaker        PoolStaker
	stakerPool        StakerPool
}

func (k *TestEndPoolHandleKeeper) PoolExist(_ sdk.Context, asset common.Asset) bool {
	return k.currentPool.Asset.Equals(asset)
}

func (k *TestEndPoolHandleKeeper) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.currentPool, nil
}

func (k *TestEndPoolHandleKeeper) SetPool(_ sdk.Context, pool Pool) error {
	k.currentPool = pool
	return nil
}

// IsActiveObserver see whether it is an active observer
func (k *TestEndPoolHandleKeeper) IsActiveObserver(_ sdk.Context, addr sdk.AccAddress) bool {
	return k.activeNodeAccount.NodeAddress.Equals(addr)
}

func (k *TestEndPoolHandleKeeper) GetNodeAccount(_ sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	if k.activeNodeAccount.NodeAddress.Equals(addr) {
		return k.activeNodeAccount, nil
	}
	return NodeAccount{}, errors.New("not exist")
}

func (k *TestEndPoolHandleKeeper) GetPoolStaker(_ sdk.Context, _ common.Asset) (PoolStaker, error) {
	return k.poolStaker, nil
}

func (k *TestEndPoolHandleKeeper) GetStakerPool(_ sdk.Context, _ common.Address) (StakerPool, error) {
	return k.stakerPool, nil
}

func (k *TestEndPoolHandleKeeper) SetStakerPool(_ sdk.Context, sp StakerPool) {
	k.stakerPool = sp
}

func (k *TestEndPoolHandleKeeper) SetPoolStaker(_ sdk.Context, ps PoolStaker) {
	k.poolStaker = ps
}

func (k *TestEndPoolHandleKeeper) AddIncompleteEvents(_ sdk.Context, _ Event) error {
	if k.failAddEvent {
		return errors.New("fail to add incomplete events")
	}
	return nil
}

func (k *TestEndPoolHandleKeeper) GetIncompleteEvents(_ sdk.Context) (Events, error) {
	if k.failStakeEvent {
		return nil, errors.New("fail to get incomplete events")
	}
	return nil, nil
}

func (k *TestEndPoolHandleKeeper) GetLastEventID(_ sdk.Context) (int64, error) {
	return 0, nil
}

func (k *TestEndPoolHandleKeeper) GetAdminConfigDefaultPoolStatus(_ sdk.Context, _ sdk.AccAddress) PoolStatus {
	return PoolEnabled
}

func (s *HandlerEndPoolSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)

	activeNodeAccount := GetRandomNodeAccount(NodeActive)
	bnbAddr := GetRandomBNBAddress()
	keeper := &TestEndPoolHandleKeeper{
		activeNodeAccount: activeNodeAccount,
		currentPool: Pool{
			BalanceRune:  sdk.ZeroUint(),
			BalanceAsset: sdk.ZeroUint(),
			Asset:        common.BNBAsset,
			PoolUnits:    sdk.ZeroUint(),
			PoolAddress:  "",
			Status:       PoolEnabled,
		},
		poolStaker: PoolStaker{
			Asset:      common.BNBAsset,
			TotalUnits: sdk.ZeroUint(),
			Stakers:    nil,
		},
		stakerPool: StakerPool{
			RuneAddress:  bnbAddr,
			AssetAddress: bnbAddr,
			PoolUnits:    nil,
		},
	}

	txOutStore := NewTxStoreDummy()
	handler := NewEndPoolHandler(keeper, txOutStore)
	ver := semver.MustParse("0.1.0")

	stakeTxHash := GetRandomTxHash()
	tx := common.NewTx(
		stakeTxHash,
		bnbAddr,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
		common.BNBGasFeeSingleton,
		"",
	)
	msgSetStake := NewMsgSetStakeData(
		tx,
		common.BNBAsset,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		bnbAddr,
		activeNodeAccount.NodeAddress)

	constAccessor := constants.GetConstantValues(ver)
	stakeHandler := NewStakeHandler(keeper)
	stakeResult := stakeHandler.Run(ctx, msgSetStake, ver, constAccessor)
	c.Assert(stakeResult.Code, Equals, sdk.CodeOK)

	p, err := keeper.GetPool(ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(p.Empty(), Equals, false)
	c.Assert(p.BalanceRune.Uint64(), Equals, msgSetStake.RuneAmount.Uint64())
	c.Assert(p.BalanceAsset.Uint64(), Equals, msgSetStake.AssetAmount.Uint64())
	c.Assert(p.Status, Equals, PoolEnabled)
	txOutStore.NewBlock(1, constAccessor)

	// EndPool again
	msgEndPool1 := NewMsgEndPool(common.BNBAsset, tx, activeNodeAccount.NodeAddress)
	result1 := handler.handle(ctx, msgEndPool1, ver, constAccessor)
	c.Assert(result1.Code, Equals, sdk.CodeOK, Commentf("%+v\n", result1))
	p1, err := keeper.GetPool(ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Check(p1.Status, Equals, PoolBootstrap)
	c.Check(p1.BalanceAsset.Uint64(), Equals, uint64(0))
	c.Check(p1.BalanceRune.Uint64(), Equals, uint64(0))
	txOut := txOutStore.GetBlockOut()
	c.Check(txOut, NotNil)
	c.Check(len(txOut.TxArray) > 0, Equals, true)
	c.Check(txOut.Height, Equals, uint64(1))
	totalAsset := sdk.ZeroUint()
	totalRune := sdk.ZeroUint()
	for _, item := range txOut.TxArray {
		c.Assert(item.ToAddress.Equals(bnbAddr), Equals, true)
		if item.Coin.Asset.IsRune() {
			totalRune = totalRune.Add(item.Coin.Amount)
		} else {
			totalAsset = totalAsset.Add(item.Coin.Amount)
		}
	}
	c.Assert(totalAsset.Equal(msgSetStake.AssetAmount), Equals, true, Commentf("%d %d", totalAsset.Uint64(), msgSetStake.AssetAmount.Uint64()))
	c.Assert(totalRune.Equal(msgSetStake.RuneAmount), Equals, true)
}
