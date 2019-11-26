package thorchain

import (
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type HandlerPoolDataSuite struct {
	w       handlerTestWrapper
	handler PoolDataHandler
}

var _ = Suite(&HandlerPoolDataSuite{})

func (s *HandlerPoolDataSuite) SetupSuite(c *C) {
	s.w = getHandlerTestWrapper(c, 1, true, false)
	s.handler = NewPoolDataHandler(s.w.keeper)
}

func (s *HandlerPoolDataSuite) TestValidate(c *C) {
	// happy path
	msg := NewMsgSetPoolData(common.BNBAsset, PoolEnabled, s.w.activeNodeAccount.NodeAddress)
	err := s.handler.Validate(s.w.ctx, msg, 1)
	c.Assert(err, IsNil)

	// inactive node account
	msg = NewMsgSetPoolData(common.BNBAsset, PoolEnabled, s.w.notActiveNodeAccount.NodeAddress)
	err = s.handler.Validate(s.w.ctx, msg, 1)
	c.Assert(err, Equals, notAuthorized)

	// invalid msg
	msg = MsgSetPoolData{}
	err = s.handler.Validate(s.w.ctx, msg, 1)
	c.Assert(err, NotNil)
}

func (s *HandlerPoolDataSuite) TestHandle(c *C) {
	msg := NewMsgSetPoolData(common.BNBAsset, PoolEnabled, s.w.activeNodeAccount.NodeAddress)
	err := s.handler.Handle(s.w.ctx, msg, 1)
	c.Assert(err, IsNil)
}
