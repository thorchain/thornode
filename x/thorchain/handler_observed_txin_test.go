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
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	keeper := &TestObservedTxInValidateKeeper{
		isActive: true,
	}

	handler := NewObservedTxInHandler(keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr)

	// happy path
	ver := semver.MustParse("0.1.0")
	pk := GetRandomPubKey()
	txs := ObservedTxs{NewObservedTx(GetRandomTx(), sdk.NewUint(12), pk)}
	txs[0].Tx.ToAddress, err = pk.GetAddress(txs[0].Tx.Coins[0].Asset.Chain)
	c.Assert(err, IsNil)
	msg := NewMsgObservedTxIn(txs, GetRandomBech32Addr())
	err = handler.Validate(ctx, msg, ver)
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
	activeNodeAccounts NodeAccounts
}

func (k TestObservedTxInHandleKeeper) ListActiveNodeAccounts(ctx sdk.Context) (NodeAccounts, error) {
	return k.activeNodeAccounts, nil
}

func (s *HandlerObservedTxInSuite) TestHandle(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	ver := semver.MustParse("0.1.0")

	keeper := &TestObservedTxInHandleKeeper{}

	handler := NewObservedTxInHandler(keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr)

	pk := GetRandomPubKey()
	txs := ObservedTxs{NewObservedTx(GetRandomTx(), sdk.NewUint(12), pk)}
	txs[0].Tx.ToAddress, err = pk.GetAddress(txs[0].Tx.Coins[0].Asset.Chain)
	c.Assert(err, IsNil)
	msg := NewMsgObservedTxIn(txs, GetRandomBech32Addr())
	err = handler.Handle(ctx, msg, ver)
	c.Assert(err, IsNil)
}
