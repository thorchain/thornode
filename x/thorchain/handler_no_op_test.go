package thorchain

import (
	"github.com/blang/semver"
	. "gopkg.in/check.v1"
)

type HandlerNoOpSuite struct{}

var _ = Suite(&HandlerNoOpSuite{})

type TestNoOpKeeper struct {
	Keeper
}

func newTestNoOpKeeper(keeper Keeper) TestNoOpKeeper {
	return TestNoOpKeeper{
		Keeper: keeper,
	}
}

func (s *HandlerNoOpSuite) TestValidate(c *C) {
	ctx, k := setupKeeperForTest(c)
	keeper := newTestNoOpKeeper(k)
	handler := NewNoOpHandler(keeper)

	// happy path
	ver := semver.MustParse("0.1.0")
	signer := GetRandomBech32Addr()
	msg := NewMsgNoOp(GetRandomObservedTx(), signer)
	err := handler.Validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.Validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, errInvalidVersion)

	// invalid msg
	msg = MsgNoOp{}
	err = handler.Validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

func (s *HandlerNoOpSuite) TestHandle(c *C) {
	ver := semver.MustParse("0.1.0")
	ctx, k := setupKeeperForTest(c)
	keeper := newTestNoOpKeeper(k)
	handler := NewNoOpHandler(keeper)

	signer := GetRandomBech32Addr()
	msg := NewMsgNoOp(GetRandomObservedTx(), signer)

	err := handler.Handle(ctx, msg, ver)
	c.Assert(err, IsNil)
}
