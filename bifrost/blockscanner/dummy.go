package blockscanner

import "gitlab.com/thorchain/thornode/bifrost/thorclient/types"

type DummyFetcher struct {
	Tx  types.TxIn
	Err error
}

func NewDummyFetcher(tx types.TxIn, err error) DummyFetcher {
	return DummyFetcher{
		Tx:  tx,
		Err: err,
	}
}

func (d DummyFetcher) FetchTxs(height int64) (types.TxIn, error) {
	return d.Tx, d.Err
}

func (d DummyFetcher) FetchLastHeight() (int64, error) {
	return 0, nil
}
