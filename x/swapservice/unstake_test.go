package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/x/swapservice/mocks"
)

type UnstakeSuite struct{}

var _ = Suite(&UnstakeSuite{})

func (s *UnstakeSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (s UnstakeSuite) TestCalculateUnsake(c *C) {
	inputs := []struct {
		name                  string
		poolUnit              sdk.Uint
		poolRune              sdk.Uint
		poolAsset             sdk.Uint
		stakerUnit            sdk.Uint
		percentage            sdk.Uint
		expectedWithdrawRune  sdk.Uint
		expectedWithdrawAsset sdk.Uint
		expectedUnitLeft      sdk.Uint
		expectedErr           error
	}{
		{
			name:                  "zero-poolunit",
			poolUnit:              sdk.ZeroUint(),
			poolRune:              sdk.ZeroUint(),
			poolAsset:             sdk.ZeroUint(),
			stakerUnit:            sdk.ZeroUint(),
			percentage:            sdk.ZeroUint(),
			expectedWithdrawRune:  sdk.ZeroUint(),
			expectedWithdrawAsset: sdk.ZeroUint(),
			expectedUnitLeft:      sdk.ZeroUint(),
			expectedErr:           errors.New("poolUnits can't be zero"),
		},

		{
			name:                  "zero-poolrune",
			poolUnit:              sdk.NewUint(500 * common.One),
			poolRune:              sdk.ZeroUint(),
			poolAsset:             sdk.ZeroUint(),
			stakerUnit:            sdk.ZeroUint(),
			percentage:            sdk.ZeroUint(),
			expectedWithdrawRune:  sdk.ZeroUint(),
			expectedWithdrawAsset: sdk.ZeroUint(),
			expectedUnitLeft:      sdk.ZeroUint(),
			expectedErr:           errors.New("pool rune balance can't be zero"),
		},

		{
			name:                  "zero-poolasset",
			poolUnit:              sdk.NewUint(500 * common.One),
			poolRune:              sdk.NewUint(500 * common.One),
			poolAsset:             sdk.ZeroUint(),
			stakerUnit:            sdk.ZeroUint(),
			percentage:            sdk.ZeroUint(),
			expectedWithdrawRune:  sdk.ZeroUint(),
			expectedWithdrawAsset: sdk.ZeroUint(),
			expectedUnitLeft:      sdk.ZeroUint(),
			expectedErr:           errors.New("pool asset balance can't be zero"),
		},
		{
			name:                  "negative-stakerUnit",
			poolUnit:              sdk.NewUint(500 * common.One),
			poolRune:              sdk.NewUint(500 * common.One),
			poolAsset:             sdk.NewUint(5100 * common.One),
			stakerUnit:            sdk.ZeroUint(),
			percentage:            sdk.ZeroUint(),
			expectedWithdrawRune:  sdk.ZeroUint(),
			expectedWithdrawAsset: sdk.ZeroUint(),
			expectedUnitLeft:      sdk.ZeroUint(),
			expectedErr:           errors.New("staker unit can't be zero"),
		},

		{
			name:                  "percentage-larger-than-100",
			poolUnit:              sdk.NewUint(500 * common.One),
			poolRune:              sdk.NewUint(500 * common.One),
			poolAsset:             sdk.NewUint(500 * common.One),
			stakerUnit:            sdk.NewUint(100 * common.One),
			percentage:            sdk.NewUint(12000),
			expectedWithdrawRune:  sdk.ZeroUint(),
			expectedWithdrawAsset: sdk.ZeroUint(),
			expectedUnitLeft:      sdk.ZeroUint(),
			expectedErr:           errors.Errorf("withdraw basis point %s is not valid", sdk.NewUint(12000)),
		},
		{
			name:                  "unstake-1",
			poolUnit:              sdk.NewUint(700 * common.One),
			poolRune:              sdk.NewUint(700 * common.One),
			poolAsset:             sdk.NewUint(700 * common.One),
			stakerUnit:            sdk.NewUint(200 * common.One),
			percentage:            sdk.NewUint(10000),
			expectedUnitLeft:      sdk.ZeroUint(),
			expectedWithdrawAsset: sdk.NewUint(200 * common.One),
			expectedWithdrawRune:  sdk.NewUint(200 * common.One),
			expectedErr:           nil,
		},
	}

	for _, item := range inputs {
		c.Logf("name:%s", item.name)
		withDrawRune, withDrawAsset, unitAfter, err := calculateUnstake(item.poolUnit, item.poolRune, item.poolAsset, item.stakerUnit, item.percentage)
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}
		c.Logf("expected rune:%s,rune:%s", item.expectedWithdrawRune, withDrawRune)
		c.Check(item.expectedWithdrawRune.Uint64(), Equals, withDrawRune.Uint64())
		c.Check(item.expectedWithdrawAsset.Uint64(), Equals, withDrawAsset.Uint64())
		c.Check(item.expectedUnitLeft.Uint64(), Equals, unitAfter.Uint64())
	}
}

// TestValidateUnstake is to test validateUnstake function
func (s UnstakeSuite) TestValidateUnstake(c *C) {
	accountAddr := GetRandomNodeAccount(NodeWhiteListed).NodeAddress
	publicAddress, err := common.NewAddress("bnb1g0xakzh03tpa54khxyvheeu92hwzypkdce77rm")
	if nil != err {
		c.Error("fail to create new BNB Address")
	}
	inputs := []struct {
		name          string
		msg           MsgSetUnStake
		expectedError error
	}{
		{
			name: "empty-public-address",
			msg: MsgSetUnStake{
				PublicAddress:       "",
				WithdrawBasisPoints: sdk.NewUint(10000),
				Asset:               common.BNBAsset,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			expectedError: errors.New("empty public address"),
		},
		{
			name: "empty-withdraw-basis-points",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.ZeroUint(),
				Asset:               common.BNBAsset,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			expectedError: nil,
		},
		{
			name: "empty-request-txhash",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Asset:               common.BNBAsset,
				RequestTxHash:       "",
				Signer:              accountAddr,
			},
			expectedError: errors.New("request tx hash is empty"),
		},
		{
			name: "empty-asset",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Asset:               common.Asset{},
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			expectedError: errors.New("empty asset"),
		},
		{
			name: "invalid-basis-point",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10001),
				Asset:               common.BNBAsset,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			expectedError: errors.New("withdraw basis points 10001 is invalid"),
		},
		{
			name: "invalid-pool-notexist",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Asset:               common.Asset{Chain: "BNB", Ticker: "NOTEXIST", Symbol: "NOTEXIST"},
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			expectedError: errors.New("pool-BNB.NOTEXIST doesn't exist"),
		},
		{
			name: "all-good",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Asset:               common.BNBAsset,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			expectedError: nil,
		},
	}

	for _, item := range inputs {
		ctx, _ := setupKeeperForTest(c)
		ps := mocks.MockPoolStorage{}
		c.Logf("name:%s", item.name)
		err := validateUnstake(ctx, ps, item.msg)
		if item.expectedError != nil {
			c.Assert(err, NotNil)
			c.Assert(err.Error(), Equals, item.expectedError.Error())
			continue
		}
		c.Assert(err, IsNil)
	}
}

func (UnstakeSuite) TestUnstake(c *C) {
	ps := mocks.MockPoolStorage{}
	accountAddr := GetRandomNodeAccount(NodeWhiteListed).NodeAddress
	publicAddress, err := common.NewAddress("bnb1g0xakzh03tpa54khxyvheeu92hwzypkdce77rm")
	if nil != err {
		c.Error("fail to create new BNB Address")
	}
	testCases := []struct {
		name          string
		msg           MsgSetUnStake
		ps            poolStorage
		runeAmount    sdk.Uint
		assetAmount   sdk.Uint
		expectedError error
	}{
		{
			name: "empty-public-address",
			msg: MsgSetUnStake{
				PublicAddress:       "",
				WithdrawBasisPoints: sdk.NewUint(10000),
				Asset:               common.BNBAsset,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: errors.New("empty public address"),
		},
		{
			name: "empty-withdraw-basis-points",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.ZeroUint(),
				Asset:               common.BNBAsset,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: errors.New("nothing to withdraw"),
		},
		{
			name: "empty-request-txhash",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Asset:               common.BNBAsset,
				RequestTxHash:       "",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: errors.New("request tx hash is empty"),
		},
		{
			name: "empty-asset",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Asset:               common.Asset{},
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: errors.New("empty asset"),
		},

		{
			name: "invalid-basis-point",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10001),
				Asset:               common.BNBAsset,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: errors.New("withdraw basis points 10001 is invalid"),
		},
		{
			name: "invalid-pool-notexist",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Asset:               common.Asset{Chain: "BNB", Ticker: "NOTEXIST", Symbol: "NOTEXIST"},
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: errors.New("pool-BNB.NOTEXIST doesn't exist"),
		},
		{
			name: "invalid-pool-staker-notexist",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Asset:               common.Asset{Chain: "BNB", Ticker: "NOTEXISTSTICKER", Symbol: "NOTEXISTSTICKER"},
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: errors.New("can't find pool staker: you asked for it"),
		},
		{
			name: "invalid-staker-pool-notexist",
			msg: MsgSetUnStake{
				PublicAddress:       common.Address("NOTEXISTSTAKER"),
				WithdrawBasisPoints: sdk.NewUint(10000),
				Asset:               common.BNBAsset,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: errors.New("can't find staker pool: you asked for it"),
		},
		{
			name: "nothing-to-withdraw",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Asset:               common.BNBAsset,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: errors.New("nothing to withdraw"),
		},
		{
			name: "all-good",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Asset:               common.BNBAsset,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            getInMemoryPoolStorageForUnstake(c),
			runeAmount:    sdk.NewUint(100 * common.One),
			assetAmount:   sdk.NewUint(100 * common.One),
			expectedError: nil,
		},
		{
			name: "all-good-half",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(5000),
				Asset:               common.BNBAsset,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            getInMemoryPoolStorageForUnstake(c),
			runeAmount:    sdk.NewUint(50 * common.One),
			assetAmount:   sdk.NewUint(50 * common.One),
			expectedError: nil,
		},
	}
	for _, tc := range testCases {
		ctx, _ := setupKeeperForTest(c)
		c.Logf("name:%s", tc.name)
		r, asset, _, err := unstake(ctx, tc.ps, tc.msg)
		if tc.expectedError != nil {
			c.Assert(err, NotNil)
			c.Check(err.Error(), Equals, tc.expectedError.Error())
			c.Check(r.Uint64(), Equals, tc.runeAmount.Uint64())
			c.Check(asset.Uint64(), Equals, tc.assetAmount.Uint64())
			continue
		}
		c.Assert(err, IsNil)
		c.Check(r.Uint64(), Equals, tc.runeAmount.Uint64())
		c.Check(asset.Uint64(), Equals, tc.assetAmount.Uint64())
	}
}

func getInMemoryPoolStorageForUnstake(c *C) poolStorage {
	publicAddress, err := common.NewAddress("bnb1g0xakzh03tpa54khxyvheeu92hwzypkdce77rm")
	if nil != err {
		c.Error("fail to create new BNB Address")
	}

	ctx, _ := setupKeeperForTest(c)

	store := NewMockInMemoryPoolStorage()
	pool := Pool{
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
		Asset:        common.BNBAsset,
		PoolUnits:    sdk.NewUint(100 * common.One),
		PoolAddress:  publicAddress,
		Status:       PoolEnabled,
	}
	store.SetPool(ctx, pool)
	poolStaker := PoolStaker{
		Asset:      common.BNBAsset,
		TotalUnits: sdk.NewUint(100 * common.One),
		Stakers: []StakerUnit{
			StakerUnit{
				StakerID: publicAddress,
				Units:    sdk.NewUint(100 * common.One),
			},
		},
	}
	store.SetPoolStaker(ctx, common.BNBAsset, poolStaker)
	stakerPool := StakerPool{
		StakerID: publicAddress,
		PoolUnits: []*StakerPoolItem{
			&StakerPoolItem{
				Asset: common.BNBAsset,
				Units: sdk.NewUint(100 * common.One),
				StakeDetails: []StakeTxDetail{
					StakeTxDetail{
						RequestTxHash: common.TxID("28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"),
						RuneAmount:    sdk.NewUint(100 * common.One),
						AssetAmount:   sdk.NewUint(100 * common.One),
					},
				},
			},
		},
	}
	store.SetStakerPool(ctx, publicAddress, stakerPool)
	return store
}
