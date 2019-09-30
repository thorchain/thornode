package swapservice

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/statechain/cmd"
	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

type HandlerSuite struct{}

var _ = Suite(&HandlerSuite{})

func (s *HandlerSuite) SetUpSuite(c *C) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
}

// nolint: deadcode unused
// create a codec used only for testing
func makeTestCodec() *codec.Codec {
	var cdc = codec.New()

	bank.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

var (
	multiPerm      = "multiple permissions account"
	randomPerm     = "random permission"
	holder         = "holder"
	keySwapService = sdk.NewKVStoreKey(StoreKey)
)

func setupKeeperForTest(c *C) (sdk.Context, Keeper) {

	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySwapService, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
	c.Assert(err, IsNil)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "statechain"}, false, log.NewNopLogger())
	cdc := makeTestCodec()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	ak := auth.NewAccountKeeper(cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bk := bank.NewBaseKeeper(ak, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, nil)

	maccPerms := map[string][]string{
		auth.FeeCollectorName: nil,
		holder:                nil,
		supply.Minter:         []string{supply.Minter},
		supply.Burner:         []string{supply.Burner},
		multiPerm:             []string{supply.Minter, supply.Burner, supply.Staking},
		randomPerm:            []string{"random"},
		ModuleName:            {supply.Minter},
	}
	supplyKeeper := supply.NewKeeper(cdc, keySupply, ak, bk, maccPerms)
	totalSupply := sdk.NewCoins(sdk.NewCoin("bep", sdk.NewInt(1000*common.One)))
	supplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))
	k := NewKeeper(bk, supplyKeeper, keySwapService, cdc)
	return ctx, k
}

func (HandlerSuite) TestHandleMsgApply(c *C) {
	ctx, k := setupKeeperForTest(c)
	nodeAddr, err := sdk.AccAddressFromBech32("bep1munaxncdrr305vf0ljzydyqryncxx8cp06v65u")
	c.Check(err, IsNil)
	bond := sdk.NewUint(100)
	txID, err := common.NewTxID("33913271624A28F9EF4F53CA90BEABADF196E791EA2C281740159FEBE161B620")
	c.Check(err, IsNil)
	signer, err := sdk.AccAddressFromBech32("bep1a8amdesvfjtxamagh6hmrkcrej73rzas2jny7z")
	c.Check(err, IsNil)

	// Not Authorized
	msgApply := NewMsgApply(nodeAddr, bond, txID, signer)
	c.Assert(msgApply.ValidateBasic(), IsNil)
	result := handleMsgApply(ctx, k, msgApply)
	c.Assert(result.IsOK(), Equals, false)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	// add observer
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	bepConsPubKey := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdrg`
	nodeAccount := NewNodeAccount(signer, NodeActive, NewTrustAccount(bnb, signer, bepConsPubKey))
	k.SetNodeAccount(ctx, nodeAccount)
	// nodeAccoutn already exist
	result = handleMsgApply(ctx, k, msgApply)
	c.Assert(result.IsOK(), Equals, false)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)

	// invalid Msg
	invalidMsgApply := NewMsgApply(sdk.AccAddress{}, bond, txID, signer)
	invalidMsgApplyResult := handleMsgApply(ctx, k, invalidMsgApply)
	c.Assert(invalidMsgApplyResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Assert(invalidMsgApplyResult.IsOK(), Equals, false)

	// less than minimum bond
	msgApplyLessThanMinimumBond := NewMsgApply(nodeAddr, sdk.NewUint(1000), txID, signer)
	lessThanMinimumBondResult := handleMsgApply(ctx, k, msgApplyLessThanMinimumBond)
	c.Assert(lessThanMinimumBondResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Assert(lessThanMinimumBondResult.IsOK(), Equals, false)

	msgApply1 := NewMsgApply(nodeAddr, sdk.NewUint(100*common.One), txID, signer)
	result = handleMsgApply(ctx, k, msgApply1)
	c.Assert(result.IsOK(), Equals, true)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	coins := k.coinKeeper.GetCoins(ctx, nodeAddr)
	c.Assert(coins.AmountOf("bep").Int64(), Equals, int64(1000))

	// apply again shohuld fail
	resultExist := handleMsgApply(ctx, k, msgApply1)
	c.Assert(resultExist.IsOK(), Equals, false)
	c.Assert(resultExist.Code, Equals, sdk.CodeUnknownRequest)

}

func (HandlerSuite) TestHandleMsgSetTrustAccount(c *C) {
	ctx, k := setupKeeperForTest(c)
	nodeAddr, err := sdk.AccAddressFromBech32("bep1munaxncdrr305vf0ljzydyqryncxx8cp06v65u")
	c.Check(err, IsNil)

	signer, err := sdk.AccAddressFromBech32("bep1a8amdesvfjtxamagh6hmrkcrej73rzas2jny7z")
	c.Check(err, IsNil)
	// add observer
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	bepConsPubKey := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdrg`

	trustAccount := NewTrustAccount(bnb, signer, bepConsPubKey)
	msgTrustAccount := types.NewMsgSetTrustAccount(trustAccount, signer)
	unAuthorizedResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(unAuthorizedResult.Code, Equals, sdk.CodeUnauthorized)
	c.Check(unAuthorizedResult.IsOK(), Equals, false)
	nodeAccount := NewNodeAccount(signer, NodeActive, NewTrustAccount(common.NoBnbAddress, sdk.AccAddress{}, ""))
	k.SetNodeAccount(ctx, nodeAccount)

	activeFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(activeFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(activeFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeDisabled, NewTrustAccount(common.NoBnbAddress, sdk.AccAddress{}, ""))
	k.SetNodeAccount(ctx, nodeAccount)

	disabledFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(disabledFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(disabledFailResult.IsOK(), Equals, false)

	k.SetNodeAccount(ctx, NewNodeAccount(signer, NodeWhiteListed, NewTrustAccount(bnb, nodeAddr, bepConsPubKey)))

	notUniqueFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(notUniqueFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(notUniqueFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeWhiteListed, NewTrustAccount(common.NoBnbAddress, sdk.AccAddress{}, ""))
	k.SetNodeAccount(ctx, nodeAccount)

	success := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(success.Code, Equals, sdk.CodeOK)
	c.Check(success.IsOK(), Equals, true)

}
