package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type HandlerLeaveSuite struct{}

var _ = Suite(&HandlerLeaveSuite{})

func (HandlerLeaveSuite) TestLeaveHandler_NotActiveNodeLeave(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	vault := GetRandomVault()
	w.keeper.SetVault(w.ctx, vault)
	leaveHandler := NewLeaveHandler(w.keeper, w.validatorMgr, w.versionedTxOutStore, NewVersionedEventMgr())
	acc2 := GetRandomNodeAccount(NodeStandby)
	acc2.Bond = sdk.NewUint(100 * common.One)
	c.Assert(w.keeper.SetNodeAccount(w.ctx, acc2), IsNil)
	ygg := NewVault(w.ctx.BlockHeight(), ActiveVault, YggdrasilVault, acc2.PubKeySet.Secp256k1, common.Chains{common.RuneAsset().Chain})
	c.Assert(w.keeper.SetVault(w.ctx, ygg), IsNil)

	FundModule(c, w.ctx, w.keeper, AsgardName, 100)

	txID := GetRandomTxHash()
	tx := common.NewTx(
		txID,
		acc2.BondAddress,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.RuneAsset(), sdk.OneUint())},
		BNBGasFeeSingleton,
		"LEAVE",
	)
	msgLeave := NewMsgLeave(tx, w.activeNodeAccount.NodeAddress)
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)
	result := leaveHandler.Run(w.ctx, msgLeave, ver, constAccessor)
	c.Assert(result.Code, Equals, sdk.CodeOK, Commentf("%+v", result))
	result1 := leaveHandler.Run(w.ctx, msgLeave, semver.Version{}, constAccessor)
	c.Assert(result1.Code, Equals, CodeBadVersion)
}

func (HandlerLeaveSuite) TestLeaveHandler_ActiveNodeLeave(c *C) {
	var err error
	w := getHandlerTestWrapper(c, 1, true, false)
	leaveHandler := NewLeaveHandler(w.keeper, w.validatorMgr, w.versionedTxOutStore, NewVersionedEventMgr())
	acc2 := GetRandomNodeAccount(NodeActive)
	acc2.Bond = sdk.NewUint(100 * common.One)
	c.Assert(w.keeper.SetNodeAccount(w.ctx, acc2), IsNil)
	txID := GetRandomTxHash()
	tx := common.NewTx(
		txID,
		acc2.BondAddress,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.RuneAsset(), sdk.OneUint())},
		BNBGasFeeSingleton,
		"",
	)
	msgLeave := NewMsgLeave(tx, w.activeNodeAccount.NodeAddress)
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)
	result := leaveHandler.Run(w.ctx, msgLeave, ver, constAccessor)
	c.Assert(result.Code, Equals, sdk.CodeOK)

	acc2, err = w.keeper.GetNodeAccountByBondAddress(w.ctx, acc2.BondAddress)
	c.Assert(err, IsNil)
	c.Check(acc2.Bond.Equal(sdk.NewUint(10000000001)), Equals, true, Commentf("Bond:%d\n", acc2.Bond.Uint64()))
}

func (HandlerLeaveSuite) TestLeaveValidation(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)
	testCases := []struct {
		name         string
		msgLeave     MsgLeave
		expectedCode sdk.CodeType
	}{
		{
			name: "empty from address should fail",
			msgLeave: NewMsgLeave(common.Tx{
				ID:          GetRandomTxHash(),
				Chain:       common.BNBChain,
				FromAddress: "",
				ToAddress:   GetRandomBNBAddress(),
				Coins: common.Coins{
					common.NewCoin(common.BNBAsset, sdk.NewUint(common.One)),
				},
				Gas: common.Gas{
					common.NewCoin(common.BNBAsset, sdk.NewUint(common.One)),
				},
				Memo: "",
			}, w.activeNodeAccount.NodeAddress),
			expectedCode: sdk.CodeUnknownRequest,
		},
		{
			name: "empty tx id should fail",
			msgLeave: NewMsgLeave(common.Tx{
				ID:          common.BlankTxID,
				Chain:       common.BNBChain,
				FromAddress: GetRandomBNBAddress(),
				ToAddress:   GetRandomBNBAddress(),
				Coins: common.Coins{
					common.NewCoin(common.BNBAsset, sdk.NewUint(common.One)),
				},
				Gas: common.Gas{
					common.NewCoin(common.BNBAsset, sdk.NewUint(common.One)),
				},
				Memo: "",
			}, w.activeNodeAccount.NodeAddress),
			expectedCode: sdk.CodeUnknownRequest,
		},
		{
			name: "empty signer should fail",
			msgLeave: NewMsgLeave(common.Tx{
				ID:          GetRandomTxHash(),
				Chain:       common.BNBChain,
				FromAddress: GetRandomBNBAddress(),
				ToAddress:   GetRandomBNBAddress(),
				Coins: common.Coins{
					common.NewCoin(common.BNBAsset, sdk.NewUint(common.One)),
				},
				Gas: common.Gas{
					common.NewCoin(common.BNBAsset, sdk.NewUint(common.One)),
				},
				Memo: "",
			}, sdk.AccAddress{}),
			expectedCode: sdk.CodeUnknownRequest,
		},
		{
			name: "empty signer should fail",
			msgLeave: NewMsgLeave(common.Tx{
				ID:          GetRandomTxHash(),
				Chain:       common.BNBChain,
				FromAddress: GetRandomBNBAddress(),
				ToAddress:   GetRandomBNBAddress(),
				Coins: common.Coins{
					common.NewCoin(common.BNBAsset, sdk.NewUint(common.One)),
				},
				Gas: common.Gas{
					common.NewCoin(common.BNBAsset, sdk.NewUint(common.One)),
				},
				Memo: "",
			}, GetRandomNodeAccount(NodeActive).NodeAddress),
			expectedCode: sdk.CodeUnauthorized,
		},
	}
	for _, item := range testCases {
		c.Log(item.name)
		leaveHandler := NewLeaveHandler(w.keeper, w.validatorMgr, w.versionedTxOutStore, NewVersionedEventMgr())
		c.Assert(leaveHandler.Run(w.ctx, item.msgLeave, ver, constAccessor).Code, Equals, item.expectedCode)
	}
}
