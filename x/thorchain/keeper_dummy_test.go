package thorchain

import (
	"errors"
	"fmt"

	"github.com/blang/semver"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/tendermint/tendermint/libs/log"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

var (
	kaboom    = errors.New("Kaboom!!!")
	kaboomSdk = sdk.NewError(DefaultCodespace, 404, "kaboom!!!")
)

type KVStoreDummy struct{}

func (k KVStoreDummy) Cdc() *codec.Codec       { return makeTestCodec() }
func (k KVStoreDummy) Supply() supply.Keeper   { return supply.Keeper{} }
func (k KVStoreDummy) CoinKeeper() bank.Keeper { return bank.BaseKeeper{} }
func (k KVStoreDummy) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", ModuleName))
}

func (k KVStoreDummy) GetKey(_ sdk.Context, prefix dbPrefix, key string) string {
	return fmt.Sprintf("%s/1/%s", prefix, key)
}

func (k KVStoreDummy) GetRuneBalaceOfModule(ctx sdk.Context, moduleName string) sdk.Uint {
	return sdk.ZeroUint()
}

func (k KVStoreDummy) SendFromModuleToModule(ctx sdk.Context, from, to string, coin common.Coin) sdk.Error {
	return kaboomSdk
}

func (k KVStoreDummy) SendFromAccountToModule(ctx sdk.Context, from sdk.AccAddress, to string, coin common.Coin) sdk.Error {
	return kaboomSdk
}

func (k KVStoreDummy) SendFromModuleToAccount(ctx sdk.Context, from string, to sdk.AccAddress, coin common.Coin) sdk.Error {
	return kaboomSdk
}

func (k KVStoreDummy) SetLastSignedHeight(_ sdk.Context, _ int64) { return }
func (k KVStoreDummy) GetLastSignedHeight(_ sdk.Context) (int64, error) {
	return 0, kaboom
}

func (k KVStoreDummy) SetLastChainHeight(_ sdk.Context, _ common.Chain, _ int64) error {
	return kaboom
}

func (k KVStoreDummy) GetLastChainHeight(_ sdk.Context, _ common.Chain) (int64, error) {
	return 0, kaboom
}

func (k KVStoreDummy) GetRagnarokBlockHeight(_ sdk.Context) (int64, error) {
	return 0, kaboom
}
func (k KVStoreDummy) SetRagnarokBlockHeight(_ sdk.Context, _ int64) {}
func (k KVStoreDummy) RagnarokInProgress(_ sdk.Context) bool         { return false }
func (k KVStoreDummy) GetPoolBalances(_ sdk.Context, _, _ common.Asset) (sdk.Uint, sdk.Uint) {
	return sdk.ZeroUint(), sdk.ZeroUint()
}

func (k KVStoreDummy) GetPoolIterator(_ sdk.Context) sdk.Iterator {
	return nil
}
func (k KVStoreDummy) SetPoolData(_ sdk.Context, _ common.Asset, _ PoolStatus) {}
func (k KVStoreDummy) GetPoolDataIterator(_ sdk.Context) sdk.Iterator          { return nil }
func (k KVStoreDummy) EnableAPool(_ sdk.Context)                               {}

func (k KVStoreDummy) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return Pool{}, kaboom
}
func (k KVStoreDummy) GetPools(_ sdk.Context) (Pools, error)                        { return nil, kaboom }
func (k KVStoreDummy) SetPool(_ sdk.Context, _ Pool) error                          { return kaboom }
func (k KVStoreDummy) PoolExist(_ sdk.Context, _ common.Asset) bool                 { return false }
func (k KVStoreDummy) GetStakerIterator(_ sdk.Context, _ common.Asset) sdk.Iterator { return nil }
func (k KVStoreDummy) GetStaker(_ sdk.Context, _ common.Asset, _ common.Address) (Staker, error) {
	return Staker{}, kaboom
}
func (k KVStoreDummy) SetStaker(_ sdk.Context, _ Staker)                 {}
func (k KVStoreDummy) RemoveStaker(_ sdk.Context, _ Staker)              {}
func (k KVStoreDummy) TotalActiveNodeAccount(_ sdk.Context) (int, error) { return 0, kaboom }
func (k KVStoreDummy) ListNodeAccountsWithBond(_ sdk.Context) (NodeAccounts, error) {
	return nil, kaboom
}

func (k KVStoreDummy) ListNodeAccountsByStatus(_ sdk.Context, _ NodeStatus) (NodeAccounts, error) {
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
func (k KVStoreDummy) EnsureNodeKeysUnique(_ sdk.Context, _ string, _ common.PubKeySet) error {
	return kaboom
}
func (k KVStoreDummy) GetNodeAccountIterator(_ sdk.Context) sdk.Iterator { return nil }

func (k KVStoreDummy) GetNodeAccountSlashPoints(_ sdk.Context, _ sdk.AccAddress) (int64, error) {
	return 0, kaboom
}
func (k KVStoreDummy) SetNodeAccountSlashPoints(_ sdk.Context, _ sdk.AccAddress, _ int64) {}
func (k KVStoreDummy) ResetNodeAccountSlashPoints(_ sdk.Context, _ sdk.AccAddress)        {}
func (k KVStoreDummy) IncNodeAccountSlashPoints(_ sdk.Context, _ sdk.AccAddress, _ int64) error {
	return kaboom
}

func (k KVStoreDummy) DecNodeAccountSlashPoints(_ sdk.Context, _ sdk.AccAddress, _ int64) error {
	return kaboom
}
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

func (k KVStoreDummy) GetKeygenBlock(_ sdk.Context, _ int64) (KeygenBlock, error) {
	return KeygenBlock{}, kaboom
}
func (k KVStoreDummy) SetKeygenBlock(_ sdk.Context, _ KeygenBlock) error      { return kaboom }
func (k KVStoreDummy) GetKeygenBlockIterator(_ sdk.Context) sdk.Iterator      { return nil }
func (k KVStoreDummy) GetTxOut(_ sdk.Context, _ int64) (*TxOut, error)        { return nil, kaboom }
func (k KVStoreDummy) SetTxOut(_ sdk.Context, _ *TxOut) error                 { return kaboom }
func (k KVStoreDummy) AppendTxOut(_ sdk.Context, _ int64, _ *TxOutItem) error { return kaboom }
func (k KVStoreDummy) GetTxOutIterator(_ sdk.Context) sdk.Iterator            { return nil }
func (k KVStoreDummy) AddToLiquidityFees(_ sdk.Context, _ common.Asset, _ sdk.Uint) error {
	return kaboom
}

func (k KVStoreDummy) GetTotalLiquidityFees(_ sdk.Context, _ uint64) (sdk.Uint, error) {
	return sdk.ZeroUint(), kaboom
}

func (k KVStoreDummy) GetPoolLiquidityFees(_ sdk.Context, _ uint64, _ common.Asset) (sdk.Uint, error) {
	return sdk.ZeroUint(), kaboom
}

func (k KVStoreDummy) GetEvent(_ sdk.Context, _ int64) (Event, error) { return Event{}, kaboom }
func (k KVStoreDummy) GetEventsIterator(_ sdk.Context) sdk.Iterator   { return nil }
func (k KVStoreDummy) UpsertEvent(_ sdk.Context, _ Event) error       { return kaboom }
func (k KVStoreDummy) GetPendingEventID(_ sdk.Context, _ common.TxID) ([]int64, error) {
	return nil, kaboom
}

func (k KVStoreDummy) GetEventsIDByTxHash(ctx sdk.Context, txID common.TxID) ([]int64, error) {
	return nil, kaboom
}
func (k KVStoreDummy) GetCurrentEventID(_ sdk.Context) (int64, error)    { return 0, kaboom }
func (k KVStoreDummy) SetCurrentEventID(_ sdk.Context, _ int64)          {}
func (k KVStoreDummy) GetAllPendingEvents(_ sdk.Context) (Events, error) { return nil, kaboom }

func (k KVStoreDummy) GetChains(_ sdk.Context) (common.Chains, error)  { return nil, kaboom }
func (k KVStoreDummy) SetChains(_ sdk.Context, _ common.Chains)        {}
func (k KVStoreDummy) GetVaultIterator(_ sdk.Context) sdk.Iterator     { return nil }
func (k KVStoreDummy) VaultExists(_ sdk.Context, _ common.PubKey) bool { return false }
func (k KVStoreDummy) FindPubKeyOfAddress(_ sdk.Context, _ common.Address, _ common.Chain) (common.PubKey, error) {
	return common.EmptyPubKey, kaboom
}
func (k KVStoreDummy) SetVault(_ sdk.Context, _ Vault) error { return kaboom }
func (k KVStoreDummy) GetVault(_ sdk.Context, _ common.PubKey) (Vault, error) {
	return Vault{}, kaboom
}
func (k KVStoreDummy) GetAsgardVaults(_ sdk.Context) (Vaults, error) { return nil, kaboom }
func (k KVStoreDummy) GetAsgardVaultsByStatus(_ sdk.Context, _ VaultStatus) (Vaults, error) {
	return nil, kaboom
}
func (k KVStoreDummy) DeleteVault(_ sdk.Context, _ common.PubKey) error { return kaboom }

func (k KVStoreDummy) GetReservesContributors(_ sdk.Context) (ReserveContributors, error) {
	return nil, kaboom
}

func (k KVStoreDummy) SetReserveContributors(_ sdk.Context, _ ReserveContributors) error {
	return kaboom
}

func (k KVStoreDummy) HasValidVaultPools(_ sdk.Context) (bool, error)  { return false, kaboom }
func (k KVStoreDummy) AddFeeToReserve(_ sdk.Context, _ sdk.Uint) error { return kaboom }
func (k KVStoreDummy) GetVaultData(_ sdk.Context) (VaultData, error)   { return VaultData{}, kaboom }
func (k KVStoreDummy) SetVaultData(_ sdk.Context, _ VaultData) error   { return kaboom }
func (k KVStoreDummy) UpdateVaultData(_ sdk.Context, _ constants.ConstantValues, gasManager GasManager, manager EventManager) error {
	return kaboom
}

func (k KVStoreDummy) SetTssKeysignFailVoter(_ sdk.Context, tss TssKeysignFailVoter) {
}

func (k KVStoreDummy) GetTssKeysignFailVoterIterator(_ sdk.Context) sdk.Iterator {
	return nil
}

func (k KVStoreDummy) GetTssKeysignFailVoter(_ sdk.Context, _ string) (TssKeysignFailVoter, error) {
	return TssKeysignFailVoter{}, kaboom
}

func (k KVStoreDummy) GetGas(_ sdk.Context, _ common.Asset) ([]sdk.Uint, error) {
	return nil, kaboom
}
func (k KVStoreDummy) SetGas(_ sdk.Context, _ common.Asset, _ []sdk.Uint) {}
func (k KVStoreDummy) GetGasIterator(ctx sdk.Context) sdk.Iterator        { return nil }

func (k KVStoreDummy) ListTxMarker(_ sdk.Context, _ string) (TxMarkers, error) {
	return nil, kaboom
}
func (k KVStoreDummy) SetTxMarkers(_ sdk.Context, _ string, _ TxMarkers) error  { return kaboom }
func (k KVStoreDummy) AppendTxMarker(_ sdk.Context, _ string, _ TxMarker) error { return kaboom }

func (k KVStoreDummy) SetErrataTxVoter(_ sdk.Context, _ ErrataTxVoter)     {}
func (k KVStoreDummy) GetErrataTxVoterIterator(_ sdk.Context) sdk.Iterator { return nil }
func (k KVStoreDummy) GetErrataTxVoter(_ sdk.Context, _ common.TxID, _ common.Chain) (ErrataTxVoter, error) {
	return ErrataTxVoter{}, kaboom
}
func (k KVStoreDummy) SetBanVoter(_ sdk.Context, _ BanVoter) {}
func (k KVStoreDummy) GetBanVoter(_ sdk.Context, _ sdk.AccAddress) (BanVoter, error) {
	return BanVoter{}, kaboom
}
func (k KVStoreDummy) SetSwapQueueItem(ctx sdk.Context, msg MsgSwap) error { return kaboom }
func (k KVStoreDummy) GetSwapQueueIterator(ctx sdk.Context) sdk.Iterator   { return nil }
func (k KVStoreDummy) RemoveSwapQueueItem(ctx sdk.Context, _ common.TxID)  {}
func (k KVStoreDummy) GetSwapQueueItem(ctx sdk.Context, txID common.TxID) (MsgSwap, error) {
	return MsgSwap{}, kaboom
}
func (k KVStoreDummy) GetMimir(_ sdk.Context, key string) (int64, error) { return 0, kaboom }
func (k KVStoreDummy) SetMimir(_ sdk.Context, key string, value int64)   {}
func (k KVStoreDummy) GetMimirIterator(ctx sdk.Context) sdk.Iterator     { return nil }

// a mock sdk.Iterator implementation for testing purposes
type DummyIterator struct {
	sdk.Iterator
	placeholder int
	keys        [][]byte
	values      [][]byte
	err         error
}

func NewDummyIterator() *DummyIterator {
	return &DummyIterator{
		keys:   make([][]byte, 0),
		values: make([][]byte, 0),
	}
}

func (iter *DummyIterator) AddItem(key, value []byte) {
	iter.keys = append(iter.keys, key)
	iter.values = append(iter.values, value)
}

func (iter *DummyIterator) Next() {
	iter.placeholder++
}

func (iter *DummyIterator) Valid() bool {
	return iter.placeholder < len(iter.keys)
}

func (iter *DummyIterator) Key() []byte {
	return iter.keys[iter.placeholder]
}

func (iter *DummyIterator) Value() []byte {
	return iter.values[iter.placeholder]
}

func (iter *DummyIterator) Close() {
	iter.placeholder = 0
}

func (iter *DummyIterator) Error() error {
	return iter.err
}

func (iter *DummyIterator) Domain() (start, end []byte) {
	return nil, nil
}
