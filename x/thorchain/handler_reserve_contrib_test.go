package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type HandlerReserveContributorSuite struct{}

type TestReserveContributorerKeeper struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestReserveContributorerKeeper) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

var _ = Suite(&HandlerReserveContributorSuite{})

func (s *HandlerReserveContributorSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestReserveContributorerKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewReserveContributorHandler(keeper)
	// happy path
	ver := semver.MustParse("0.1.0")
	res := NewReserveContributor(GetRandomBNBAddress(), sdk.NewUint(34))
	msg := NewMsgReserveContributor(res, keeper.na.NodeAddress)
	err := handler.Validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.Validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// inactive node account
	keeper.na = GetRandomNodeAccount(NodeStandby)
	msg = NewMsgReserveContributor(res, keeper.na.NodeAddress)
	err = handler.Validate(ctx, msg, ver)
	c.Assert(err, Equals, notAuthorized)

	// invalid msg
	msg = MsgReserveContributor{}
	err = handler.Validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

type TestReserveContributorKeeper struct {
	KVStoreDummy
	vault    VaultData
	contribs ReserveContributors
}

func (s *TestReserveContributorKeeper) GetVaultData(_ sdk.Context) (VaultData, error) {
	return s.vault, nil
}

func (s *TestReserveContributorKeeper) SetVaultData(_ sdk.Context, data VaultData) error {
	s.vault = data
	return nil
}

func (s *TestReserveContributorKeeper) GetReservesContributors(_ sdk.Context) (ReserveContributors, error) {
	return s.contribs, nil
}

func (s *TestReserveContributorKeeper) SetReserveContributors(_ sdk.Context, contribs ReserveContributors) error {
	s.contribs = contribs
	return nil
}

func (s *HandlerReserveContributorSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)
	ver := semver.MustParse("0.1.0")

	keeper := &TestReserveContributorKeeper{
		vault: NewVaultData(),
	}

	handler := NewReserveContributorHandler(keeper)

	addr := GetRandomBNBAddress()
	res := NewReserveContributor(addr, sdk.NewUint(23*common.One))
	msg := NewMsgReserveContributor(res, GetRandomBech32Addr())

	err := handler.Handle(ctx, msg, ver)
	c.Assert(err, IsNil)
	c.Check(keeper.vault.TotalReserve.Equal(sdk.NewUint(23*common.One)), Equals, true)
	c.Assert(keeper.contribs, HasLen, 1)
	c.Assert(keeper.contribs[0].Amount.Equal(sdk.NewUint(23*common.One)), Equals, true)
	c.Assert(keeper.contribs[0].Address.Equals(addr), Equals, true)
}
