package thorchain

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	sdkRest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"gitlab.com/thorchain/thornode/constants"

	"gitlab.com/thorchain/thornode/x/thorchain/client/cli"
	"gitlab.com/thorchain/thornode/x/thorchain/client/rest"
)

// type check to ensure the interface is properly implemented
var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// app module Basics object
type AppModuleBasic struct{}

func (AppModuleBasic) Name() string {
	return ModuleName
}

func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return ModuleCdc.MustMarshalJSON(DefaultGenesisState())
}

// Validation check of the Genesis
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data GenesisState
	err := ModuleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	// Once json successfully marshalled, passes along to genesis.go
	return ValidateGenesis(data)
}

// Register rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	rest.RegisterRoutes(ctx, rtr, StoreKey)
	sdkRest.RegisterTxRoutes(ctx, rtr)
	sdkRest.RegisterRoutes(ctx, rtr, StoreKey)
}

// Get the root query command of this module
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueryCmd(StoreKey, cdc)
}

// Get the root tx command of this module
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(StoreKey, cdc)
}

type AppModule struct {
	AppModuleBasic
	keeper                   Keeper
	coinKeeper               bank.Keeper
	supplyKeeper             supply.Keeper
	txOutStore               VersionedTxOutStore
	validatorMgr             VersionedValidatorManager
	versionedVaultManager    VersionedVaultManager
	versionedGasManager      VersionedGasManager
	versionedObserverManager VersionedObserverManager
	versionedEventManager    VersionedEventManager
}

// NewAppModule creates a new AppModule Object
func NewAppModule(k Keeper, bankKeeper bank.Keeper, supplyKeeper supply.Keeper) AppModule {
	versionedEventManager := NewVersionedEventMgr()
	versionedTxOutStore := NewVersionedTxOutStore(versionedEventManager)
	versionedVaultMgr := NewVersionedVaultMgr(versionedTxOutStore, versionedEventManager)
	return AppModule{
		AppModuleBasic:           AppModuleBasic{},
		keeper:                   k,
		coinKeeper:               bankKeeper,
		supplyKeeper:             supplyKeeper,
		txOutStore:               versionedTxOutStore,
		validatorMgr:             NewVersionedValidatorMgr(k, versionedTxOutStore, versionedVaultMgr, versionedEventManager),
		versionedVaultManager:    versionedVaultMgr,
		versionedGasManager:      NewVersionedGasMgr(),
		versionedObserverManager: NewVersionedObserverMgr(),
		versionedEventManager:    versionedEventManager,
	}
}

func (AppModule) Name() string {
	return ModuleName
}

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

func (am AppModule) Route() string {
	return RouterKey
}

func (am AppModule) NewHandler() sdk.Handler {
	return NewExternalHandler(am.keeper, am.txOutStore, am.validatorMgr, am.versionedVaultManager, am.versionedObserverManager, am.versionedGasManager, am.versionedEventManager)
}

func (am AppModule) QuerierRoute() string {
	return ModuleName
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper, am.validatorMgr)
}

func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	ctx.Logger().Debug("Begin Block", "height", req.Header.Height)
	version := am.keeper.GetLowestActiveVersion(ctx)
	am.keeper.ClearObservingAddresses(ctx)
	obMgr, err := am.versionedObserverManager.GetObserverManager(ctx, version)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("observer manager that compatible with version :%s is not available", version))
		return
	}
	obMgr.BeginBlock()

	gasMgr, err := am.versionedGasManager.GetGasManager(ctx, version)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("gas manager that compatible with version :%s is not available", version))
		return
	}
	gasMgr.BeginBlock()
	constantValues := constants.GetConstantValues(version)
	if constantValues == nil {
		ctx.Logger().Error(fmt.Sprintf("constants for version(%s) is not available", version))
		return
	}

	slasher, err := NewSlasher(am.keeper, version, am.versionedEventManager)
	if err != nil {
		ctx.Logger().Error("fail to create slasher", "error", err)
	}
	slasher.BeginBlock(ctx, req, constantValues)

	if err := am.validatorMgr.BeginBlock(ctx, version, constantValues); err != nil {
		ctx.Logger().Error("Fail to begin block on validator", "error", err)
	}
	txStore, err := am.txOutStore.GetTxOutStore(ctx, am.keeper, version)
	if err != nil {
		ctx.Logger().Error("fail to get tx out store", "error", err)
		return
	}
	txStore.NewBlock(req.Header.Height, constantValues)
}

func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	ctx.Logger().Debug("End Block", "height", req.Height)

	version := am.keeper.GetLowestActiveVersion(ctx)
	constantValues := constants.GetConstantValues(version)
	if constantValues == nil {
		ctx.Logger().Error(fmt.Sprintf("constants for version(%s) is not available", version))
		return nil
	}
	txStore, err := am.txOutStore.GetTxOutStore(ctx, am.keeper, version)
	if err != nil {
		ctx.Logger().Error("fail to get tx out store", "error", err)
		return nil
	}
	eventMgr, err := am.versionedEventManager.GetEventManager(ctx, version)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Events manager that compatible with version :%s is not available", version))
		return nil
	}

	swapQueue, err := NewVersionedSwapQ(am.txOutStore, am.versionedEventManager).GetSwapQueue(ctx, am.keeper, version)
	if err != nil {
		ctx.Logger().Error("fail to get swap queue", "error", err)
	} else {
		if err := swapQueue.EndBlock(ctx, version, constantValues); err != nil {
			ctx.Logger().Error("fail to process swap queue", "error", err)
		}
	}

	slasher, err := NewSlasher(am.keeper, version, am.versionedEventManager)
	if err != nil {
		ctx.Logger().Error("fail to create slasher", "error", err)
		return nil
	}
	// slash node accounts for not observing any accepted inbound tx
	if err := slasher.LackObserving(ctx, constantValues); err != nil {
		ctx.Logger().Error("Unable to slash for lack of observing:", "error", err)
	}
	if err := slasher.LackSigning(ctx, constantValues, txStore); err != nil {
		ctx.Logger().Error("Unable to slash for lack of signing:", "error", err)
	}
	newPoolCycle := constantValues.GetInt64Value(constants.NewPoolCycle)
	// Enable a pool every newPoolCycle
	if ctx.BlockHeight()%newPoolCycle == 0 {
		if err := enableNextPool(ctx, am.keeper, eventMgr); err != nil {
			ctx.Logger().Error("Unable to enable a pool", "error", err)
		}
	}

	// fail stale pending events
	signingTransPeriod := constantValues.GetInt64Value(constants.SigningTransactionPeriod)
	pendingEvents, err := am.keeper.GetAllPendingEvents(ctx)
	if err != nil {
		ctx.Logger().Error("Unable to get all pending events", "error", err)
	}
	for _, evt := range pendingEvents {
		if evt.Height+(2*signingTransPeriod) < ctx.BlockHeight() {
			evt.Status = EventFail
			if err := am.keeper.UpsertEvent(ctx, evt); err != nil {
				ctx.Logger().Error("Unable to update pending event", "error", err)
			}
		}
	}

	obMgr, err := am.versionedObserverManager.GetObserverManager(ctx, version)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("observer manager that compatible with version :%s is not available", version))
		return nil
	}
	obMgr.EndBlock(ctx, am.keeper)
	gasMgr, err := am.versionedGasManager.GetGasManager(ctx, version)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("gas manager that compatible with version :%s is not available", version))
		return nil
	}
	// update vault data to account for block rewards and reward units
	if err := am.keeper.UpdateVaultData(ctx, constantValues, gasMgr, eventMgr); err != nil {
		ctx.Logger().Error("fail to save vault", "error", err)
	}
	vaultMgr, err := am.versionedVaultManager.GetVaultManager(ctx, am.keeper, version)
	if err != nil {
		ctx.Logger().Error("fail to get a valid vault manager", "error", err)
		return nil
	}

	if err := vaultMgr.EndBlock(ctx, version, constantValues); err != nil {
		ctx.Logger().Error("fail to end block for vault manager", "error", err)
	}

	validators := am.validatorMgr.EndBlock(ctx, version, constantValues)

	// Fill up Yggdrasil vaults
	// We do this AFTER validatorMgr.EndBlock, because we don't want to send
	// funds to a yggdrasil vault that is being churned out this block.
	if err := Fund(ctx, am.keeper, txStore, constantValues); err != nil {
		ctx.Logger().Error("unable to fund yggdrasil", "error", err)
	}
	gasMgr.EndBlock(ctx, am.keeper, eventMgr)

	return validators
}

func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	return InitGenesis(ctx, am.keeper, genesisState)
}

func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return ModuleCdc.MustMarshalJSON(gs)
}
