package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/mocks"
)

type UnstakeSuite struct{}

var _ = Suite(&UnstakeSuite{})

func (s *UnstakeSuite) SetUpSuite(c *C) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("rune", "runepub")
}

func (s UnstakeSuite) TestCalculateUnsake(c *C) {
	inputs := []struct {
		name                  string
		poolUnit              sdk.Uint
		poolRune              sdk.Uint
		poolToken             sdk.Uint
		stakerUnit            sdk.Uint
		percentage            sdk.Uint
		expectedWithdrawRune  sdk.Uint
		expectedWithdrawToken sdk.Uint
		expectedUnitLeft      sdk.Uint
		expectedErr           error
	}{
		{
			name:                  "zero-poolunit",
			poolUnit:              sdk.ZeroUint(),
			poolRune:              sdk.ZeroUint(),
			poolToken:             sdk.ZeroUint(),
			stakerUnit:            sdk.ZeroUint(),
			percentage:            sdk.ZeroUint(),
			expectedWithdrawRune:  sdk.ZeroUint(),
			expectedWithdrawToken: sdk.ZeroUint(),
			expectedUnitLeft:      sdk.ZeroUint(),
			expectedErr:           errors.New("poolUnits can't be zero"),
		},

		{
			name:                  "zero-poolrune",
			poolUnit:              sdk.NewUint(500 * One),
			poolRune:              sdk.ZeroUint(),
			poolToken:             sdk.ZeroUint(),
			stakerUnit:            sdk.ZeroUint(),
			percentage:            sdk.ZeroUint(),
			expectedWithdrawRune:  sdk.ZeroUint(),
			expectedWithdrawToken: sdk.ZeroUint(),
			expectedUnitLeft:      sdk.ZeroUint(),
			expectedErr:           errors.New("pool rune balance can't be zero"),
		},

		{
			name:                  "zero-pooltoken",
			poolUnit:              sdk.NewUint(500 * One),
			poolRune:              sdk.NewUint(500 * One),
			poolToken:             sdk.ZeroUint(),
			stakerUnit:            sdk.ZeroUint(),
			percentage:            sdk.ZeroUint(),
			expectedWithdrawRune:  sdk.ZeroUint(),
			expectedWithdrawToken: sdk.ZeroUint(),
			expectedUnitLeft:      sdk.ZeroUint(),
			expectedErr:           errors.New("pool token balance can't be zero"),
		},
		{
			name:                  "negative-stakerUnit",
			poolUnit:              sdk.NewUint(500 * One),
			poolRune:              sdk.NewUint(500 * One),
			poolToken:             sdk.NewUint(5100 * One),
			stakerUnit:            sdk.ZeroUint(),
			percentage:            sdk.ZeroUint(),
			expectedWithdrawRune:  sdk.ZeroUint(),
			expectedWithdrawToken: sdk.ZeroUint(),
			expectedUnitLeft:      sdk.ZeroUint(),
			expectedErr:           errors.New("staker unit can't be zero"),
		},

		{
			name:                  "percentage-larger-than-100",
			poolUnit:              sdk.NewUint(500 * One),
			poolRune:              sdk.NewUint(500 * One),
			poolToken:             sdk.NewUint(500 * One),
			stakerUnit:            sdk.NewUint(100 * One),
			percentage:            sdk.NewUint(12000),
			expectedWithdrawRune:  sdk.ZeroUint(),
			expectedWithdrawToken: sdk.ZeroUint(),
			expectedUnitLeft:      sdk.ZeroUint(),
			expectedErr:           errors.Errorf("withdraw basis point %s is not valid", sdk.NewUint(12000)),
		},
		{
			name:                  "unstake-1",
			poolUnit:              sdk.NewUint(700 * One),
			poolRune:              sdk.NewUint(700 * One),
			poolToken:             sdk.NewUint(700 * One),
			stakerUnit:            sdk.NewUint(200 * One),
			percentage:            sdk.NewUint(10000),
			expectedUnitLeft:      sdk.ZeroUint(),
			expectedWithdrawToken: sdk.NewUint(200 * One),
			expectedWithdrawRune:  sdk.NewUint(200 * One),
			expectedErr:           nil,
		},
	}

	for _, item := range inputs {
		c.Logf("name:%s", item.name)
		withDrawRune, withDrawToken, unitAfter, err := calculateUnstake(item.poolUnit, item.poolRune, item.poolToken, item.stakerUnit, item.percentage)
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}
		c.Logf("expected rune:%s,rune:%s", item.expectedWithdrawRune, withDrawRune)
		c.Check(item.expectedWithdrawRune.Uint64(), Equals, withDrawRune.Uint64())
		c.Check(item.expectedWithdrawToken.Uint64(), Equals, withDrawToken.Uint64())
		c.Check(item.expectedUnitLeft.Uint64(), Equals, unitAfter.Uint64())
	}
}

// TestValidateUnstake is to test validateUnstake function
func (s UnstakeSuite) TestValidateUnstake(c *C) {
	accountAddr, err := sdk.AccAddressFromBech32("rune1375qq0afqr5a6xmh0xspk2jh4wqnmm4024vm6j")
	if nil != err {
		c.Errorf("fail to create account address error:%s", err)
	}
	publicAddress, err := common.NewBnbAddress("tbnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
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
				Ticker:              common.BNBTicker,
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
				Ticker:              common.BNBTicker,
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
				Ticker:              common.BNBTicker,
				RequestTxHash:       "",
				Signer:              accountAddr,
			},
			expectedError: errors.New("request tx hash is empty"),
		},
		{
			name: "empty-ticker",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Ticker:              "",
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			expectedError: errors.New("empty ticker"),
		},
		{
			name: "invalid-basis-point",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10001),
				Ticker:              common.BNBTicker,
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
				Ticker:              common.Ticker("NOTEXIST"),
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			expectedError: errors.New("pool-NOTEXIST doesn't exist"),
		},
		{
			name: "all-good",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			expectedError: nil,
		},
	}

	for _, item := range inputs {
		ctx := GetCtx("test")
		ps := mocks.MockPoolStorage{}
		c.Logf("name:%s", item.name)
		err := validateUnstake(ctx, ps, item.msg)
		if item.expectedError != nil {
			c.Assert(err, NotNil)
			c.Check(err.Error(), Equals, item.expectedError.Error())
			continue
		}
		c.Assert(err, IsNil)

	}
}
func (UnstakeSuite) TestUnstake(c *C) {
	ps := mocks.MockPoolStorage{}
	accountAddr, err := sdk.AccAddressFromBech32("rune1375qq0afqr5a6xmh0xspk2jh4wqnmm4024vm6j")
	if nil != err {
		c.Errorf("fail to create account address error:%s", err)
	}
	publicAddress, err := common.NewBnbAddress("tbnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	if nil != err {
		c.Error("fail to create new BNB Address")
	}
	testCases := []struct {
		name          string
		msg           MsgSetUnStake
		ps            poolStorage
		runeAmount    sdk.Uint
		tokenAmount   sdk.Uint
		expectedError error
	}{
		{
			name: "empty-public-address",
			msg: MsgSetUnStake{
				PublicAddress:       "",
				WithdrawBasisPoints: sdk.NewUint(10000),
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			tokenAmount:   sdk.ZeroUint(),
			expectedError: errors.New("empty public address"),
		},
		{
			name: "empty-withdraw-basis-points",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.ZeroUint(),
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			tokenAmount:   sdk.ZeroUint(),
			expectedError: errors.New("nothing to withdraw"),
		},
		{
			name: "empty-request-txhash",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Ticker:              common.BNBTicker,
				RequestTxHash:       "",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			tokenAmount:   sdk.ZeroUint(),
			expectedError: errors.New("request tx hash is empty"),
		},
		{
			name: "empty-ticker",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Ticker:              "",
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			tokenAmount:   sdk.ZeroUint(),
			expectedError: errors.New("empty ticker"),
		},

		{
			name: "invalid-basis-point",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10001),
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			tokenAmount:   sdk.ZeroUint(),
			expectedError: errors.New("withdraw basis points 10001 is invalid"),
		},
		{
			name: "invalid-pool-notexist",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Ticker:              common.Ticker("NOTEXIST"),
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			tokenAmount:   sdk.ZeroUint(),
			expectedError: errors.New("pool-NOTEXIST doesn't exist"),
		},
		{
			name: "invalid-pool-staker-notexist",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Ticker:              common.Ticker("NOTEXISTSTICKER"),
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			tokenAmount:   sdk.ZeroUint(),
			expectedError: errors.New("can't find pool staker: you asked for it"),
		},
		{
			name: "invalid-staker-pool-notexist",
			msg: MsgSetUnStake{
				PublicAddress:       common.BnbAddress("NOTEXISTSTAKER"),
				WithdrawBasisPoints: sdk.NewUint(10000),
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			tokenAmount:   sdk.ZeroUint(),
			expectedError: errors.New("can't find staker pool: you asked for it"),
		},
		{
			name: "nothing-to-withdraw",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			tokenAmount:   sdk.ZeroUint(),
			expectedError: errors.New("nothing to withdraw"),
		},
		{
			name: "all-good",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(10000),
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            getInMemoryPoolStorageForUnstake(c),
			runeAmount:    sdk.NewUint(100 * One),
			tokenAmount:   sdk.NewUint(100 * One),
			expectedError: nil,
		},
		{
			name: "all-good-half",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: sdk.NewUint(5000),
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            getInMemoryPoolStorageForUnstake(c),
			runeAmount:    sdk.NewUint(50 * One),
			tokenAmount:   sdk.NewUint(50 * One),
			expectedError: nil,
		},
	}
	for _, tc := range testCases {
		ctx := GetCtx("test")
		c.Logf("name:%s", tc.name)
		rune, token, _, err := unstake(ctx, tc.ps, tc.msg)
		if tc.expectedError != nil {
			c.Assert(err, NotNil)
			c.Check(err.Error(), Equals, tc.expectedError.Error())
			c.Check(rune.Uint64(), Equals, tc.runeAmount.Uint64())
			c.Check(token.Uint64(), Equals, tc.tokenAmount.Uint64())
			continue
		}
		c.Assert(err, IsNil)
		c.Check(rune.Uint64(), Equals, tc.runeAmount.Uint64())
		c.Check(token.Uint64(), Equals, tc.tokenAmount.Uint64())
	}
}

func getInMemoryPoolStorageForUnstake(c *C) poolStorage {
	publicAddress, err := common.NewBnbAddress("tbnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	if nil != err {
		c.Error("fail to create new BNB Address")
	}

	ctx := GetCtx("test")

	store := NewMockInMemoryPoolStorage()
	pool := Pool{
		BalanceRune:  sdk.NewUint(100 * One),
		BalanceToken: sdk.NewUint(100 * One),
		Ticker:       common.BNBTicker,
		PoolUnits:    sdk.NewUint(100 * One),
		PoolAddress:  publicAddress,
		Status:       PoolEnabled,
	}
	store.SetPool(ctx, pool)
	poolStaker := PoolStaker{
		Ticker:     common.BNBTicker,
		TotalUnits: sdk.NewUint(100 * One),
		Stakers: []StakerUnit{
			StakerUnit{
				StakerID: publicAddress,
				Units:    sdk.NewUint(100 * One),
			},
		},
	}
	store.SetPoolStaker(ctx, common.BNBTicker, poolStaker)
	stakerPool := StakerPool{
		StakerID: publicAddress,
		PoolUnits: []*StakerPoolItem{
			&StakerPoolItem{
				Ticker: common.BNBTicker,
				Units:  sdk.NewUint(100 * One),
				StakeDetails: []StakeTxDetail{
					StakeTxDetail{
						RequestTxHash: common.TxID("28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"),
						RuneAmount:    sdk.NewUint(100 * One),
						TokenAmount:   sdk.NewUint(100 * One),
					},
				},
			},
		},
	}
	store.SetStakerPool(ctx, publicAddress, stakerPool)
	return store
}
