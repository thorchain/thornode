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
	keeper Keeper
}

// NewBondHandler create new BondHandler
func NewBondHandler(keeper Keeper) BondHandler {
	return BondHandler{keeper: keeper}
}

func (h BondHandler) validate(ctx sdk.Context, msg MsgBond, version semver.Version, constAccessor constants.ConstantValues) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, version, msg, constAccessor)
	}
	return errBadVersion
}

func (h BondHandler) validateV1(ctx sdk.Context, version semver.Version, msg MsgBond, constAccessor constants.ConstantValues) sdk.Error {
	if err := msg.ValidateBasic(); nil != err {
		return err
	}

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized("msg is not signed by an active node account")
	}
	minimumBond := constAccessor.GetInt64Value(constants.MinimumBondInRune)
	minValidatorBond := sdk.NewUint(uint64(minimumBond))
	if msg.Bond.LT(minValidatorBond) {
		return sdk.ErrUnknownRequest(fmt.Sprintf("not enough rune to be whitelisted , minimum validator bond (%s) , bond(%s)", minValidatorBond.String(), msg.Bond))
	}

	nodeAccount, err := h.keeper.GetNodeAccount(ctx, msg.NodeAddress)
	if nil != err {
		return sdk.ErrInternal(fmt.Sprintf("fail to get node account(%s): %s", msg.NodeAddress, err))
	}

	if !nodeAccount.IsEmpty() {
		return sdk.ErrInternal(fmt.Sprintf("node account(%s) already exist", msg.NodeAddress))
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
		"request hash", msg.RequestTxHash,
		"bond", msg.Bond)
	if err := h.validate(ctx, msg, version, constAccessor); nil != err {
		ctx.Logger().Error("msg bond fail validation", err)
		return err.Result()
	}

	if err := h.handle(ctx, msg, version, constAccessor); nil != err {
		ctx.Logger().Error("fail to process msg bond", err)
		return err.Result()
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
	// white list the given bep address
	nodeAccount := NewNodeAccount(msg.NodeAddress, NodeWhiteListed, emptyPubKeySet, "", msg.Bond, msg.BondAddress, ctx.BlockHeight())
	if err := h.keeper.SetNodeAccount(ctx, nodeAccount); nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to save node account(%s): %w", nodeAccount, err).Error())
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent("new_node",
			sdk.NewAttribute("address", msg.NodeAddress.String()),
		))

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
	if nil != err {
		ctx.Logger().Error("fail to mint gas assets", "err", err)
		return sdk.ErrInternal(fmt.Errorf("fail to mint gas assets: %w", err).Error())
	}
	if err := h.keeper.Supply().SendCoinsFromModuleToAccount(ctx, ModuleName, msg.NodeAddress, coinsToMint); nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to send newly minted gas asset to node address(%s): %w", msg.NodeAddress, err).Error())
	}
	return nil
}
