package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// BondHandler a handler to process bond
type BondHandler struct {
	keeper                Keeper
	versionedEventManager VersionedEventManager
}

// NewBondHandler create new BondHandler
func NewBondHandler(keeper Keeper, versionedEventManager VersionedEventManager) BondHandler {
	return BondHandler{
		keeper:                keeper,
		versionedEventManager: versionedEventManager,
	}
}

func (h BondHandler) validate(ctx sdk.Context, msg MsgBond, version semver.Version, constAccessor constants.ConstantValues) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, version, msg, constAccessor)
	}
	return errBadVersion
}

func (h BondHandler) validateV1(ctx sdk.Context, version semver.Version, msg MsgBond, constAccessor constants.ConstantValues) sdk.Error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized("msg is not signed by an active node account")
	}
	minBond, err := h.keeper.GetMimir(ctx, constants.MinimumBondInRune.String())
	if minBond < 0 || err != nil {
		minBond = constAccessor.GetInt64Value(constants.MinimumBondInRune)
	}
	minValidatorBond := sdk.NewUint(uint64(minBond))

	nodeAccount, err := h.keeper.GetNodeAccount(ctx, msg.NodeAddress)
	if err != nil {
		return sdk.ErrInternal(fmt.Sprintf("fail to get node account(%s): %s", msg.NodeAddress, err))
	}

	bond := msg.Bond.Add(nodeAccount.Bond)
	if (bond).LT(minValidatorBond) {
		return sdk.ErrUnknownRequest(fmt.Sprintf("not enough rune to be whitelisted , minimum validator bond (%s) , bond(%s)", minValidatorBond.String(), bond))
	}

	maxBond, err := h.keeper.GetMimir(ctx, "MaximumBondInRune")
	if maxBond > 0 && err != nil {
		maxValidatorBond := sdk.NewUint(uint64(maxBond))
		if bond.GT(maxValidatorBond) {
			return sdk.ErrUnknownRequest(fmt.Sprintf("too much bond, max validator bond (%s), bond(%s)", maxValidatorBond.String(), bond))
		}
	}

	return nil
}

// Run execute the handler
func (h BondHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgBond)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info("receive MsgBond",
		"node address", msg.NodeAddress,
		"request hash", msg.TxIn.ID,
		"bond", msg.Bond)
	if err := h.validate(ctx, msg, version, constAccessor); err != nil {
		ctx.Logger().Error("msg bond fail validation", "error", err)
		return err.Result()
	}

	if err := h.handle(ctx, msg, version, constAccessor); err != nil {
		ctx.Logger().Error("fail to process msg bond", "error", err)
		return err.Result()
	}
	eventMgr, err := h.versionedEventManager.GetEventManager(ctx, version)
	if err != nil {
		ctx.Logger().Error("fail to get event manager", "error", err)
		return errFailGetEventManager.Result()
	}
	bondEvent := NewEventBond(msg.Bond, BondPaid, msg.TxIn)
	if err := eventMgr.EmitBondEvent(ctx, h.keeper, bondEvent); err != nil {
		return sdk.NewError(DefaultCodespace, CodeFailSaveEvent, "fail to emit bond event").Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (h BondHandler) handle(ctx sdk.Context, msg MsgBond, version semver.Version, constAccessor constants.ConstantValues) sdk.Error {
	// THORNode will not have pub keys at the moment, so have to leave it empty
	emptyPubKeySet := common.PubKeySet{
		Secp256k1: common.EmptyPubKey,
		Ed25519:   common.EmptyPubKey,
	}

	nodeAccount, err := h.keeper.GetNodeAccount(ctx, msg.NodeAddress)
	if err != nil {
		return sdk.ErrInternal(fmt.Sprintf("fail to get node account(%s): %s", msg.NodeAddress, err))
	}

	if nodeAccount.Status == NodeUnknown {
		// white list the given bep address
		nodeAccount = NewNodeAccount(msg.NodeAddress, NodeWhiteListed, emptyPubKeySet, "", sdk.ZeroUint(), msg.BondAddress, ctx.BlockHeight())
		ctx.EventManager().EmitEvent(
			sdk.NewEvent("new_node",
				sdk.NewAttribute("address", msg.NodeAddress.String()),
			))
	}

	nodeAccount.Bond = nodeAccount.Bond.Add(msg.Bond)

	if err := h.keeper.SetNodeAccount(ctx, nodeAccount); err != nil {
		return sdk.ErrInternal(fmt.Errorf("fail to save node account(%s): %w", nodeAccount, err).Error())
	}
	return h.mintGasAsset(ctx, msg, constAccessor)
}

func (h BondHandler) mintGasAsset(ctx sdk.Context, msg MsgBond, constAccessor constants.ConstantValues) sdk.Error {
	whiteListGasAsset := constAccessor.GetInt64Value(constants.WhiteListGasAsset)
	coinsToMint, err := sdk.ParseCoins(fmt.Sprintf("%dthor", whiteListGasAsset))
	if err != nil {
		return sdk.ErrInternal(fmt.Errorf("fail to parse coins: %w", err).Error())
	}
	// mint some gas asset
	err = h.keeper.Supply().MintCoins(ctx, ModuleName, coinsToMint)
	if err != nil {
		ctx.Logger().Error("fail to mint gas assets", "error", err)
		return sdk.ErrInternal(fmt.Errorf("fail to mint gas assets: %w", err).Error())
	}
	if err := h.keeper.Supply().SendCoinsFromModuleToAccount(ctx, ModuleName, msg.NodeAddress, coinsToMint); err != nil {
		return sdk.ErrInternal(fmt.Errorf("fail to send newly minted gas asset to node address(%s): %w", msg.NodeAddress, err).Error())
	}
	return nil
}
