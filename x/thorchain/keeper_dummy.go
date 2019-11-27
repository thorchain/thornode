package thorchain

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"

	"gitlab.com/thorchain/thornode/common"
)

var kaboom = errors.New("Kaboom!!!")

type KVStoreDummy struct{}

func (k KVStoreDummy) Cdc() *codec.Codec       { return codec.New() }
func (k KVStoreDummy) Supply() supply.Keeper   { return supply.Keeper{} }
func (k KVStoreDummy) CoinKeeper() bank.Keeper { return bank.BaseKeeper{} }
func (k KVStoreDummy) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", ModuleName))
}
func (k KVStoreDummy) SetLastSignedHeight(ctx sdk.Context, height sdk.Uint) { return }
func (k KVStoreDummy) GetLastSignedHeight(ctx sdk.Context) sdk.Uint         { return sdk.ZeroUint() }
func (k KVStoreDummy) SetLastChainHeight(ctx sdk.Context, chain common.Chain, height sdk.Uint) error {
	return kaboom
}
func (k KVStoreDummy) GetLastChainHeight(ctx sdk.Context, chain common.Chain) sdk.Uint {
	return sdk.ZeroUint()
}
func (k KVStoreDummy) GetPool(ctx sdk.Context, asset common.Asset) Pool { return Pool{} }
func (k KVStoreDummy) SetPool(ctx sdk.Context, pool Pool)               {}
func (k KVStoreDummy) GetPoolBalances(ctx sdk.Context, asset, asset2 common.Asset) (sdk.Uint, sdk.Uint) {
	return sdk.ZeroUint(), sdk.ZeroUint()
}
func (k KVStoreDummy) SetPoolData(ctx sdk.Context, asset common.Asset, ps PoolStatus) {}
func (k KVStoreDummy) GetPoolDataIterator(ctx sdk.Context) sdk.Iterator               { return nil }
func (k KVStoreDummy) EnableAPool(ctx sdk.Context)                                    {}
func (k KVStoreDummy) PoolExist(ctx sdk.Context, asset common.Asset) bool             { return false }
func (k KVStoreDummy) GetPoolIndex(ctx sdk.Context) (PoolIndex, error)                { return nil, kaboom }
func (k KVStoreDummy) SetPoolIndex(ctx sdk.Context, pi PoolIndex)                     {}
func (k KVStoreDummy) AddToPoolIndex(ctx sdk.Context, asset common.Asset) error       { return kaboom }
func (k KVStoreDummy) RemoveFromPoolIndex(ctx sdk.Context, asset common.Asset) error  { return kaboom }
func (k KVStoreDummy) GetPoolStakerIterator(ctx sdk.Context) sdk.Iterator             { return nil }
func (k KVStoreDummy) GetPoolStaker(ctx sdk.Context, asset common.Asset) (PoolStaker, error) {
	return PoolStaker{}, kaboom
}
func (k KVStoreDummy) SetPoolStaker(ctx sdk.Context, asset common.Asset, ps PoolStaker) {}
func (k KVStoreDummy) GetStakerPoolIterator(ctx sdk.Context) sdk.Iterator               { return nil }
func (k KVStoreDummy) GetStakerPool(ctx sdk.Context, stakerID common.Address) (StakerPool, error) {
	return StakerPool{}, kaboom
}
func (k KVStoreDummy) SetStakerPool(ctx sdk.Context, stakerID common.Address, sp StakerPool) {}
func (k KVStoreDummy) TotalNodeAccounts(ctx sdk.Context) int                                 { return 0 }
func (k KVStoreDummy) TotalActiveNodeAccount(ctx sdk.Context) (int, error)                   { return 0, kaboom }
func (k KVStoreDummy) ListNodeAccounts(ctx sdk.Context) (NodeAccounts, error)                { return nil, kaboom }
func (k KVStoreDummy) ListNodeAccountsByStatus(ctx sdk.Context, status NodeStatus) (NodeAccounts, error) {
	return nil, kaboom
}
func (k KVStoreDummy) ListActiveNodeAccounts(ctx sdk.Context) (NodeAccounts, error) {
	return nil, kaboom
}
func (k KVStoreDummy) GetLowestActiveVersion(ctx sdk.Context) int64                { return 0 }
func (k KVStoreDummy) IsWhitelistedNode(ctx sdk.Context, addr sdk.AccAddress) bool { return false }
func (k KVStoreDummy) GetNodeAccount(ctx sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	return NodeAccount{}, kaboom
}
func (k KVStoreDummy) GetNodeAccountByPubKey(ctx sdk.Context, pk common.PubKey) (NodeAccount, error) {
	return NodeAccount{}, kaboom
}
func (k KVStoreDummy) GetNodeAccountByBondAddress(ctx sdk.Context, addr common.Address) (NodeAccount, error) {
	return NodeAccount{}, kaboom
}
func (k KVStoreDummy) SetNodeAccount(ctx sdk.Context, na NodeAccount)                        {}
func (k KVStoreDummy) SlashNodeAccountBond(ctx sdk.Context, na *NodeAccount, slash sdk.Uint) {}
func (k KVStoreDummy) SlashNodeAccountRewards(ctx sdk.Context, na *NodeAccount, pts int64)   {}
func (k KVStoreDummy) EnsureTrustAccountUnique(ctx sdk.Context, consensusPubKey string, pubKeys common.PubKeys) error {
	return kaboom
}
func (k KVStoreDummy) GetNodeAccountIterator(ctx sdk.Context) sdk.Iterator                 { return nil }
func (k KVStoreDummy) SetActiveObserver(ctx sdk.Context, addr sdk.AccAddress)              {}
func (k KVStoreDummy) RemoveActiveObserver(ctx sdk.Context, addr sdk.AccAddress)           {}
func (k KVStoreDummy) IsActiveObserver(ctx sdk.Context, addr sdk.AccAddress) bool          { return false }
func (k KVStoreDummy) GetObservingAddresses(ctx sdk.Context) []sdk.AccAddress              { return nil }
func (k KVStoreDummy) AddObservingAddresses(ctx sdk.Context, inAddresses []sdk.AccAddress) {}
func (k KVStoreDummy) ClearObservingAddresses(ctx sdk.Context)                             {}
func (k KVStoreDummy) SetTxInVoter(ctx sdk.Context, tx TxInVoter)                          {}
func (k KVStoreDummy) GetTxInVoterIterator(ctx sdk.Context) sdk.Iterator                   { return nil }
func (k KVStoreDummy) GetTxInVoter(ctx sdk.Context, hash common.TxID) TxInVoter            { return TxInVoter{} }
func (k KVStoreDummy) CheckTxHash(ctx sdk.Context, hash common.TxID) bool                  { return false }
func (k KVStoreDummy) GetTxInIndexIterator(ctx sdk.Context) sdk.Iterator                   { return nil }
func (k KVStoreDummy) GetTxInIndex(ctx sdk.Context, height uint64) (TxInIndex, error) {
	return TxInIndex{}, kaboom
}
func (k KVStoreDummy) SetTxInIndex(ctx sdk.Context, height uint64, index TxInIndex) {}
func (k KVStoreDummy) AddToTxInIndex(ctx sdk.Context, height uint64, id common.TxID) error {
	return kaboom
}
func (k KVStoreDummy) SetTxOut(ctx sdk.Context, blockOut *TxOut)               {}
func (k KVStoreDummy) GetTxOutIterator(ctx sdk.Context) sdk.Iterator           { return nil }
func (k KVStoreDummy) GetTxOut(ctx sdk.Context, height uint64) (*TxOut, error) { return nil, kaboom }
func (k KVStoreDummy) AddToLiquidityFees(ctx sdk.Context, pool Pool, fee sdk.Uint) error {
	return kaboom
}
func (k KVStoreDummy) getLiquidityFees(ctx sdk.Context, height uint64, prefix dbPrefix) (sdk.Uint, error) {
	return sdk.ZeroUint(), kaboom
}
func (k KVStoreDummy) GetTotalLiquidityFees(ctx sdk.Context, height uint64) (sdk.Uint, error) {
	return sdk.ZeroUint(), kaboom
}
func (k KVStoreDummy) GetPoolLiquidityFees(ctx sdk.Context, height uint64, pool Pool) (sdk.Uint, error) {
	return sdk.ZeroUint(), kaboom
}
func (k KVStoreDummy) GetIncompleteEvents(ctx sdk.Context) (Events, error)   { return nil, kaboom }
func (k KVStoreDummy) SetIncompleteEvents(ctx sdk.Context, events Events)    {}
func (k KVStoreDummy) AddIncompleteEvents(ctx sdk.Context, event Event)      {}
func (k KVStoreDummy) GetCompleteEventIterator(ctx sdk.Context) sdk.Iterator { return nil }
func (k KVStoreDummy) GetCompletedEvent(ctx sdk.Context, id int64) (Event, error) {
	return Event{}, kaboom
}
func (k KVStoreDummy) SetCompletedEvent(ctx sdk.Context, event Event)                  {}
func (k KVStoreDummy) CompleteEvents(ctx sdk.Context, in []common.TxID, out common.Tx) {}
func (k KVStoreDummy) GetLastEventID(ctx sdk.Context) int64                            { return 0 }
func (k KVStoreDummy) SetLastEventID(ctx sdk.Context, id int64)                        {}
func (k KVStoreDummy) SetPoolAddresses(ctx sdk.Context, addresses *PoolAddresses)      {}
func (k KVStoreDummy) GetPoolAddresses(ctx sdk.Context) PoolAddresses                  { return PoolAddresses{} }
func (k KVStoreDummy) SetValidatorMeta(ctx sdk.Context, meta ValidatorMeta)            {}
func (k KVStoreDummy) GetValidatorMeta(ctx sdk.Context) ValidatorMeta                  { return ValidatorMeta{} }
func (k KVStoreDummy) GetChains(ctx sdk.Context) common.Chains                         { return nil }
func (k KVStoreDummy) SupportedChain(ctx sdk.Context, chain common.Chain) bool         { return false }
func (k KVStoreDummy) AddChain(ctx sdk.Context, chain common.Chain)                    {}
func (k KVStoreDummy) GetYggdrasilIterator(ctx sdk.Context) sdk.Iterator               { return nil }
func (k KVStoreDummy) YggdrasilExists(ctx sdk.Context, pk common.PubKey) bool          { return false }
func (k KVStoreDummy) FindPubKeyOfAddress(ctx sdk.Context, addr common.Address, chain common.Chain) (common.PubKey, error) {
	return common.EmptyPubKey, kaboom
}
func (k KVStoreDummy) SetYggdrasil(ctx sdk.Context, ygg Yggdrasil)                          {}
func (k KVStoreDummy) GetYggdrasil(ctx sdk.Context, pk common.PubKey) Yggdrasil             { return Yggdrasil{} }
func (k KVStoreDummy) HasValidYggdrasilPools(ctx sdk.Context) (bool, error)                 { return false, kaboom }
func (k KVStoreDummy) GetReservesContributors(ctx sdk.Context) ReserveContributors          { return nil }
func (k KVStoreDummy) SetReserveContributors(ctx sdk.Context, contribs ReserveContributors) {}
func (k KVStoreDummy) GetVaultData(ctx sdk.Context) VaultData                               { return VaultData{} }
func (k KVStoreDummy) SetVaultData(ctx sdk.Context, data VaultData)                         {}
func (k KVStoreDummy) UpdateVaultData(ctx sdk.Context)                                      {}
func (k KVStoreDummy) SetAdminConfig(ctx sdk.Context, config AdminConfig)                   {}
func (k KVStoreDummy) GetAdminConfigDefaultPoolStatus(ctx sdk.Context, addr sdk.AccAddress) PoolStatus {
	return PoolSuspended
}
func (k KVStoreDummy) GetAdminConfigGSL(ctx sdk.Context, addr sdk.AccAddress) common.Amount {
	return common.ZeroAmount
}
func (k KVStoreDummy) GetAdminConfigStakerAmtInterval(ctx sdk.Context, addr sdk.AccAddress) common.Amount {
	return common.ZeroAmount
}
func (k KVStoreDummy) GetAdminConfigMinValidatorBond(ctx sdk.Context, addr sdk.AccAddress) sdk.Uint {
	return sdk.ZeroUint()
}
func (k KVStoreDummy) GetAdminConfigWhiteListGasAsset(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return nil
}
func (k KVStoreDummy) GetAdminConfigBnbAddressType(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) common.Address {
	return common.NoAddress
}
func (k KVStoreDummy) GetAdminConfigUintType(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) sdk.Uint {
	return sdk.ZeroUint()
}
func (k KVStoreDummy) GetAdminConfigAmountType(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) common.Amount {
	return common.ZeroAmount
}
func (k KVStoreDummy) GetAdminConfigCoinsType(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) sdk.Coins {
	return nil
}
func (k KVStoreDummy) GetAdminConfigInt64(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) int64 {
	return 0
}
func (k KVStoreDummy) GetAdminConfigValue(ctx sdk.Context, kkey AdminConfigKey, addr sdk.AccAddress) (val string, err error) {
	return "", kaboom
}
func (k KVStoreDummy) GetAdminConfigIterator(ctx sdk.Context) sdk.Iterator { return nil }
