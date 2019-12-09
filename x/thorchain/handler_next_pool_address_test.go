package thorchain

import (
	"errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type HandlerNextPoolAddressSuite struct{}

var _ = Suite(&HandlerNextPoolAddressSuite{})

// TestNextPoolAddressKeeper is a mock keeper structure to facilitate testing
type TestNextPoolAddressKeeper struct {
	KVStoreDummy
	activeNodeAccount  NodeAccount
	failSetNodeAccount bool
}

func (tnpa *TestNextPoolAddressKeeper) SetNodeAccount(_ sdk.Context, _ NodeAccount) error {
	if tnpa.failSetNodeAccount {
		return errors.New("fail to set node account")
	}
	return nil
}

// TestConfirmNextPoolAddr trying to test happy path
func (HandlerNextPoolAddressSuite) TestConfirmNextPoolAddr(c *C) {
	ctx, _ := setupKeeperForTest(c)
	activeNodeAccount := GetRandomNodeAccount(NodeActive)
	k := &TestNextPoolAddressKeeper{
		activeNodeAccount: activeNodeAccount,
	}
	addrs := NewPoolAddresses(nil, GetRandomPoolPubKeys(), nil, 500, 100)
	dummyPoolAddrMgr := &PoolAddressDummyMgr{
		currentPoolAddresses: addrs,
	}

	validatorMgr := &ValidatorDummyMgr{
		meta: &ValidatorMeta{
			Nominated:                     NodeAccounts{GetRandomNodeAccount(NodeStandby)},
			RotateAtBlockHeight:           100,
			RotateWindowOpenAtBlockHeight: 90,
			Queued:                        NodeAccounts{GetRandomNodeAccount(NodeActive)},
			LeaveQueue:                    nil,
			LeaveOpenWindow:               80,
			LeaveProcessAt:                85,
			Ragnarok:                      false,
		},
	}
	handler := NewHandlerNextPoolAddress(k, dummyPoolAddrMgr, validatorMgr, NewTxStoreDummy())
	chainSenderAddr := dummyPoolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(common.BNBChain)
	senderAddr, err := chainSenderAddr.GetAddress()
	c.Assert(err, IsNil)
	msgNextPoolAddr := NewMsgNextPoolAddress(
		GetRandomTx(),
		GetRandomPubKey(),
		senderAddr,
		common.BNBChain,
		activeNodeAccount.NodeAddress)
	ver := semver.MustParse("0.1.0")
	dummyPoolAddrMgr.SetRotateWindowOpen(true)
	result := handler.Run(ctx, msgNextPoolAddr, ver)
	c.Assert(result.Code, Equals, sdk.CodeOK)

	// invalid version should fail

	result = handler.Run(ctx, msgNextPoolAddr, semver.Version{})
	c.Assert(result.Code, Equals, CodeBadVersion)

	// invalid message should fail
	result = handler.Run(ctx, sdk.NewTestMsg(), ver)
	c.Assert(result.Code, Equals, CodeInvalidMessage)
}

func (HandlerNextPoolAddressSuite) TestHandlerMsgConfirmNextPoolAddress_validation(c *C) {
	ctx, _ := setupKeeperForTest(c)
	activeNodeAccount := GetRandomNodeAccount(NodeActive)
	k := &TestNextPoolAddressKeeper{
		activeNodeAccount: activeNodeAccount,
	}
	addrs := NewPoolAddresses(nil, GetRandomPoolPubKeys(), nil, 500, 100)
	dummyPoolAddrMgr := &PoolAddressDummyMgr{
		currentPoolAddresses: addrs,
	}

	validatorMgr := &ValidatorDummyMgr{
		meta: &ValidatorMeta{
			Nominated:                     NodeAccounts{GetRandomNodeAccount(NodeStandby)},
			RotateAtBlockHeight:           100,
			RotateWindowOpenAtBlockHeight: 90,
			Queued:                        NodeAccounts{GetRandomNodeAccount(NodeActive)},
			LeaveQueue:                    nil,
			LeaveOpenWindow:               80,
			LeaveProcessAt:                85,
			Ragnarok:                      false,
		},
	}
	handler := NewHandlerNextPoolAddress(k, dummyPoolAddrMgr, validatorMgr, NewTxStoreDummy())
	ver := semver.MustParse("0.1.0")
	testCases := []struct {
		name           string
		msg            MsgNextPoolAddress
		expectedResult sdk.CodeType
	}{
		{
			name:           "rotation window not open should fail",
			msg:            NewMsgNextPoolAddress(GetRandomTx(), GetRandomPubKey(), GetRandomBNBAddress(), common.BNBChain, GetRandomNodeAccount(NodeActive).NodeAddress),
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name:           "empty next pool pub key should fail",
			msg:            NewMsgNextPoolAddress(GetRandomTx(), common.EmptyPubKey, GetRandomBNBAddress(), common.BNBChain, GetRandomNodeAccount(NodeActive).NodeAddress),
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name:           "empty sender should fail",
			msg:            NewMsgNextPoolAddress(GetRandomTx(), GetRandomPubKey(), common.NoAddress, common.BNBChain, GetRandomNodeAccount(NodeActive).NodeAddress),
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name:           "not BNB chain should fail",
			msg:            NewMsgNextPoolAddress(GetRandomTx(), GetRandomPubKey(), GetRandomBNBAddress(), common.EmptyChain, GetRandomNodeAccount(NodeActive).NodeAddress),
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name: "invalid tx should fail",
			msg: NewMsgNextPoolAddress(common.Tx{
				ID:          "",
				Chain:       "",
				FromAddress: "",
				ToAddress:   "",
				Coins:       nil,
				Gas:         nil,
				Memo:        "",
			}, GetRandomPubKey(), GetRandomBNBAddress(), common.EmptyChain, GetRandomNodeAccount(NodeActive).NodeAddress),
			expectedResult: sdk.CodeUnknownRequest,
		},
	}

	for _, tc := range testCases {
		c.Assert(handler.Run(ctx, tc.msg, ver).Code, Equals, tc.expectedResult, Commentf(tc.name))
	}
}

func (HandlerNextPoolAddressSuite) TestHandlerMsgConfirmNextPoolAddress_PoolAddrValidation(c *C) {
	ctx, _ := setupKeeperForTest(c)
	activeNodeAccount := GetRandomNodeAccount(NodeActive)
	k := &TestNextPoolAddressKeeper{
		activeNodeAccount: activeNodeAccount,
	}
	addrs := NewPoolAddresses(nil, GetRandomPoolPubKeys(), nil, 500, 100)
	dummyPoolAddrMgr := &PoolAddressDummyMgr{
		currentPoolAddresses: addrs,
	}

	validatorMgr := &ValidatorDummyMgr{
		meta: &ValidatorMeta{
			Nominated:                     NodeAccounts{GetRandomNodeAccount(NodeStandby)},
			RotateAtBlockHeight:           100,
			RotateWindowOpenAtBlockHeight: 90,
			Queued:                        NodeAccounts{GetRandomNodeAccount(NodeActive)},
			LeaveQueue:                    nil,
			LeaveOpenWindow:               80,
			LeaveProcessAt:                85,
			Ragnarok:                      false,
		},
	}

	handler := NewHandlerNextPoolAddress(k, NewPoolAddressDummyMgr(), validatorMgr, NewTxStoreDummy())
	handler.poolAddrManager.SetRotateWindowOpen(true)
	ver := semver.MustParse("0.1.0")
	msg := NewMsgNextPoolAddress(GetRandomTx(), GetRandomPubKey(), GetRandomBNBAddress(), common.BNBChain, GetRandomNodeAccount(NodeActive).NodeAddress)
	result := handler.Run(ctx, msg, ver)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)

	// sender is not current pool address should fail
	handler1 := NewHandlerNextPoolAddress(&TestNextPoolAddressKeeper{
		activeNodeAccount:  activeNodeAccount,
		failSetNodeAccount: true,
	}, dummyPoolAddrMgr, validatorMgr, NewTxStoreDummy())
	handler1.poolAddrManager.SetRotateWindowOpen(true)
	result1 := handler1.Run(ctx, msg, ver)
	c.Assert(result1.Code, Equals, sdk.CodeInvalidAddress)

	addr, err := dummyPoolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(common.BNBChain).GetAddress()
	c.Assert(err, IsNil)
	// fail to save node account should fail
	msg1 := NewMsgNextPoolAddress(GetRandomTx(), GetRandomPubKey(), addr, common.BNBChain, GetRandomNodeAccount(NodeActive).NodeAddress)
	handler2 := NewHandlerNextPoolAddress(k, dummyPoolAddrMgr, validatorMgr, NewTxStoreDummy())
	handler2.poolAddrManager.SetRotateWindowOpen(true)
	result2 := handler1.Run(ctx, msg1, ver)
	c.Assert(result2.Code, Equals, sdk.CodeInternal)

}
