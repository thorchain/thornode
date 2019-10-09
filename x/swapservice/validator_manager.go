package swapservice

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	EventTypeValidatorManager   = `validator_manager`
	EventTypeNominatedValidator = `validator_nominated`
	EventTypeQueuedValidator    = `validator_queued`
	EventTypeValidatorActive    = `validator_active`
	EventTypeValidatorStandby   = `validator_standby`
)

// ValidatorManager is to manage a list of validators , and rotate them
type ValidatorManager struct {
	k    Keeper
	Meta *ValidatorMeta
}

// NewValidatorManager create a new instance of ValidatorManager
func NewValidatorManager(k Keeper) *ValidatorManager {
	return &ValidatorManager{
		k: k,
	}
}

// BeginBlock when block begin
func (vm *ValidatorManager) BeginBlock(ctx sdk.Context, height int64) {
	if height == 1 {
		vm.Meta = &ValidatorMeta{}
		if err := vm.setupValidatorNodes(ctx, height); nil != err {
			ctx.Logger().Error("fail to setup validator nodes", err)
		}
		vm.k.SetValidatorMeta(ctx, *vm.Meta)
	}

	// restore vm.meta from data store
	if vm.Meta == nil {
		meta := vm.k.GetValidatorMeta(ctx)
		vm.Meta = &meta
	}
}

// EndBlock when block end
func (vm *ValidatorManager) EndBlock(ctx sdk.Context, height int64) []abci.ValidatorUpdate {
	if height == vm.Meta.RotateWindowOpenAtBlockHeight {
		if err := vm.prepareAddNode(ctx, height); nil != err {
			ctx.Logger().Error("fail to prepare add nodes", err)
		}
		vm.k.SetValidatorMeta(ctx, *vm.Meta)
	}
	if height == vm.Meta.RotateAtBlockHeight {
		defer func() {
			vm.k.SetValidatorMeta(ctx, *vm.Meta)
		}()
		queueNode := vm.Meta.Queued
		rotated, err := vm.rotateValidatorNodes(ctx, height)
		if nil != err {
			ctx.Logger().Error("fail to rotate validator nodes")
			return nil
		}
		if rotated {
			activeNodes, err := vm.k.ListActiveNodeAccounts(ctx)
			if nil != err {
				ctx.Logger().Error("fail to list all active node accounts")
				return nil
			}
			validators := make([]abci.ValidatorUpdate, 0, len(activeNodes))
			for _, item := range activeNodes {
				pk, err := sdk.GetConsPubKeyBech32(item.Accounts.ValidatorBEPConsPubKey)
				if nil != err {
					ctx.Logger().Error("fail to parse consensus public key", "key", item.Accounts.ValidatorBEPConsPubKey)
					continue
				}
				validators = append(validators, abci.ValidatorUpdate{
					PubKey: tmtypes.TM2PB.PubKey(pk),
					Power:  100,
				})
			}
			if !queueNode.IsEmpty() {
				// node to be removed as validator
				pk, err := sdk.GetConsPubKeyBech32(queueNode.Accounts.ValidatorBEPConsPubKey)
				if nil != err {
					ctx.Logger().Error("fail to parse consensus public key", "key", queueNode.Accounts.ValidatorBEPConsPubKey)
				}
				validators = append(validators, abci.ValidatorUpdate{
					PubKey: tmtypes.TM2PB.PubKey(pk),
					Power:  0,
				})
			}

			return validators
		}
	}

	return nil
}

func (vm *ValidatorManager) rotateValidatorNodes(ctx sdk.Context, height int64) (bool, error) {
	if vm.Meta.RotateAtBlockHeight != height {
		// it is not an error , just not a good time
		return false, nil
	}
	rotatePerBlockHeight := vm.k.GetAdminConfigRotatePerBlockHeight(ctx, sdk.AccAddress{})
	defer func() {
		vm.Meta.Nominated = NodeAccount{}
		vm.Meta.Queued = NodeAccount{}
		vm.Meta.RotateAtBlockHeight += rotatePerBlockHeight
	}()
	if vm.Meta.Nominated.IsEmpty() {
		ctx.Logger().Info("no node get nominated , so no rotate")
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorManager,
				sdk.NewAttribute("action", "abort"),
				sdk.NewAttribute("reason", "no nominated nodes")))
		return false, nil
	}

	nominatedNodeAccount, err := vm.k.GetNodeAccount(ctx, vm.Meta.Nominated.NodeAddress)
	if nil != err {
		return false, errors.Wrap(err, "fail to get nominated account from data store")
	}
	if nominatedNodeAccount.Status != NodeReady {
		vm.Meta.Nominated = NodeAccount{}
		vm.Meta.Queued = NodeAccount{}
		// set them to standby
		nominatedNodeAccount.UpdateStatus(NodeStandby)
		vm.k.SetNodeAccount(ctx, nominatedNodeAccount)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorManager,
				sdk.NewAttribute("action", "abort"),
				sdk.NewAttribute("reason", "nominated node not ready")))
		ctx.Logger().Info("nominated account is not ready , abort rotation, try it nex time")
		return false, nil
	}

	// set to active
	nominatedNodeAccount.UpdateStatus(NodeActive)
	vm.k.SetNodeAccount(ctx, nominatedNodeAccount)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeValidatorActive,
			sdk.NewAttribute("bep_address", nominatedNodeAccount.NodeAddress.String()),
			sdk.NewAttribute("consensus_public_key", nominatedNodeAccount.Accounts.ValidatorBEPConsPubKey)))

	if !vm.Meta.Queued.IsEmpty() {
		outNode := vm.Meta.Queued
		vm.Meta.Queued.UpdateStatus(NodeStandby)
		vm.k.SetNodeAccount(ctx, vm.Meta.Queued)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorStandby,
				sdk.NewAttribute("bep_address", outNode.NodeAddress.String()),
				sdk.NewAttribute("consensus_public_key", outNode.Accounts.ValidatorBEPConsPubKey)))
	}

	return true, nil
}

func (vm *ValidatorManager) prepareAddNode(ctx sdk.Context, height int64) error {
	if height != vm.Meta.RotateWindowOpenAtBlockHeight {
		// it is not an error , just not a good time to do this yet
		return nil
	}
	rotatePerBlockHeight := vm.k.GetAdminConfigRotatePerBlockHeight(ctx, sdk.AccAddress{})
	defer func() {
		vm.Meta.RotateWindowOpenAtBlockHeight += rotatePerBlockHeight
	}()
	// who should be added , an who need to removed
	standbyNodes, err := vm.k.ListNodeAccountsByStatus(ctx, NodeStandby)
	if nil != err {
		return errors.Wrap(err, "fail to get all standby nodes")
	}
	if len(standbyNodes) == 0 {
		ctx.Logger().Info("no standby nodes")
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorManager,
				sdk.NewAttribute("action", "abort"),
				sdk.NewAttribute("reason", "no standby nodes")))
		return nil
	}
	sort.Slice(standbyNodes, func(i, j int) bool {
		return standbyNodes[i].StatusSince < standbyNodes[j].StatusSince
	})
	vm.Meta.Nominated = standbyNodes.First()
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeNominatedValidator,
			sdk.NewAttribute("bep_address", vm.Meta.Nominated.NodeAddress.String()),
			sdk.NewAttribute("consensus_public_key", vm.Meta.Nominated.Accounts.ValidatorBEPConsPubKey)))

	activeNodes, err := vm.k.ListActiveNodeAccounts(ctx)
	if nil != err {
		return errors.Wrap(err, "fail to get active node accounts")
	}
	desireValidatorSet := vm.k.GetAdminConfigDesireValidatorSet(ctx, sdk.AccAddress{})
	if int64(len(activeNodes)) >= desireValidatorSet {
		sort.Slice(activeNodes, func(i, j int) bool {
			return activeNodes[i].StatusSince < activeNodes[j].StatusSince
		})
		vm.Meta.Queued = activeNodes.First()
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeQueuedValidator,
				sdk.NewAttribute("bep_address", vm.Meta.Queued.NodeAddress.String()),
				sdk.NewAttribute("consensus_public_key", vm.Meta.Queued.Accounts.ValidatorBEPConsPubKey)))
	}
	return nil
}

// setupValidatorNodes only works when statechain start up
func (vm *ValidatorManager) setupValidatorNodes(ctx sdk.Context, height int64) error {
	if height != 1 {
		ctx.Logger().Info("only need to setup validator node when start up", "height", height)
		return nil
	}

	iter := vm.k.GetNodeAccountIterator(ctx)
	defer iter.Close()
	readyNodes := NodeAccounts{}
	activeCandidateNodes := NodeAccounts{}
	for ; iter.Valid(); iter.Next() {
		var na NodeAccount
		if err := vm.k.cdc.UnmarshalBinaryBare(iter.Value(), &na); nil != err {
			ctx.Logger().Error("fail to unmarshal node account", "error", err)
			return errors.Wrap(err, "fail to unmarshal node account")
		}
		// when we first start , we only care about these two status
		switch na.Status {
		case NodeReady:
			readyNodes = append(readyNodes, na)
		case NodeActive:
			activeCandidateNodes = append(activeCandidateNodes, na)
			//vm.ValidatorNodes.ActiveValidators = append(vm.ValidatorNodes.ActiveValidators, na)
		}
	}
	totalActiveValidators := len(activeCandidateNodes)
	totalNominatedValidators := len(readyNodes)
	if totalActiveValidators == 0 && totalNominatedValidators == 0 {
		return errors.New("no validators available")
	}

	sort.Sort(activeCandidateNodes)
	sort.Sort(readyNodes)
	activeCandidateNodes = append(activeCandidateNodes, readyNodes...)
	desireValidatorSet := vm.k.GetAdminConfigDesireValidatorSet(ctx, sdk.AccAddress{})
	for idx, item := range activeCandidateNodes {
		if int64(idx) < desireValidatorSet {
			item.UpdateStatus(NodeActive)

		} else {
			item.UpdateStatus(NodeStandby)
		}
		vm.k.SetNodeAccount(ctx, item)
	}
	rotatePerBlockHeight := vm.k.GetAdminConfigRotatePerBlockHeight(ctx, sdk.AccAddress{})
	validatorChangeWindow := vm.k.GetAdminConfigValidatorsChangeWindow(ctx, sdk.AccAddress{})
	vm.Meta.RotateAtBlockHeight = rotatePerBlockHeight + 1
	vm.Meta.RotateWindowOpenAtBlockHeight = rotatePerBlockHeight + 1 - validatorChangeWindow
	return nil
}
