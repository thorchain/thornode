package thorchain

import (
	"fmt"

	"github.com/blang/semver"
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
func (k KVStoreDummy) GetKey(_ sdk.Context, prefix dbPrefix, key string) string {
	return fmt.Sprintf("%s/1/%s", prefix, key)
}
func (k KVStoreDummy) SetLastSignedHeight(_ sdk.Context, _ sdk.Uint) { return }
func (k KVStoreDummy) GetLastSignedHeight(_ sdk.Context) (sdk.Uint, error) {
	return sdk.ZeroUint(), kaboom
}
func (k KVStoreDummy) SetLastChainHeight(_ sdk.Context, _ common.Chain, _ sdk.Uint) error {
	return kaboom
}
func (k KVStoreDummy) GetLastChainHeight(_ sdk.Context, _ common.Chain) (sdk.Uint, error) {
	return sdk.ZeroUint(), kaboom
}
func (k KVStoreDummy) GetPoolBalances(_ sdk.Context, _, _ common.Asset) (sdk.Uint, sdk.Uint) {
	return sdk.ZeroUint(), sdk.ZeroUint()
}
func (k KVStoreDummy) GetPoolIterator(_ sdk.Context) sdk.Iterator {
	return nil
}
func (k KVStoreDummy) SetPoolData(_ sdk.Context, _ common.Asset, _ PoolStatus) {}
func (k KVStoreDummy) GetPoolDataIterator(_ sdk.Context) sdk.Iterator          { return nil }
func (k KVStoreDummy) EnableAPool(_ sdk.Context)                               {}
func (k KVStoreDummy) GetPoolIndex(_ sdk.Context) (PoolIndex, error)           { return nil, kaboom }
func (k KVStoreDummy) SetPoolIndex(_ sdk.Context, _ PoolIndex)                 {}
func (k KVStoreDummy) AddToPoolIndex(_ sdk.Context, _ common.Asset) error      { return kaboom }
func (k KVStoreDummy) RemoveFromPoolIndex(_ sdk.Context, _ common.Asset) error { return kaboom }

func (k KVStoreDummy) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return Pool{}, kaboom
}
func (k KVStoreDummy) GetPools(_ sdk.Context) (Pools, error)            { return nil, kaboom }
func (k KVStoreDummy) SetPool(_ sdk.Context, _ Pool) error              { return kaboom }
func (k KVStoreDummy) PoolExist(_ sdk.Context, _ common.Asset) bool     { return false }
func (k KVStoreDummy) GetPoolStakerIterator(_ sdk.Context) sdk.Iterator { return nil }
func (k KVStoreDummy) GetPoolStaker(_ sdk.Context, _ common.Asset) (PoolStaker, error) {
	return PoolStaker{}, kaboom
}
func (k KVStoreDummy) SetPoolStaker(_ sdk.Context, _ PoolStaker)        {}
func (k KVStoreDummy) GetStakerPoolIterator(_ sdk.Context) sdk.Iterator { return nil }
func (k KVStoreDummy) GetStakerPool(_ sdk.Context, _ common.Address) (StakerPool, error) {
	return StakerPool{}, kaboom
}
func (k KVStoreDummy) SetStakerPool(ctx sdk.Context, sp StakerPool)           {}
func (k KVStoreDummy) TotalActiveNodeAccount(ctx sdk.Context) (int, error)    { return 0, kaboom }
func (k KVStoreDummy) ListNodeAccounts(ctx sdk.Context) (NodeAccounts, error) { return nil, kaboom }
func (k KVStoreDummy) ListNodeAccountsByStatus(ctx sdk.Context, status NodeStatus) (NodeAccounts, error) {
	return nil, kaboom
}
func (k KVStoreDummy) ListActiveNodeAccounts(_ sdk.Context) (NodeAccounts, error) {
	return nil, kaboom
}
func (k KVStoreDummy) GetLowestActiveVersion(_ sdk.Context) semver.Version { return semver.Version{} }
func (k KVStoreDummy) GetMinJoinVersion(_ sdk.Context) semver.Version      { return semver.Version{} }
func (k KVStoreDummy) GetNodeAccount(_ sdk.Context, _ sdk.AccAddress) (NodeAccount, error) {
	return NodeAccount{}, kaboom
}
func (k KVStoreDummy) GetNodeAccountByPubKey(_ sdk.Context, _ common.PubKey) (NodeAccount, error) {
	return NodeAccount{}, kaboom
}
func (k KVStoreDummy) GetNodeAccountByBondAddress(_ sdk.Context, _ common.Address) (NodeAccount, error) {
	return NodeAccount{}, kaboom
}
func (k KVStoreDummy) SetNodeAccount(_ sdk.Context, _ NodeAccount) error { return kaboom }
func (k KVStoreDummy) EnsureTrustAccountUnique(_ sdk.Context, _ string, _ common.PubKeys) error {
	return kaboom
}
func (k KVStoreDummy) GetNodeAccountIterator(_ sdk.Context) sdk.Iterator     { return nil }
func (k KVStoreDummy) SetActiveObserver(_ sdk.Context, _ sdk.AccAddress)     {}
func (k KVStoreDummy) RemoveActiveObserver(_ sdk.Context, _ sdk.AccAddress)  {}
func (k KVStoreDummy) IsActiveObserver(_ sdk.Context, _ sdk.AccAddress) bool { return false }
func (k KVStoreDummy) GetObservingAddresses(_ sdk.Context) ([]sdk.AccAddress, error) {
	return nil, kaboom
}
func (k KVStoreDummy) AddObservingAddresses(_ sdk.Context, _ []sdk.AccAddress) error { return kaboom }
func (k KVStoreDummy) ClearObservingAddresses(_ sdk.Context)                         {}
func (k KVStoreDummy) SetObservedTxVoter(_ sdk.Context, _ ObservedTxVoter)           {}
func (k KVStoreDummy) GetObservedTxVoterIterator(_ sdk.Context) sdk.Iterator         { return nil }
func (k KVStoreDummy) GetObservedTxVoter(_ sdk.Context, _ common.TxID) (ObservedTxVoter, error) {
	return ObservedTxVoter{}, kaboom
}
func (k KVStoreDummy) SetTssVoter(_ sdk.Context, _ TssVoter)          {}
func (k KVStoreDummy) GetTssVoterIterator(_ sdk.Context) sdk.Iterator { return nil }
func (k KVStoreDummy) GetTssVoter(_ sdk.Context, _ string) (TssVoter, error) {
	return TssVoter{}, kaboom
}

func (k KVStoreDummy) GetTxOut(_ sdk.Context, _ uint64) (*TxOut, error) { return nil, kaboom }
func (k KVStoreDummy) SetTxOut(_ sdk.Context, _ *TxOut) error           { return kaboom }
func (k KVStoreDummy) GetTxOutIterator(_ sdk.Context) sdk.Iterator      { return nil }
func (k KVStoreDummy) AddToLiquidityFees(_ sdk.Context, _ common.Asset, _ sdk.Uint) error {
	return kaboom
}
func (k KVStoreDummy) GetTotalLiquidityFees(_ sdk.Context, _ uint64) (sdk.Uint, error) {
	return sdk.ZeroUint(), kaboom
}
func (k KVStoreDummy) GetPoolLiquidityFees(_ sdk.Context, _ uint64, _ common.Asset) (sdk.Uint, error) {
	return sdk.ZeroUint(), kaboom
}
func (k KVStoreDummy) GetIncompleteEvents(_ sdk.Context) (Events, error)   { return nil, kaboom }
func (k KVStoreDummy) SetIncompleteEvents(_ sdk.Context, _ Events)         {}
func (k KVStoreDummy) AddIncompleteEvents(_ sdk.Context, _ Event) error    { return kaboom }
func (k KVStoreDummy) GetCompleteEventIterator(_ sdk.Context) sdk.Iterator { return nil }
func (k KVStoreDummy) GetCompletedEvent(_ sdk.Context, _ int64) (Event, error) {
	return Event{}, kaboom
}
func (k KVStoreDummy) SetCompletedEvent(_ sdk.Context, _ Event)         {}
func (k KVStoreDummy) GetLastEventID(_ sdk.Context) (int64, error)      { return 0, kaboom }
func (k KVStoreDummy) SetLastEventID(_ sdk.Context, _ int64)            {}
func (k KVStoreDummy) SetPoolAddresses(_ sdk.Context, _ *PoolAddresses) {}
func (k KVStoreDummy) GetPoolAddresses(_ sdk.Context) (PoolAddresses, error) {
	return PoolAddresses{}, kaboom
}
func (k KVStoreDummy) GetChains(_ sdk.Context) (common.Chains, error)      { return nil, kaboom }
func (k KVStoreDummy) SetChains(_ sdk.Context, _ common.Chains)            {}
func (k KVStoreDummy) GetYggdrasilIterator(_ sdk.Context) sdk.Iterator     { return nil }
func (k KVStoreDummy) YggdrasilExists(_ sdk.Context, _ common.PubKey) bool { return false }
func (k KVStoreDummy) FindPubKeyOfAddress(_ sdk.Context, _ common.Address, _ common.Chain) (common.PubKey, error) {
	return common.EmptyPubKey, kaboom
}
func (k KVStoreDummy) SetYggdrasil(_ sdk.Context, _ Yggdrasil) error { return kaboom }
func (k KVStoreDummy) GetYggdrasil(_ sdk.Context, _ common.PubKey) (Yggdrasil, error) {
	return Yggdrasil{}, kaboom
}
func (k KVStoreDummy) GetReservesContributors(_ sdk.Context) (ReserveContributors, error) {
	return nil, kaboom
}
func (k KVStoreDummy) SetReserveContributors(_ sdk.Context, _ ReserveContributors) error {
	return kaboom
}

func (k KVStoreDummy) HasValidYggdrasilPools(_ sdk.Context) (bool, error) { return false, kaboom }
func (k KVStoreDummy) AddFeeToReserve(_ sdk.Context, _ sdk.Uint) error    { return kaboom }
func (k KVStoreDummy) GetVaultData(_ sdk.Context) (VaultData, error)      { return VaultData{}, kaboom }
func (k KVStoreDummy) SetVaultData(_ sdk.Context, _ VaultData) error      { return kaboom }
func (k KVStoreDummy) UpdateVaultData(_ sdk.Context) error                { return kaboom }
func (k KVStoreDummy) SetAdminConfig(_ sdk.Context, _ AdminConfig)        {}
func (k KVStoreDummy) GetAdminConfigDefaultPoolStatus(_ sdk.Context, _ sdk.AccAddress) PoolStatus {
	return PoolSuspended
}
func (k KVStoreDummy) GetAdminConfigWhiteListGasAsset(_ sdk.Context, _ sdk.AccAddress) sdk.Coins {
	return nil
}
func (k KVStoreDummy) GetAdminConfigBnbAddressType(_ sdk.Context, _ AdminConfigKey, _ string, _ sdk.AccAddress) common.Address {
	return common.NoAddress
}
func (k KVStoreDummy) GetAdminConfigUintType(_ sdk.Context, _ AdminConfigKey, _ string, _ sdk.AccAddress) sdk.Uint {
	return sdk.ZeroUint()
}
func (k KVStoreDummy) GetAdminConfigCoinsType(_ sdk.Context, _ AdminConfigKey, _ string, _ sdk.AccAddress) sdk.Coins {
	return nil
}
func (k KVStoreDummy) GetAdminConfigInt64(_ sdk.Context, _ AdminConfigKey, _ string, _ sdk.AccAddress) int64 {
	return 0
}
func (k KVStoreDummy) GetAdminConfigValue(_ sdk.Context, _ AdminConfigKey, _ sdk.AccAddress) (val string, err error) {
	return "", kaboom
}
func (k KVStoreDummy) GetAdminConfigIterator(_ sdk.Context) sdk.Iterator { return nil }
