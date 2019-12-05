package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type HandlerObservedTxInSuite struct{}

type TestObservedTxInValidateKeeper struct {
	KVStoreDummy
	isActive bool
}

func (k *TestObservedTxInValidateKeeper) IsActiveObserver(ctx sdk.Context, signer sdk.AccAddress) bool {
	return k.isActive
}

var _ = Suite(&HandlerObservedTxInSuite{})

func (s *HandlerObservedTxInSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	keeper := &TestObservedTxInValidateKeeper{
		isActive: true,
	}

	handler := NewObservedTxInHandler(keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr)

	// happy path
	ver := semver.MustParse("0.1.0")
	txs := ObservedTxs{NewObservedTx(GetRandomTx(), sdk.NewUint(12), GetRandomPubKey())}
	msg := NewMsgObservedTxIn(txs, GetRandomBech32Addr())
	err := handler.Validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.Validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// inactive node account
	keeper.isActive = false
	msg = NewMsgObservedTxIn(txs, GetRandomBech32Addr())
	err = handler.Validate(ctx, msg, ver)
	c.Assert(err, Equals, notAuthorized)

	// invalid msg
	msg = MsgObservedTxIn{}
	err = handler.Validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

type TestObservedTxInHandleKeeper struct {
	KVStoreDummy
}

func (s *HandlerObservedTxInSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	ver := semver.MustParse("0.1.0")

	keeper := &TestObservedTxInHandleKeeper{}

	handler := NewObservedTxInHandler(keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr)

	txs := ObservedTxs{NewObservedTx(GetRandomTx(), sdk.NewUint(12), GetRandomPubKey())}
	msg := NewMsgObservedTxIn(txs, GetRandomBech32Addr())
	err := handler.Handle(ctx, msg, ver)
	c.Assert(err, IsNil)
}
