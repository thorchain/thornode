package thorchain

import (
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	EventTypeValidatorManager   = `validator_manager`
	EventTypeNominatedValidator = `validator_nominated`
	EventTypeQueuedValidator    = `validator_queued`
	EventTypeValidatorActive    = `validator_active`
	EventTypeValidatorStandby   = `validator_standby`

	genesisBlockHeight = 1
	minValidatorSet    = 4
)

// ValidatorManager is to manage a list of validators , and rotate them
type ValidatorManager struct {
	k              Keeper
	Meta           *ValidatorMeta
	rotationPolicy ValidatorRotationPolicy
}

// NewValidatorManager create a new instance of ValidatorManager
func NewValidatorManager(k Keeper) *ValidatorManager {
	return &ValidatorManager{
		k: k,
	}
}

// BeginBlock when block begin
func (vm *ValidatorManager) BeginBlock(ctx sdk.Context, height int64) {
	vm.rotationPolicy = GetValidatorRotationPolicy(ctx, vm.k)
	if err := vm.rotationPolicy.IsValid(); nil != err {
		ctx.Logger().Error("invalid rotation policy", err)
	}
	if height == genesisBlockHeight {
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
func (vm *ValidatorManager) EndBlock(ctx sdk.Context) []abci.ValidatorUpdate {
	height := ctx.BlockHeight()
	if height != vm.Meta.RotateWindowOpenAtBlockHeight &&
		height != vm.Meta.RotateAtBlockHeight {
		return nil
	}
	defer func() {
		vm.k.SetValidatorMeta(ctx, *vm.Meta)
	}()
	if height == vm.Meta.RotateWindowOpenAtBlockHeight {
		if err := vm.prepareAddNode(ctx, height); nil != err {
			ctx.Logger().Error("fail to prepare add nodes", err)
		}
		return nil
	}
	if height == vm.Meta.RotateAtBlockHeight {
		//  queueNode put in here on purpose, so we know who had been queued before we actually rotate them
		queueNode := vm.Meta.Queued
		rotated, err := vm.rotateValidatorNodes(ctx)
		if nil != err {
			ctx.Logger().Error("fail to rotate validator nodes", err)
			return nil
		}

		if !rotated {
			return nil
		}

		activeNodes, err := vm.k.ListActiveNodeAccounts(ctx)
		if nil != err {
			ctx.Logger().Error("fail to list all active node accounts")
			return nil
		}
		validators := make([]abci.ValidatorUpdate, 0, len(activeNodes))
		for _, item := range activeNodes {
			pk, err := sdk.GetConsPubKeyBech32(item.ValidatorConsPubKey)
			if nil != err {
				ctx.Logger().Error("fail to parse consensus public key", "key", item.ValidatorConsPubKey)
				continue
			}
			validators = append(validators, abci.ValidatorUpdate{
				PubKey: tmtypes.TM2PB.PubKey(pk),
				Power:  100,
			})
		}
		for _, item := range queueNode {
			na, err := vm.k.GetNodeAccount(ctx, item.NodeAddress)
			if nil != err {
				ctx.Logger().Error("fail to get node account", err, "node address", item.NodeAddress.String())
				continue
			}
			if na.Status == NodeStandby {
				// node to be removed as validator
				pk, err := sdk.GetConsPubKeyBech32(na.ValidatorConsPubKey)
				if nil != err {
					ctx.Logger().Error("fail to parse consensus public key", "key", na.ValidatorConsPubKey)
					continue
				}
				validators = append(validators, abci.ValidatorUpdate{
					PubKey: tmtypes.TM2PB.PubKey(pk),
					Power:  0,
				})
			}
		}
		return validators
	}

	return nil
}
func (vm *ValidatorManager) rotateValidatorNodes(ctx sdk.Context) (bool, error) {
	if vm.Meta.RotateAtBlockHeight != ctx.BlockHeight() {
		// it is not an error , just not a good time
		return false, nil
	}

	defer func() {
		vm.Meta.Nominated = nil
		vm.Meta.Queued = nil
		vm.Meta.RotateAtBlockHeight += vm.rotationPolicy.RotatePerBlockHeight
		vm.Meta.RotateWindowOpenAtBlockHeight += vm.rotationPolicy.RotatePerBlockHeight
	}()
	if vm.Meta.Nominated.IsEmpty() {
		ctx.Logger().Info("no nodes get nominated , so no rotate")
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorManager,
				sdk.NewAttribute("action", "abort"),
				sdk.NewAttribute("reason", "no nominated nodes")))
		return false, nil
	}
	hasRotateIn := false
	for _, item := range vm.Meta.Nominated {
		nominatedNodeAccount, err := vm.k.GetNodeAccount(ctx, item.NodeAddress)
		if nil != err {
			return false, fmt.Errorf("fail to get nominated account from data store: %w", err)
		}

		if nominatedNodeAccount.Status != NodeReady {
			// set them to standby, do we need to slash the validator? we nominated them but they are not ready
			nominatedNodeAccount.UpdateStatus(NodeStandby, ctx.BlockHeight())
			vm.k.SetNodeAccount(ctx, nominatedNodeAccount)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(EventTypeValidatorManager,
					sdk.NewAttribute("bep_address", nominatedNodeAccount.NodeAddress.String()),
					sdk.NewAttribute("consensus_public_key", nominatedNodeAccount.ValidatorConsPubKey),
					sdk.NewAttribute("action", "abort"),
					sdk.NewAttribute("reason", "node not ready")))
			ctx.Logger().Info(fmt.Sprintf("nominated account %s is not ready , abort rotation, try it nex time", item.NodeAddress))
			// go to the next
			continue
		}
		// set to active
		nominatedNodeAccount.UpdateStatus(NodeActive, ctx.BlockHeight())
		vm.k.SetNodeAccount(ctx, nominatedNodeAccount)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorActive,
				sdk.NewAttribute("bep_address", nominatedNodeAccount.NodeAddress.String()),
				sdk.NewAttribute("consensus_public_key", nominatedNodeAccount.ValidatorConsPubKey)))
		hasRotateIn = true
	}
	if !hasRotateIn {
		return false, errors.New("none of the nominated node is ready, give up")
	}

	if !vm.Meta.Queued.IsEmpty() {
		for _, item := range vm.Meta.Queued {
			item.UpdateStatus(NodeStandby, ctx.BlockHeight())
			vm.k.SetNodeAccount(ctx, item)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(EventTypeValidatorStandby,
					sdk.NewAttribute("bep_address", item.NodeAddress.String()),
					sdk.NewAttribute("consensus_public_key", item.ValidatorConsPubKey)))
		}
	}

	return true, nil
}
func (vm *ValidatorManager) prepareAddNode(ctx sdk.Context, height int64) error {
	if height != vm.Meta.RotateWindowOpenAtBlockHeight {
		// it is not an error , just not a good time to do this yet
		return nil
	}
	// who should be added , and who need to removed
	standbyNodes, err := vm.k.ListNodeAccountsByStatus(ctx, NodeStandby)
	if nil != err {
		return fmt.Errorf("fail to get all standby nodes,%w", err)
	}
	if len(standbyNodes) == 0 {
		ctx.Logger().Info("no standby nodes")
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorManager,
				sdk.NewAttribute("action", "abort"),
				sdk.NewAttribute("reason", "no standby nodes")))
		return nil
	}
	sort.Sort(standbyNodes)
	activeNodes, err := vm.k.ListActiveNodeAccounts(ctx)
	if nil != err {
		return fmt.Errorf("fail to get active node accounts: %w", err)
	}
	totalActiveNodes := len(activeNodes)
	rotateIn := vm.rotationPolicy.RotateInNumBeforeFull
	rotateOut := vm.rotationPolicy.RotateOutNumBeforeFull
	if int64(totalActiveNodes) >= vm.rotationPolicy.DesireValidatorSet {
		// we are full
		rotateIn = vm.rotationPolicy.RotateNumAfterFull
		rotateOut = vm.rotationPolicy.RotateNumAfterFull
	}
	if int64(len(standbyNodes)) > rotateIn {
		standbyNodes = standbyNodes[:rotateIn]
	}
	vm.Meta.Nominated = standbyNodes
	for _, item := range vm.Meta.Nominated {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeNominatedValidator,
				sdk.NewAttribute("bep_address", item.NodeAddress.String()),
				sdk.NewAttribute("consensus_public_key", item.ValidatorConsPubKey)))
	}
	// we need to set a minimum validator set , if we have less than the minimum , then we don't rotate out
	if totalActiveNodes <= minValidatorSet {
		rotateOut = 0
	}

	if rotateOut > 0 {
		activeNodesBySlash := NodeAccountsBySlashingPoint(activeNodes)
		sort.Sort(activeNodesBySlash)
		// Queue the first few nodes to be rotated out
		vm.Meta.Queued = NodeAccounts(activeNodesBySlash[:rotateOut])
		for _, item := range vm.Meta.Queued {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(EventTypeQueuedValidator,
					sdk.NewAttribute("bep_address", item.NodeAddress.String()),
					sdk.NewAttribute("consensus_public_key", item.ValidatorConsPubKey)))
		}
	}
	return nil
}

// setupValidatorNodes it is one off it only get called when genesis
func (vm *ValidatorManager) setupValidatorNodes(ctx sdk.Context, height int64) error {
	if height != genesisBlockHeight {
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
			return fmt.Errorf("fail to unmarshal node account, %w", err)
		}
		// when we first start , we only care about these two status
		switch na.Status {
		case NodeReady:
			readyNodes = append(readyNodes, na)
		case NodeActive:
			activeCandidateNodes = append(activeCandidateNodes, na)
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
	desireValidatorSet := vm.rotationPolicy.DesireValidatorSet
	for idx, item := range activeCandidateNodes {
		if int64(idx) < desireValidatorSet {
			item.UpdateStatus(NodeActive, ctx.BlockHeight())
		} else {
			item.UpdateStatus(NodeStandby, ctx.BlockHeight())
		}
		vm.k.SetNodeAccount(ctx, item)
	}
	vm.Meta.RotateAtBlockHeight = vm.rotationPolicy.RotatePerBlockHeight + 1
	vm.Meta.RotateWindowOpenAtBlockHeight = vm.rotationPolicy.RotatePerBlockHeight + 1 - vm.rotationPolicy.ValidatorChangeWindow
	return nil
}
