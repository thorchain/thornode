package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type HandlerVersionSuite struct{}

type TestVersionlKeeper struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestVersionlKeeper) GetNodeAccount(_ sdk.Context, _ sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

func (k *TestVersionlKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	k.na = na
	return nil
}

var _ = Suite(&HandlerVersionSuite{})

func (s *HandlerVersionSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestVersionlKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewVersionHandler(keeper)
	// happy path
	ver := semver.MustParse("0.1.0")
	msg := NewMsgSetVersion(ver, keeper.na.NodeAddress)
	err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, errBadVersion)

	// invalid msg
	msg = MsgSetVersion{}
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

func (s *HandlerVersionSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)
	ver := semver.MustParse("0.1.0")

	keeper := &TestVersionlKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewVersionHandler(keeper)

	msg := NewMsgSetVersion(semver.MustParse("2.0.0"), GetRandomBech32Addr())
	err := handler.handle(ctx, msg, ver)
	c.Assert(err, IsNil)
	c.Check(keeper.na.Version.String(), Equals, "2.0.0")
}
