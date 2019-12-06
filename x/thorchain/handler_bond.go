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

func (bh BondHandler) validate(ctx sdk.Context, msg MsgBond, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return bh.validateV1(ctx, msg)
	}
	return errBadVersion
}

func (bh BondHandler) validateV1(ctx sdk.Context, msg MsgBond) sdk.Error {
	if err := msg.ValidateBasic(); nil != err {
		return err
	}

	if !isSignedByActiveNodeAccounts(ctx, bh.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized("msg is not signed by an active node account")
	}

	minValidatorBond := sdk.NewUint(constants.MinimumBondInRune)
	if msg.Bond.LT(minValidatorBond) {
		return sdk.ErrUnknownRequest(fmt.Sprintf("not enough rune to be whitelisted , minimum validator bond (%s) , bond(%s)", minValidatorBond.String(), msg.Bond))
	}

	nodeAccount, err := bh.keeper.GetNodeAccount(ctx, msg.NodeAddress)
	if nil != err {
		return sdk.ErrInternal(fmt.Sprintf("fail to get node account(%s): %s", msg.NodeAddress, err))
	}

	if !nodeAccount.IsEmpty() {
		return sdk.ErrInternal(fmt.Sprintf("node account(%s) already exist", msg.NodeAddress))
	}

	return nil
}

// Run execute the handler
func (bh BondHandler) Run(ctx sdk.Context, msg MsgBond, version semver.Version) sdk.Result {
	ctx.Logger().Info("receive MsgBond",
		"node address", msg.NodeAddress,
		"request hash", msg.RequestTxHash,
		"bond", msg.Bond)
	if err := bh.validate(ctx, msg, version); nil != err {
		ctx.Logger().Error("msg bond fail validation", err)
		return err.Result()
	}

	if err := bh.handle(ctx, msg, version); nil != err {
		ctx.Logger().Error("fail to process msg bond", err)
		return err.Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (bh BondHandler) handle(ctx sdk.Context, msg MsgBond, version semver.Version) sdk.Error {
	// THORNode will not have pub keys at the moment, so have to leave it empty
	emptyPubKeys := common.PubKeys{
		Secp256k1: common.EmptyPubKey,
		Ed25519:   common.EmptyPubKey,
	}
	// white list the given bep address
	nodeAccount := NewNodeAccount(msg.NodeAddress, NodeWhiteListed, emptyPubKeys, "", msg.Bond, msg.BondAddress, ctx.BlockHeight())
	if err := bh.keeper.SetNodeAccount(ctx, nodeAccount); nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to save node account(%s): %w", nodeAccount, err).Error())
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent("new_node",
			sdk.NewAttribute("address", msg.NodeAddress.String()),
		))

	return bh.mintGasAsset(ctx, msg)
}

func (bh BondHandler) mintGasAsset(ctx sdk.Context, msg MsgBond) sdk.Error {
	coinsToMint := bh.keeper.GetAdminConfigWhiteListGasAsset(ctx, sdk.AccAddress{})
	// mint some gas asset
	err := bh.keeper.Supply().MintCoins(ctx, ModuleName, coinsToMint)
	if nil != err {
		ctx.Logger().Error("fail to mint gas assets", "err", err)
		return sdk.ErrInternal(fmt.Errorf("fail to mint gas assets: %w", err).Error())
	}
	if err := bh.keeper.Supply().SendCoinsFromModuleToAccount(ctx, ModuleName, msg.NodeAddress, coinsToMint); nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to send newly minted gas asset to node address(%s): %w", msg.NodeAddress, err).Error())
	}
	return nil
}
