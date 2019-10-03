package swapservice

import (
	"encoding/json"

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

	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/client/cli"
	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/client/rest"
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
	txOutStore   *TxOutStore
	poolMgr      *PoolAddressManager
}

// NewAppModule creates a new AppModule Object
func NewAppModule(k Keeper, bankKeeper bank.Keeper, supplyKeeper supply.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
		coinKeeper:     bankKeeper,
		supplyKeeper:   supplyKeeper,
		txOutStore:     NewTxOutStore(k),
		poolMgr:        NewPoolAddressManager(k),
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
	return NewHandler(am.keeper, am.poolMgr, am.txOutStore)
}
func (am AppModule) QuerierRoute() string {
	return ModuleName
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper, am.poolMgr)
}

func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	ctx.Logger().Debug("Begin Block", "height", req.Header.Height)
	am.poolMgr.BeginBlock(ctx, req.Header.Height)
	am.txOutStore.NewBlock(uint64(req.Header.Height), am.poolMgr.GetCurrentPoolAddresses().Current)
}

func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	ctx.Logger().Debug("End Block", "height", req.Height)
	am.poolMgr.EndBlock(ctx, req.Height)
	am.txOutStore.CommitBlock(ctx)
	var validatorUpdate []abci.ValidatorUpdate

	return validatorUpdate
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
