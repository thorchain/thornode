package thorchain

import (
	"github.com/blang/semver"
	"gitlab.com/thorchain/thornode/constants"
	. "gopkg.in/check.v1"
)

type HandlerMimirSuite struct{}

var _ = Suite(&HandlerMimirSuite{})

func (s *HandlerMimirSuite) TestValidate(c *C) {
	ctx, keeper := setupKeeperForTest(c)

	handler := NewMimirHandler(keeper)
	// happy path
	ver := constants.SWVersion
	msg := NewMsgMimir("foo", 44, ADMIN)
	err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, errBadVersion)

	// invalid msg
	msg = MsgMimir{}
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

func (s *HandlerMimirSuite) TestHandle(c *C) {
	ctx, keeper := setupKeeperForTest(c)
	ver := constants.SWVersion

	handler := NewMimirHandler(keeper)

	msg := NewMsgMimir("foo", 55, GetRandomBech32Addr())
	sdkErr := handler.handle(ctx, msg, ver)
	c.Assert(sdkErr, IsNil)
	val, err := keeper.GetMimir(ctx, "foo")
	c.Assert(err, IsNil)
	c.Check(val, Equals, int64(55))
}
