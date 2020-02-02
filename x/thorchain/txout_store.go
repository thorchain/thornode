package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

type TxOutStore interface {
	NewBlock(height int64, constAccessor constants.ConstantValues)
	CommitBlock(ctx sdk.Context)
	GetBlockOut() *TxOut
	ClearOutboundItems()
	GetOutboundItems() []*TxOutItem
	TryAddTxOutItem(ctx sdk.Context, toi *TxOutItem) (bool, error)
}

// GetTxOutStore will return an implementation of the txout store that
func GetTxOutStore(keeper Keeper, version semver.Version) (TxOutStore, error) {
	if version.GTE(semver.MustParse("0.1.0")) {
		return NewTxOutStorageV1(keeper), nil
	}
	return nil, errInvalidVersion
}
