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
	keeper       Keeper
	coinKeeper   bank.Keeper
	supplyKeeper supply.Keeper
	txOutStore   TxOutStore
	validatorMgr ValidatorManager
	vaultMgr     VaultManager
}

// NewAppModule creates a new AppModule Object
func NewAppModule(k Keeper, bankKeeper bank.Keeper, supplyKeeper supply.Keeper) AppModule {
	txStore := NewTxOutStorage(k)
	vaultMgr := NewVaultMgr(k, txStore)
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
		coinKeeper:     bankKeeper,
		supplyKeeper:   supplyKeeper,
		txOutStore:     txStore,
		validatorMgr:   NewValidatorMgr(k, txStore, vaultMgr),
		vaultMgr:       vaultMgr,
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
	return NewHandler(am.keeper, am.txOutStore, am.validatorMgr, am.vaultMgr)
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
	constantValues := constants.GetConstantValues(version)
	if nil == constantValues {
		ctx.Logger().Error(fmt.Sprintf("constants for version(%s) is not available", version))
	} else {
		if err := am.validatorMgr.BeginBlock(ctx, constantValues); err != nil {
			ctx.Logger().Error("Fail to begin block on validator", err)
		}

		am.txOutStore.NewBlock(uint64(req.Header.Height), constantValues)
	}
}

func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	ctx.Logger().Debug("End Block", "height", req.Height)

	version := am.keeper.GetLowestActiveVersion(ctx)
	constantValues := constants.GetConstantValues(version)
	if nil == constantValues {
		ctx.Logger().Error(fmt.Sprintf("constants for version(%s) is not available", version))
	} else {
		slasher := NewSlasher(am.keeper, am.txOutStore)
		// slash node accounts for not observing any accepted inbound tx
		if err := slasher.LackObserving(ctx, constantValues); err != nil {
			ctx.Logger().Error("Unable to slash for lack of observing:", err)
		}
		if err := slasher.LackSigning(ctx, constantValues); err != nil {
			ctx.Logger().Error("Unable to slash for lack of signing:", err)
		}
		newPoolCycle := constantValues.GetInt64Value(constants.NewPoolCycle)
		// Enable a pool every newPoolCycle
		if ctx.BlockHeight()%newPoolCycle == 0 {
			if err := enableNextPool(ctx, am.keeper); err != nil {
				ctx.Logger().Error("Unable to enable a pool", err)
			}
		}

		// Fill up Yggdrasil vaults
		err := Fund(ctx, am.keeper, am.txOutStore, constantValues)
		if err != nil {
			ctx.Logger().Error("Unable to fund Yggdrasil", err)
		}

		// update vault data to account for block rewards and reward units
		if err := am.keeper.UpdateVaultData(ctx, constantValues); nil != err {
			ctx.Logger().Error("fail to save vault", err)
		}

		if err := am.vaultMgr.EndBlock(ctx, constantValues); err != nil {
			ctx.Logger().Error("fail to end block for vault manager", err)
		}

		am.txOutStore.CommitBlock(ctx)
		return am.validatorMgr.EndBlock(ctx, constantValues)
	}
	return nil
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
