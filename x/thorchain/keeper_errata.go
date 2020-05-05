package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperErrataTx interface {
	SetErrataTxVoter(_ sdk.Context, _ ErrataTxVoter)
	GetErrataTxVoterIterator(_ sdk.Context) sdk.Iterator
	GetErrataTxVoter(_ sdk.Context, _ common.TxID, _ common.Chain) (ErrataTxVoter, error)
}

// SetErrataTxVoter - save a txin voter object
func (k KVStore) SetErrataTxVoter(ctx sdk.Context, errata ErrataTxVoter) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixErrataTx, errata.String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(errata))
}

// GetErrataTxVoterIterator iterate tx in voters
func (k KVStore) GetErrataTxVoterIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixErrataTx))
}

// GetErrataTx - gets information of a tx hash
func (k KVStore) GetErrataTxVoter(ctx sdk.Context, txID common.TxID, chain common.Chain) (ErrataTxVoter, error) {
	errata := NewErrataTxVoter(txID, chain)
	key := k.GetKey(ctx, prefixErrataTx, errata.String())

	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return errata, nil
	}

	bz := store.Get([]byte(key))
	var record ErrataTxVoter
	if err := k.cdc.UnmarshalBinaryBare(bz, &record); err != nil {
		return errata, dbError(ctx, "Unmarshal: errata tx voter", err)
	}
	return record, nil
}
