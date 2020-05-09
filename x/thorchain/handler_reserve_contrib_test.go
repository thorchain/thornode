package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type HandlerReserveContributorSuite struct{}

var _ = Suite(&HandlerReserveContributorSuite{})

type reserveContributorKeeper struct {
	Keeper
	errGetReserveContributors bool
	errSetReserveContributors bool
	errGetVaultData           bool
	errSetVaultData           bool
	errSetEvents              bool
}

func newReserveContributorKeeper(k Keeper) *reserveContributorKeeper {
	return &reserveContributorKeeper{
		Keeper: k,
	}
}

func (k *reserveContributorKeeper) GetReservesContributors(ctx sdk.Context) (ReserveContributors, error) {
	if k.errGetReserveContributors {
		return ReserveContributors{}, kaboom
	}
	return k.Keeper.GetReservesContributors(ctx)
}

func (k *reserveContributorKeeper) SetReserveContributors(ctx sdk.Context, contributors ReserveContributors) error {
	if k.errSetReserveContributors {
		return kaboom
	}
	return k.Keeper.SetReserveContributors(ctx, contributors)
}

func (k *reserveContributorKeeper) GetVaultData(ctx sdk.Context) (VaultData, error) {
	if k.errGetVaultData {
		return VaultData{}, kaboom
	}
	return k.Keeper.GetVaultData(ctx)
}

func (k *reserveContributorKeeper) SetVaultData(ctx sdk.Context, data VaultData) error {
	if k.errSetVaultData {
		return kaboom
	}
	return k.Keeper.SetVaultData(ctx, data)
}

func (k *reserveContributorKeeper) UpsertEvent(ctx sdk.Context, event Event) error {
	if k.errSetEvents {
		return kaboom
	}
	return k.Keeper.UpsertEvent(ctx, event)
}

type reserveContributorHandlerHelper struct {
	ctx                sdk.Context
	version            semver.Version
	keeper             *reserveContributorKeeper
	nodeAccount        NodeAccount
	constAccessor      constants.ConstantValues
	reserveContributor ReserveContributor
}

func newReserveContributorHandlerHelper(c *C) reserveContributorHandlerHelper {
	ctx, k := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(1023)

	version := constants.SWVersion
	keeper := newReserveContributorKeeper(k)

	// active account
	nodeAccount := GetRandomNodeAccount(NodeActive)
	nodeAccount.Bond = sdk.NewUint(100 * common.One)
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount), IsNil)
	constAccessor := constants.GetConstantValues(version)

	reserveContributor := ReserveContributor{
		Address: GetRandomBNBAddress(),
		Amount:  sdk.NewUint(100 * common.One),
	}
	return reserveContributorHandlerHelper{
		ctx:                ctx,
		version:            version,
		keeper:             keeper,
		nodeAccount:        nodeAccount,
		constAccessor:      constAccessor,
		reserveContributor: reserveContributor,
	}
}

func (h HandlerReserveContributorSuite) TestReserveContributorHandler(c *C) {
	testCases := []struct {
		name           string
		messageCreator func(helper reserveContributorHandlerHelper) sdk.Msg
		runner         func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result
		expectedResult sdk.CodeType
		validator      func(helper reserveContributorHandlerHelper, msg sdk.Msg, result sdk.Result, c *C)
	}{
		{
			name: "invalid message should return error",
			messageCreator: func(helper reserveContributorHandlerHelper) sdk.Msg {
				return NewMsgNoOp(GetRandomObservedTx(), helper.nodeAccount.NodeAddress)
			},
			runner: func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result {
				return handler.Run(helper.ctx, msg, helper.version, helper.constAccessor)
			},
			expectedResult: CodeInvalidMessage,
		},
		{
			name: "bad version should return an error",
			messageCreator: func(helper reserveContributorHandlerHelper) sdk.Msg {
				return NewMsgReserveContributor(GetRandomTx(), helper.reserveContributor, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result {
				return handler.Run(helper.ctx, msg, semver.MustParse("0.0.1"), helper.constAccessor)
			},
			expectedResult: CodeBadVersion,
		},
		{
			name: "Not signed by an active account should return an error",
			messageCreator: func(helper reserveContributorHandlerHelper) sdk.Msg {
				return NewMsgReserveContributor(GetRandomTx(), helper.reserveContributor, GetRandomBech32Addr())
			},
			runner: func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result {
				return handler.Run(helper.ctx, msg, constants.SWVersion, helper.constAccessor)
			},
			expectedResult: sdk.CodeUnauthorized,
		},
		{
			name: "empty signer should return an error",
			messageCreator: func(helper reserveContributorHandlerHelper) sdk.Msg {
				return NewMsgReserveContributor(GetRandomTx(), helper.reserveContributor, sdk.AccAddress{})
			},
			runner: func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result {
				return handler.Run(helper.ctx, msg, constants.SWVersion, helper.constAccessor)
			},
			expectedResult: sdk.CodeInvalidAddress,
		},
		{
			name: "empty contributor address should return an error",
			messageCreator: func(helper reserveContributorHandlerHelper) sdk.Msg {
				return NewMsgReserveContributor(GetRandomTx(), ReserveContributor{
					Address: common.NoAddress,
					Amount:  sdk.NewUint(100),
				}, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result {
				return handler.Run(helper.ctx, msg, constants.SWVersion, helper.constAccessor)
			},
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name: "empty contributor amount should return an error",
			messageCreator: func(helper reserveContributorHandlerHelper) sdk.Msg {
				return NewMsgReserveContributor(GetRandomTx(), ReserveContributor{
					Address: GetRandomBNBAddress(),
					Amount:  sdk.ZeroUint(),
				}, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result {
				return handler.Run(helper.ctx, msg, constants.SWVersion, helper.constAccessor)
			},
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name: "invalid tx should return an error",
			messageCreator: func(helper reserveContributorHandlerHelper) sdk.Msg {
				tx := GetRandomTx()
				tx.ID = ""
				return NewMsgReserveContributor(tx, helper.reserveContributor, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result {
				return handler.Run(helper.ctx, msg, constants.SWVersion, helper.constAccessor)
			},
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name: "fail to get reserve contributor should return an error",
			messageCreator: func(helper reserveContributorHandlerHelper) sdk.Msg {
				return NewMsgReserveContributor(GetRandomTx(), helper.reserveContributor, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result {
				helper.keeper.errGetReserveContributors = true
				return handler.Run(helper.ctx, msg, constants.SWVersion, helper.constAccessor)
			},
			expectedResult: sdk.CodeInternal,
		},
		{
			name: "fail to set reserve contributor should return an error",
			messageCreator: func(helper reserveContributorHandlerHelper) sdk.Msg {
				return NewMsgReserveContributor(GetRandomTx(), helper.reserveContributor, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result {
				helper.keeper.errSetReserveContributors = true
				return handler.Run(helper.ctx, msg, constants.SWVersion, helper.constAccessor)
			},
			expectedResult: sdk.CodeInternal,
		},
		{
			name: "fail to get vault data should return an error",
			messageCreator: func(helper reserveContributorHandlerHelper) sdk.Msg {
				return NewMsgReserveContributor(GetRandomTx(), helper.reserveContributor, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result {
				helper.keeper.errGetVaultData = true
				return handler.Run(helper.ctx, msg, constants.SWVersion, helper.constAccessor)
			},
			expectedResult: sdk.CodeInternal,
		},
		{
			name: "fail to set vault data should return an error",
			messageCreator: func(helper reserveContributorHandlerHelper) sdk.Msg {
				return NewMsgReserveContributor(GetRandomTx(), helper.reserveContributor, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result {
				helper.keeper.errSetVaultData = true
				return handler.Run(helper.ctx, msg, constants.SWVersion, helper.constAccessor)
			},
			expectedResult: sdk.CodeInternal,
		},
		{
			name: "fail to save event should return an error",
			messageCreator: func(helper reserveContributorHandlerHelper) sdk.Msg {
				return NewMsgReserveContributor(GetRandomTx(), helper.reserveContributor, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result {
				helper.keeper.errSetEvents = true
				return handler.Run(helper.ctx, msg, constants.SWVersion, helper.constAccessor)
			},
			expectedResult: sdk.CodeInternal,
		},
		{
			name: "normal reserve contribute message should return success",
			messageCreator: func(helper reserveContributorHandlerHelper) sdk.Msg {
				return NewMsgReserveContributor(GetRandomTx(), helper.reserveContributor, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler ReserveContributorHandler, helper reserveContributorHandlerHelper, msg sdk.Msg) sdk.Result {
				return handler.Run(helper.ctx, msg, constants.SWVersion, helper.constAccessor)
			},
			validator: func(helper reserveContributorHandlerHelper, msg sdk.Msg, result sdk.Result, c *C) {
				eventID, err := helper.keeper.GetCurrentEventID(helper.ctx)
				c.Assert(err, IsNil)
				c.Assert(eventID, Equals, int64(2))
				e, err := helper.keeper.GetEvent(helper.ctx, 1)
				c.Assert(err, IsNil)
				c.Assert(e.Type, Equals, NewEventReserve(helper.reserveContributor, GetRandomTx()).Type())
			},
			expectedResult: sdk.CodeOK,
		},
	}
	for _, tc := range testCases {
		helper := newReserveContributorHandlerHelper(c)
		handler := NewReserveContributorHandler(helper.keeper, NewVersionedEventMgr())
		msg := tc.messageCreator(helper)
		result := tc.runner(handler, helper, msg)
		c.Assert(result.Code, Equals, tc.expectedResult, Commentf("name:%s", tc.name))
		if tc.validator != nil {
			tc.validator(helper, msg, result, c)
		}
	}
}
