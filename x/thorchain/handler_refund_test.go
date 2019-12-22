package thorchain

import (
	. "gopkg.in/check.v1"
)

type HandlerRefundSuite struct{}

type TestRefundValidKeepr struct {
	KVStoreDummy
	na NodeAccount
}

var _ = Suite(&HandlerRefundSuite{})

func (HandlerRefundSuite) TestRefundValidation(c *C) {

}
