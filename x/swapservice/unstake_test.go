package swapservice

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/mocks"
)

type UnstakeSuite struct{}

var _ = Suite(&UnstakeSuite{})

func (s UnstakeSuite) TestCalculateUnsake(c *C) {
	inputs := []struct {
		name                  string
		poolUnit              float64
		poolRune              float64
		poolToken             float64
		stakerUnit            float64
		percentage            float64
		expectedWithdrawRune  float64
		expectedWithdrawToken float64
		expectedUnitLeft      float64
		expectedErr           error
	}{
		{
			name:        "zero-poolunit",
			poolUnit:    0,
			expectedErr: errors.New("poolUnits can't be zero or negative"),
		},
		{
			name:        "negative-poolunit",
			poolUnit:    -100,
			expectedErr: errors.New("poolUnits can't be zero or negative"),
		},
		{
			name:        "zero-poolrune",
			poolUnit:    500,
			expectedErr: errors.New("pool rune balance can't be zero or negative"),
		},
		{
			name:        "negative-poolrune",
			poolUnit:    500,
			poolRune:    -100,
			expectedErr: errors.New("pool rune balance can't be zero or negative"),
		},
		{
			name:        "zero-pooltoken",
			poolUnit:    500,
			poolRune:    500,
			poolToken:   0,
			expectedErr: errors.New("pool token balance can't be zero or negative"),
		},
		{
			name:        "negative-poolrune",
			poolUnit:    500,
			poolRune:    500,
			poolToken:   -100,
			expectedErr: errors.New("pool token balance can't be zero or negative"),
		},
		{
			name:        "negative-stakerUnit",
			poolUnit:    500,
			poolRune:    500,
			poolToken:   5100,
			stakerUnit:  -100,
			expectedErr: errors.New("staker unit can't be negative"),
		},
		{
			name:        "negative-percentage",
			poolUnit:    500,
			poolRune:    500,
			poolToken:   500,
			stakerUnit:  100,
			percentage:  -20,
			expectedErr: errors.Errorf("percentage %f is not valid", -20.0),
		},
		{
			name:        "percentage-larger-than-100",
			poolUnit:    500,
			poolRune:    500,
			poolToken:   500,
			stakerUnit:  100,
			percentage:  120,
			expectedErr: errors.Errorf("percentage %f is not valid", 120.0),
		},
		{
			name:                  "unstake-1",
			poolUnit:              700,
			poolRune:              700,
			poolToken:             700,
			stakerUnit:            200,
			percentage:            100,
			expectedUnitLeft:      0,
			expectedWithdrawToken: 200,
			expectedWithdrawRune:  200,
			expectedErr:           nil,
		},
		// TOOD add more cases in
	}

	for _, item := range inputs {
		withDrawRune, withDrawToken, unitAfter, err := calculateUnstake(item.poolUnit, item.poolRune, item.poolToken, item.stakerUnit, item.percentage)
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}
		c.Check(round(item.expectedWithdrawRune), Equals, withDrawRune)
		c.Check(round(item.expectedWithdrawToken), Equals, withDrawToken)
		c.Check(round(item.expectedUnitLeft), Equals, unitAfter)
	}
}

// TestValidateUnstake is to test validateUnstake function
func (s UnstakeSuite) TestValidateUnstake(c *C) {
	accountAddr, err := types.AccAddressFromBech32("rune1375qq0afqr5a6xmh0xspk2jh4wqnmm4024vm6j")
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
				WithdrawBasisPoints: "100",
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
				WithdrawBasisPoints: "",
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			expectedError: errors.New("empty withdraw basis points"),
		},
		{
			name: "empty-request-txhash",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: "10000",
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
				WithdrawBasisPoints: "10000",
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
				WithdrawBasisPoints: "-100",
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			expectedError: errors.New("withdraw basis points -100 is invalid"),
		},
		{
			name: "invalid-basis-point",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: "10001",
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
				WithdrawBasisPoints: "10000",
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
				WithdrawBasisPoints: "10000",
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
	accountAddr, err := types.AccAddressFromBech32("rune1375qq0afqr5a6xmh0xspk2jh4wqnmm4024vm6j")
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
		runeAmount    common.Amount
		tokenAmount   common.Amount
		expectedError error
	}{
		{
			name: "empty-public-address",
			msg: MsgSetUnStake{
				PublicAddress:       "",
				WithdrawBasisPoints: "100",
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    common.ZeroAmount,
			tokenAmount:   common.ZeroAmount,
			expectedError: errors.New("empty public address"),
		},
		{
			name: "empty-withdraw-basis-points",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: "",
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    common.ZeroAmount,
			tokenAmount:   common.ZeroAmount,
			expectedError: errors.New("empty withdraw basis points"),
		},
		{
			name: "empty-request-txhash",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: "10000",
				Ticker:              common.BNBTicker,
				RequestTxHash:       "",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    common.ZeroAmount,
			tokenAmount:   common.ZeroAmount,
			expectedError: errors.New("request tx hash is empty"),
		},
		{
			name: "empty-ticker",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: "10000",
				Ticker:              "",
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    common.ZeroAmount,
			tokenAmount:   common.ZeroAmount,
			expectedError: errors.New("empty ticker"),
		},
		{
			name: "invalid-basis-point",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: "-100",
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    common.ZeroAmount,
			tokenAmount:   common.ZeroAmount,
			expectedError: errors.New("withdraw basis points -100 is invalid"),
		},
		{
			name: "invalid-basis-point",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: "10001",
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    common.ZeroAmount,
			tokenAmount:   common.ZeroAmount,
			expectedError: errors.New("withdraw basis points 10001 is invalid"),
		},
		{
			name: "invalid-pool-notexist",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: "10000",
				Ticker:              common.Ticker("NOTEXIST"),
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    common.ZeroAmount,
			tokenAmount:   common.ZeroAmount,
			expectedError: errors.New("pool-NOTEXIST doesn't exist"),
		},
		{
			name: "invalid-pool-staker-notexist",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: "10000",
				Ticker:              common.Ticker("NOTEXISTSTICKER"),
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    common.ZeroAmount,
			tokenAmount:   common.ZeroAmount,
			expectedError: errors.New("can't find pool staker: you asked for it"),
		},
		{
			name: "invalid-staker-pool-notexist",
			msg: MsgSetUnStake{
				PublicAddress:       common.BnbAddress("NOTEXISTSTAKER"),
				WithdrawBasisPoints: "10000",
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    common.ZeroAmount,
			tokenAmount:   common.ZeroAmount,
			expectedError: errors.New("can't find staker pool: you asked for it"),
		},
		{
			name: "nothing-to-withdraw",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: "10000",
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            ps,
			runeAmount:    common.ZeroAmount,
			tokenAmount:   common.ZeroAmount,
			expectedError: errors.New("nothing to withdraw"),
		},
		{
			name: "all-good",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: "10000",
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            getInMemoryPoolStorageForUnstake(c),
			runeAmount:    common.NewAmountFromFloat(100),
			tokenAmount:   common.NewAmountFromFloat(100),
			expectedError: nil,
		},
		{
			name: "all-good-half",
			msg: MsgSetUnStake{
				PublicAddress:       publicAddress,
				WithdrawBasisPoints: "5000",
				Ticker:              common.BNBTicker,
				RequestTxHash:       "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE",
				Signer:              accountAddr,
			},
			ps:            getInMemoryPoolStorageForUnstake(c),
			runeAmount:    common.NewAmountFromFloat(50),
			tokenAmount:   common.NewAmountFromFloat(50),
			expectedError: nil,
		},
	}
	for _, tc := range testCases {
		ctx := GetCtx("test")

		rune, token, err := unstake(ctx, tc.ps, tc.msg)
		if tc.expectedError != nil {
			c.Assert(err, NotNil)
			c.Check(err.Error(), Equals, tc.expectedError.Error())
			c.Check(rune, Equals, tc.runeAmount)
			c.Check(token, Equals, tc.tokenAmount)
			continue
		}
		c.Assert(err, IsNil)
		c.Check(rune, Equals, tc.runeAmount)
		c.Check(token, Equals, tc.tokenAmount)
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
		BalanceRune:  common.NewAmountFromFloat(100),
		BalanceToken: common.NewAmountFromFloat(100),
		Ticker:       common.BNBTicker,
		PoolUnits:    common.NewAmountFromFloat(100),
		PoolAddress:  publicAddress,
		Status:       PoolEnabled,
	}
	store.SetPool(ctx, pool)
	poolStaker := PoolStaker{
		Ticker:     common.BNBTicker,
		TotalUnits: common.NewAmountFromFloat(100),
		Stakers: []StakerUnit{
			StakerUnit{
				StakerID: publicAddress,
				Units:    common.NewAmountFromFloat(100),
			},
		},
	}
	store.SetPoolStaker(ctx, common.BNBTicker, poolStaker)
	stakerPool := StakerPool{
		StakerID: publicAddress,
		PoolUnits: []*StakerPoolItem{
			&StakerPoolItem{
				Ticker: common.BNBTicker,
				Units:  common.NewAmountFromFloat(100),
				StakeDetails: []StakeTxDetail{
					StakeTxDetail{
						RequestTxHash: common.TxID("28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"),
						RuneAmount:    common.NewAmountFromFloat(100),
						TokenAmount:   common.NewAmountFromFloat(100),
					},
				},
			},
		},
	}
	store.SetStakerPool(ctx, publicAddress, stakerPool)
	return store
}
