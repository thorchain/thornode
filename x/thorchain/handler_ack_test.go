package thorchain

import (
	"errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type HandlerAckTestSuite struct{}

var _ = Suite(&HandlerAckTestSuite{})

type testActHelper struct {
	KVStoreDummy
}

func (t *testActHelper) SetNodeAccount(_ sdk.Context, _ NodeAccount) error {
	return errors.New("you ask for it")
}

func (HandlerAckTestSuite) TestAckHandler(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)

	// happy path
	// first of all , pool rotation window need to open
	blockHeight := w.poolAddrMgr.GetCurrentPoolAddresses().RotateWindowOpenAt
	ctx := w.ctx.WithBlockHeight(w.poolAddrMgr.GetCurrentPoolAddresses().RotateWindowOpenAt)
	c.Assert(w.poolAddrMgr.BeginBlock(ctx), IsNil)
	w.txOutStore.NewBlock(uint64(blockHeight))
	w.poolAddrMgr.EndBlock(ctx, w.txOutStore)

	// we need to observe next pool address
	nextPoolAddrPubKey := GetRandomPoolPubKeys()
	w.poolAddrMgr.SetObservedNextPoolAddrPubKey(nextPoolAddrPubKey)
	sender, err := nextPoolAddrPubKey.GetByChain(common.BNBChain).GetAddress()
	c.Assert(err, IsNil)
	msg := NewMsgAck(GetRandomTx(), sender, common.BNBChain, w.activeNodeAccount.NodeAddress)
	ackHandler := NewAckHandler(w.keeper, w.poolAddrMgr, w.validatorMgr)
	version, err := semver.New("0.1.0")
	c.Assert(err, IsNil)
	result := ackHandler.Run(w.ctx, msg, *version)
	c.Assert(result.Code, Equals, sdk.CodeOK)

	// invalid version
	version1 := semver.Version{}
	c.Assert(err, IsNil)
	c.Assert(ackHandler.Run(w.ctx, msg, version1).Code, Equals, CodeBadVersion)
}

func (HandlerAckTestSuite) TestAckValidateError(c *C) {

	nextPoolPubKey := GetRandomPoolPubKeys()
	sender, err := nextPoolPubKey.GetByChain(common.BNBChain).GetAddress()
	c.Assert(err, IsNil)
	testCases := []struct {
		name           string
		msgAck         MsgAck
		preTest        func(w handlerTestWrapper)
		expectedResult sdk.CodeType
	}{
		{
			name:           "empty sender",
			msgAck:         NewMsgAck(GetRandomTx(), common.NoAddress, common.BNBChain, GetRandomNodeAccount(NodeActive).NodeAddress),
			preTest:        nil,
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name:           "invalid tx",
			msgAck:         NewMsgAck(common.Tx{}, GetRandomBNBAddress(), common.BNBChain, GetRandomNodeAccount(NodeActive).NodeAddress),
			preTest:        nil,
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name:           "empty chain",
			msgAck:         NewMsgAck(GetRandomTx(), GetRandomBNBAddress(), common.EmptyChain, GetRandomNodeAccount(NodeActive).NodeAddress),
			preTest:        nil,
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name:           "none BNB chain",
			msgAck:         NewMsgAck(GetRandomTx(), GetRandomBNBAddress(), common.BTCChain, GetRandomNodeAccount(NodeActive).NodeAddress),
			preTest:        nil,
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name:   "pool rotation window not open",
			msgAck: NewMsgAck(GetRandomTx(), GetRandomBNBAddress(), common.BNBChain, GetRandomNodeAccount(NodeActive).NodeAddress),
			preTest: func(w handlerTestWrapper) {

			},
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name:   "did not observe next pool address pub key yet",
			msgAck: NewMsgAck(GetRandomTx(), sender, common.BNBChain, GetRandomNodeAccount(NodeActive).NodeAddress),
			preTest: func(w handlerTestWrapper) {
				w.poolAddrMgr.SetRotateWindowOpen(true)
			},
			expectedResult: sdk.CodeUnknownRequest,
		},
	}
	for _, item := range testCases {
		w := getHandlerTestWrapper(c, 1, true, false)
		ver := semver.MustParse("0.1.0")
		ackHandler := NewAckHandler(w.keeper, w.poolAddrMgr, w.validatorMgr)
		if item.preTest != nil {
			item.preTest(w)
		}
		result := ackHandler.Run(w.ctx, item.msgAck, ver)
		c.Assert(result.Code, Equals, item.expectedResult)
	}

}

func (HandlerAckTestSuite) TestHandlerDirectly(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	ackHandler := NewAckHandler(w.keeper, w.poolAddrMgr, w.validatorMgr)
	// if THORChain don't have pool for the given chain , it should fail
	msg := NewMsgAck(GetRandomTx(), GetRandomBNBAddress(), common.BTCChain, w.activeNodeAccount.NodeAddress)
	c.Assert(ackHandler.handle(w.ctx, msg).Code(), Equals, sdk.CodeUnknownRequest)

	// sender doesn't match the observed pub key
	w.poolAddrMgr.SetRotateWindowOpen(true)
	w.poolAddrMgr.SetObservedNextPoolAddrPubKey(GetRandomPoolPubKeys())
	msg = NewMsgAck(GetRandomTx(), GetRandomBNBAddress(), common.BTCChain, w.activeNodeAccount.NodeAddress)
	c.Assert(ackHandler.handle(w.ctx, msg).Code(), Equals, sdk.CodeUnknownRequest)

	// if THORChain fail to set node account , then it should fail
	w.poolAddrMgr.SetRotateWindowOpen(true)
	poolPubKey := GetRandomPoolPubKeys()
	w.poolAddrMgr.SetObservedNextPoolAddrPubKey(poolPubKey)
	sender, err := poolPubKey.GetByChain(common.BNBChain).GetAddress()
	c.Assert(err, IsNil)

	msg = NewMsgAck(GetRandomTx(), sender, common.BNBChain, w.activeNodeAccount.NodeAddress)
	ackHandler = NewAckHandler(&testActHelper{}, w.poolAddrMgr, w.validatorMgr)
	c.Assert(ackHandler.handle(w.ctx, msg).Code(), Equals, sdk.CodeInternal)
}
