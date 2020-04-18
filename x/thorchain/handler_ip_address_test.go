package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/constants"
	. "gopkg.in/check.v1"
)

type HandlerIPAddressSuite struct{}

type TestIPAddresslKeeper struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestIPAddresslKeeper) GetNodeAccount(_ sdk.Context, _ sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

func (k *TestIPAddresslKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	k.na = na
	return nil
}

var _ = Suite(&HandlerIPAddressSuite{})

func (s *HandlerIPAddressSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestIPAddresslKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewIPAddressHandler(keeper)
	// happy path
	ver := constants.SWVersion
	msg := NewMsgSetIPAddress("8.8.8.8", keeper.na.NodeAddress)
	err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, errBadVersion)

	// invalid msg
	msg = MsgSetIPAddress{}
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

func (s *HandlerIPAddressSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)
	ver := constants.SWVersion

	keeper := &TestIPAddresslKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewIPAddressHandler(keeper)

	msg := NewMsgSetIPAddress("192.168.0.1", GetRandomBech32Addr())
	err := handler.handle(ctx, msg, ver)
	c.Assert(err, IsNil)
	c.Check(keeper.na.IPAddress, Equals, "192.168.0.1")
}
