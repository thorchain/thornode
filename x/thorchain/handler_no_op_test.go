package thorchain

import (
	"github.com/blang/semver"
	. "gopkg.in/check.v1"
)

type HandlerNoOpSuite struct{}

type TestNoOpKeeper struct {
	KVStoreDummy
}

var _ = Suite(&HandlerNoOpSuite{})

func (s *HandlerNoOpSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)
	keeper := &TestNoOpKeeper{}
	handler := NewNoOpHandler(keeper)

	// happy path
	ver := semver.MustParse("0.1.0")
	signer := GetRandomBech32Addr()
	msg := NewMsgNoOp(signer)
	err := handler.Validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.Validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// invalid msg
	msg = MsgNoOp{}
	err = handler.Validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

func (s *HandlerNoOpSuite) TestHandle(c *C) {
	ver := semver.MustParse("0.1.0")
	ctx, _ := setupKeeperForTest(c)
	keeper := &TestNoOpKeeper{}
	handler := NewNoOpHandler(keeper)

	signer := GetRandomBech32Addr()
	msg := NewMsgNoOp(signer)

	err := handler.Handle(ctx, msg, ver)
	c.Assert(err, IsNil)
}
