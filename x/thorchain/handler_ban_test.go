package thorchain

import (
	"errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/constants"
	. "gopkg.in/check.v1"
)

var _ = Suite(&HandlerBanSuite{})

type HandlerBanSuite struct{}

type TestBanKeeper struct {
	KVStoreDummy
	ban       BanVoter
	toBan     NodeAccount
	banner1   NodeAccount
	banner2   NodeAccount
	vaultData VaultData
	err       error
}

func (k *TestBanKeeper) ListActiveNodeAccounts(_ sdk.Context) (NodeAccounts, error) {
	return NodeAccounts{k.toBan, k.banner1, k.banner2}, k.err
}

func (k *TestBanKeeper) GetNodeAccount(_ sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	if addr.Equals(k.toBan.NodeAddress) {
		return k.toBan, k.err
	}
	if addr.Equals(k.banner1.NodeAddress) {
		return k.banner1, k.err
	}
	if addr.Equals(k.banner2.NodeAddress) {
		return k.banner2, k.err
	}
	return NodeAccount{}, errors.New("could not find node account, oops")
}

func (k *TestBanKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	if na.NodeAddress.Equals(k.toBan.NodeAddress) {
		k.toBan = na
		return k.err
	}
	if na.NodeAddress.Equals(k.banner1.NodeAddress) {
		k.banner1 = na
		return k.err
	}
	if na.NodeAddress.Equals(k.banner2.NodeAddress) {
		k.banner2 = na
		return k.err
	}
	return k.err
}

func (k *TestBanKeeper) GetVaultData(ctx sdk.Context) (VaultData, error) {
	return k.vaultData, nil
}

func (k *TestBanKeeper) SetVaultData(ctx sdk.Context, data VaultData) error {
	k.vaultData = data
	return nil
}

func (k *TestBanKeeper) GetBanVoter(_ sdk.Context, addr sdk.AccAddress) (BanVoter, error) {
	return k.ban, k.err
}

func (k *TestBanKeeper) SetBanVoter(_ sdk.Context, ban BanVoter) {
	k.ban = ban
}

func (s *HandlerBanSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	toBan := GetRandomNodeAccount(NodeActive)
	banner1 := GetRandomNodeAccount(NodeActive)
	banner2 := GetRandomNodeAccount(NodeActive)

	keeper := &TestBanKeeper{
		toBan:   toBan,
		banner1: banner1,
		banner2: banner2,
	}

	handler := NewBanHandler(keeper)
	// happy path
	msg := NewMsgBan(toBan.NodeAddress, banner1.NodeAddress)
	err := handler.validate(ctx, msg, constants.SWVersion)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, errBadVersion)

	// invalid msg
	msg = MsgBan{}
	err = handler.validate(ctx, msg, constants.SWVersion)
	c.Assert(err, NotNil)
}

func (s *HandlerBanSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)
	constAccessor := constants.GetConstantValues(constants.SWVersion)
	minBond := constAccessor.GetInt64Value(constants.MinimumBondInRune)

	toBan := GetRandomNodeAccount(NodeActive)
	toBan.Bond = sdk.NewUint(uint64(minBond))
	banner1 := GetRandomNodeAccount(NodeActive)
	banner1.Bond = sdk.NewUint(uint64(minBond))
	banner2 := GetRandomNodeAccount(NodeActive)
	banner2.Bond = sdk.NewUint(uint64(minBond))

	keeper := &TestBanKeeper{
		ban:       NewBanVoter(toBan.NodeAddress),
		toBan:     toBan,
		banner1:   banner1,
		banner2:   banner2,
		vaultData: NewVaultData(),
	}

	handler := NewBanHandler(keeper)

	// ban with banner 1
	msg := NewMsgBan(toBan.NodeAddress, banner1.NodeAddress)
	result := handler.handle(ctx, msg, constants.SWVersion, constAccessor)
	c.Assert(result.IsOK(), Equals, true, Commentf("%+v", result.Log))
	c.Check(int64(keeper.banner1.Bond.Uint64()), Equals, int64(99900000))
	c.Check(int64(keeper.vaultData.TotalReserve.Uint64()), Equals, int64(100000))
	c.Check(keeper.toBan.ForcedToLeave, Equals, false)
	c.Check(keeper.ban.Signers, HasLen, 1)

	// ensure banner 1 can't ban twice
	result = handler.handle(ctx, msg, constants.SWVersion, constAccessor)
	c.Assert(result.IsOK(), Equals, true, Commentf("%+v", result.Log))
	c.Check(int64(keeper.banner1.Bond.Uint64()), Equals, int64(99900000))
	c.Check(int64(keeper.vaultData.TotalReserve.Uint64()), Equals, int64(100000))
	c.Check(keeper.toBan.ForcedToLeave, Equals, false)
	c.Check(keeper.ban.Signers, HasLen, 1)

	// ban with banner 2, which should actually ban the node account
	msg = NewMsgBan(toBan.NodeAddress, banner2.NodeAddress)
	result = handler.handle(ctx, msg, constants.SWVersion, constAccessor)
	c.Assert(result.IsOK(), Equals, true, Commentf("%+v", result.Log))
	c.Check(int64(keeper.banner2.Bond.Uint64()), Equals, int64(99900000))
	c.Check(int64(keeper.vaultData.TotalReserve.Uint64()), Equals, int64(200000))
	c.Check(keeper.toBan.ForcedToLeave, Equals, true)
	c.Check(keeper.toBan.LeaveHeight, Equals, int64(18))
	c.Check(keeper.ban.Signers, HasLen, 2)
	c.Check(keeper.ban.BlockHeight, Equals, int64(18))
}
