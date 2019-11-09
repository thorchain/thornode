package thorchain

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

	"gitlab.com/thorchain/bepswap/thornode/x/thorchain/client/cli"
	"gitlab.com/thorchain/bepswap/thornode/x/thorchain/client/rest"
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
	validatorMgr *ValidatorManager
}

// NewAppModule creates a new AppModule Object
func NewAppModule(k Keeper, bankKeeper bank.Keeper, supplyKeeper supply.Keeper) AppModule {
	poolAddrMgr := NewPoolAddressManager(k)
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
		coinKeeper:     bankKeeper,
		supplyKeeper:   supplyKeeper,
		txOutStore:     NewTxOutStore(k, poolAddrMgr),
		poolMgr:        poolAddrMgr,
		validatorMgr:   NewValidatorManager(k),
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
	return NewHandler(am.keeper, am.poolMgr, am.txOutStore, am.validatorMgr)
}
func (am AppModule) QuerierRoute() string {
	return ModuleName
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper, am.poolMgr, am.validatorMgr)
}

func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	ctx.Logger().Debug("Begin Block", "height", req.Header.Height)
	am.poolMgr.BeginBlock(ctx)
	am.validatorMgr.BeginBlock(ctx, req.Header.Height)
	am.txOutStore.NewBlock(uint64(req.Header.Height))
}

func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	ctx.Logger().Debug("End Block", "height", req.Height)

	// Enable a pool every 50,000 blocks (3 days)
	if ctx.BlockHeight()%50000 == 0 {
		am.keeper.EnableAPool(ctx)
	}

	am.poolMgr.EndBlock(ctx, am.txOutStore)
	am.txOutStore.CommitBlock(ctx)
	return am.validatorMgr.EndBlock(ctx, req.Height)
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
