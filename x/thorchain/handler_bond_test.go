package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type HandlerBondSuite struct{}

type TestBondKeeper struct {
	KVStoreDummy
	activeNodeAccount   NodeAccount
	failGetNodeAccount  NodeAccount
	notEmptyNodeAccount NodeAccount
}

func (k *TestBondKeeper) GetNodeAccount(_ sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	if k.activeNodeAccount.NodeAddress.Equals(addr) {
		return k.activeNodeAccount, nil
	}
	if k.failGetNodeAccount.NodeAddress.Equals(addr) {
		return NodeAccount{}, fmt.Errorf("you asked for this error")
	}
	if k.notEmptyNodeAccount.NodeAddress.Equals(addr) {
		return k.notEmptyNodeAccount, nil
	}
	return NodeAccount{}, nil
}

var _ = Suite(&HandlerBondSuite{})

func (HandlerBondSuite) TestBondHandler_Run(c *C) {
	ctx, k1 := setupKeeperForTest(c)
	activeNodeAccount := GetRandomNodeAccount(NodeActive)
	k := &TestBondKeeper{
		activeNodeAccount:   activeNodeAccount,
		failGetNodeAccount:  GetRandomNodeAccount(NodeActive),
		notEmptyNodeAccount: GetRandomNodeAccount(NodeActive),
	}
	// happy path
	c.Assert(k1.SetNodeAccount(ctx, activeNodeAccount), IsNil)
	handler := NewBondHandler(k1, NewVersionedEventMgr())
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)
	minimumBondInRune := constAccessor.GetInt64Value(constants.MinimumBondInRune)
	txIn := common.NewTx(
		GetRandomTxHash(),
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.NewUint(uint64(minimumBondInRune))),
		},
		common.Gas{},
		"apply",
	)
	msg := NewMsgBond(txIn, GetRandomNodeAccount(NodeStandby).NodeAddress, sdk.NewUint(uint64(minimumBondInRune)), GetRandomBNBAddress(), activeNodeAccount.NodeAddress)
	result := handler.Run(ctx, msg, ver, constAccessor)
	c.Assert(result.IsOK(), Equals, true)

	// invalid version
	handler = NewBondHandler(k, NewVersionedEventMgr())
	ver = semver.Version{}
	result = handler.Run(ctx, msg, ver, constAccessor)
	c.Assert(result.Code, Equals, CodeBadVersion)

	// simulate fail to get node account
	ver = constants.SWVersion
	msg = NewMsgBond(txIn, k.failGetNodeAccount.NodeAddress, sdk.NewUint(uint64(minimumBondInRune)), GetRandomBNBAddress(), activeNodeAccount.NodeAddress)
	result = handler.Run(ctx, msg, ver, constAccessor)
	c.Assert(result.Code, Equals, sdk.CodeInternal)

	msg = NewMsgBond(txIn, k.notEmptyNodeAccount.NodeAddress, sdk.NewUint(uint64(minimumBondInRune)), GetRandomBNBAddress(), activeNodeAccount.NodeAddress)
	result = handler.Run(ctx, msg, ver, constAccessor)
	c.Assert(result.Code, Equals, sdk.CodeInternal)
}

func (HandlerBondSuite) TestBondHandlerFailValidation(c *C) {
	ctx, k := setupKeeperForTest(c)
	activeNodeAccount := GetRandomNodeAccount(NodeActive)
	c.Assert(k.SetNodeAccount(ctx, activeNodeAccount), IsNil)
	handler := NewBondHandler(k, NewVersionedEventMgr())
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)
	minimumBondInRune := constAccessor.GetInt64Value(constants.MinimumBondInRune)
	txIn := common.NewTx(
		GetRandomTxHash(),
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.NewUint(uint64(minimumBondInRune))),
		},
		common.Gas{},
		"apply",
	)
	txInNoTxID := txIn
	txInNoTxID.ID = ""
	testCases := []struct {
		name         string
		msg          MsgBond
		expectedCode sdk.CodeType
	}{
		{
			name:         "empty node address",
			msg:          NewMsgBond(txIn, sdk.AccAddress{}, sdk.NewUint(uint64(minimumBondInRune)), GetRandomBNBAddress(), activeNodeAccount.NodeAddress),
			expectedCode: sdk.CodeUnknownRequest,
		},
		{
			name:         "zero bond",
			msg:          NewMsgBond(txIn, GetRandomNodeAccount(NodeStandby).NodeAddress, sdk.ZeroUint(), GetRandomBNBAddress(), activeNodeAccount.NodeAddress),
			expectedCode: sdk.CodeUnknownRequest,
		},
		{
			name:         "empty bond address",
			msg:          NewMsgBond(txIn, GetRandomNodeAccount(NodeStandby).NodeAddress, sdk.NewUint(uint64(minimumBondInRune)), common.Address(""), activeNodeAccount.NodeAddress),
			expectedCode: sdk.CodeUnknownRequest,
		},
		{
			name:         "empty request hash",
			msg:          NewMsgBond(txInNoTxID, GetRandomNodeAccount(NodeStandby).NodeAddress, sdk.NewUint(uint64(minimumBondInRune)), GetRandomBNBAddress(), activeNodeAccount.NodeAddress),
			expectedCode: sdk.CodeUnknownRequest,
		},
		{
			name:         "empty signer",
			msg:          NewMsgBond(txIn, GetRandomNodeAccount(NodeStandby).NodeAddress, sdk.NewUint(uint64(minimumBondInRune)), GetRandomBNBAddress(), sdk.AccAddress{}),
			expectedCode: sdk.CodeInvalidAddress,
		},
		{
			name:         "msg not signed by active account",
			msg:          NewMsgBond(txIn, GetRandomNodeAccount(NodeStandby).NodeAddress, sdk.NewUint(uint64(minimumBondInRune)), GetRandomBNBAddress(), GetRandomNodeAccount(NodeStandby).NodeAddress),
			expectedCode: sdk.CodeUnauthorized,
		},
		{
			name:         "not enough rune",
			msg:          NewMsgBond(txIn, GetRandomNodeAccount(NodeStandby).NodeAddress, sdk.NewUint(uint64(minimumBondInRune-100)), GetRandomBNBAddress(), activeNodeAccount.NodeAddress),
			expectedCode: sdk.CodeUnknownRequest,
		},
	}
	for _, item := range testCases {
		c.Log(item.name)
		result := handler.Run(ctx, item.msg, ver, constAccessor)
		c.Assert(result.Code, Equals, item.expectedCode)
	}
}
