package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"

	. "gopkg.in/check.v1"
)

type HandlerNativeTxSuite struct{}

var _ = Suite(&HandlerNativeTxSuite{})

func (s *HandlerNativeTxSuite) TestValidate(c *C) {
	ctx, k := setupKeeperForTest(c)

	addr := GetRandomBech32Addr()

	coins := common.Coins{
		common.NewCoin(common.RuneNative, sdk.NewUint(200*common.One)),
	}
	msg := NewMsgNativeTx(coins, "STAKE:BNB.BNB", addr)

	versionedTxOutStore := NewVersionedTxOutStoreDummy()
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStore)
	versionedGasMgr := NewDummyVersionedGasMgr()
	versionedObMgr := NewDummyVersionedObserverMgr()
	versionedEventManagerDummy := NewDummyVersionedEventMgr()
	versionedValidatorMgr := NewVersionedValidatorDummyMgr()

	handler := NewNativeTxHandler(k, versionedObMgr, versionedTxOutStore, versionedValidatorMgr, versionedVaultMgrDummy, versionedGasMgr, versionedEventManagerDummy)
	err := handler.validate(ctx, msg, constants.SWVersion)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, errInvalidVersion)

	// invalid msg
	msg = MsgNativeTx{}
	err = handler.validate(ctx, msg, constants.SWVersion)
	c.Assert(err, NotNil)
}

func (s *HandlerNativeTxSuite) TestHandle(c *C) {
	ctx, k := setupKeeperForTest(c)
	banker := k.CoinKeeper()
	constAccessor := constants.GetConstantValues(constants.SWVersion)

	versionedTxOutStore := NewVersionedTxOutStoreDummy()
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStore)
	versionedGasMgr := NewDummyVersionedGasMgr()
	versionedObMgr := NewDummyVersionedObserverMgr()
	versionedEventManagerDummy := NewDummyVersionedEventMgr()
	versionedValidatorMgr := NewVersionedValidatorDummyMgr()

	handler := NewNativeTxHandler(k, versionedObMgr, versionedTxOutStore, versionedValidatorMgr, versionedVaultMgrDummy, versionedGasMgr, versionedEventManagerDummy)

	addr := GetRandomBech32Addr()

	coins := common.Coins{
		common.NewCoin(common.RuneNative, sdk.NewUint(200*common.One)),
	}

	funds, err := common.NewCoin(common.RuneNative, sdk.NewUint(300*common.One)).Native()
	c.Assert(err, IsNil)
	_, err = banker.AddCoins(ctx, addr, sdk.NewCoins(funds))
	c.Assert(err, IsNil)

	msg := NewMsgNativeTx(coins, "ADD:BNB.BNB", addr)

	result := handler.handle(ctx, msg, constants.SWVersion, constAccessor)
	c.Assert(result.IsOK(), Equals, true, Commentf("%+v", result.Log))
}
