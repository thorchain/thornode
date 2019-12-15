package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

type TssHandler struct {
	keeper       Keeper
	txOutStore   TxOutStore
	poolAddrMgr  PoolAddressManager
	validatorMgr ValidatorManager
}

func NewTssHandler(keeper Keeper, txOutStore TxOutStore, poolAddrMgr PoolAddressManager, validatorMgr ValidatorManager) TssHandler {
	return TssHandler{
		keeper:       keeper,
		txOutStore:   txOutStore,
		poolAddrMgr:  poolAddrMgr,
		validatorMgr: validatorMgr,
	}
}

func (h TssHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version) sdk.Result {
	msg, ok := m.(MsgTssPool)
	if !ok {
		return errInvalidMessage.Result()
	}
	isNewSigner, err := h.validate(ctx, msg, version)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	if isNewSigner {
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}
	return h.handle(ctx, msg, version)
}

func (h TssHandler) validate(ctx sdk.Context, msg MsgTssPool, version semver.Version) (bool, error) {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return false, badVersion
	}
}

func (h TssHandler) validateV1(ctx sdk.Context, msg MsgTssPool) (bool, error) {
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return false, err
	}

	// No auth check, as not all validators are "active" who are apart of a
	// keygen process

	return false, nil
}

func (h TssHandler) handle(ctx sdk.Context, msg MsgTssPool, version semver.Version) sdk.Result {
	ctx.Logger().Info("handleMsgTssPool request", "ID:", msg.ID)
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return errBadVersion.Result()
	}
}

func (h TssHandler) rotatePoolAddress(ctx sdk.Context, voter TssVoter) error {
	chains, err := h.keeper.GetChains(ctx)
	if err != nil {
		return nil
	}

	poolpks := make(common.PoolPubKeys, len(chains))
	for i, chain := range chains {
		var err error
		poolpks[i], err = common.NewPoolPubKey(chain, 0, voter.PoolPubKey)
		if err != nil {
			return nil
		}
	}

	h.poolAddrMgr.RotatePoolAddress(ctx, poolpks, h.txOutStore)
	return nil
}

// Handle a message to observe inbound tx
func (h TssHandler) handleV1(ctx sdk.Context, msg MsgTssPool) sdk.Result {
	active, err := h.keeper.ListActiveNodeAccounts(ctx)
	if nil != err {
		err = wrapError(ctx, err, "fail to get list of active node accounts")
		return sdk.ErrInternal(err.Error()).Result()
	}

	voter, err := h.keeper.GetTssVoter(ctx, msg.ID)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	voter.Sign(msg.Signer)

	if voter.HasConensus(active) && voter.BlockHeight == 0 {
		voter.BlockHeight = ctx.BlockHeight()
		h.keeper.SetTssVoter(ctx, voter)

		if err := h.rotatePoolAddress(ctx, voter); err != nil {
			return sdk.ErrInternal(err.Error()).Result()
		}

	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
