package exchange

import (
	"github.com/binance-chain/go-sdk/keys"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/jpthor/cosmos-swap/storage"
)

// Wallets manage wallets for different symbol
type Wallets struct {
	ds     storage.GetAndPutable
	logger zerolog.Logger
}

// NewWallets create a new instance of Wallets
func NewWallets(ds storage.GetAndPutable, logger zerolog.Logger) (*Wallets, error) {
	if nil == ds {
		return nil, errors.New("ds is nil")
	}

	return &Wallets{
		ds:     ds,
		logger: logger.With().Str("module", "wallets").Logger(),
	}, nil
}

func newWallet(assetSymbol string) (*Bep2Wallet, error) {

	km, err := keys.NewKeyManager()
	if nil != err {
		return nil, errors.Wrapf(err, "fail to create key manager")
	}
	pk, err := km.ExportAsPrivateKey()
	if nil != err {
		return nil, errors.Wrap(err, "fail to export private key")
	}
	n, err := km.ExportAsMnemonic()
	if nil != err {
		return nil, errors.Wrap(err, "fail to export Mnemonic")
	}
	return &Bep2Wallet{
		AssetSymbol:   assetSymbol,
		PrivateKey:    pk,
		Mnemonic:      n,
		PublicAddress: km.GetAddr().String(),
	}, nil
}

// GetWallet return a instance of Bep2Wallet , check the datastore if it exist then just return otherwise create a new one
func (w *Wallets) GetWallet(assetSymbol string) (*Bep2Wallet, error) {
	if len(assetSymbol) == 0 {
		return nil, errors.New("assetSymbol is empty")
	}
	value, err := w.ds.Get([]byte(assetSymbol))
	if nil != err {
		if err != leveldb.ErrNotFound {
			return nil, errors.Wrapf(err, "fail to get from wallet from data store, asset symbol: %s", assetSymbol)
		}
		// create a new one
		wallet, err := newWallet(assetSymbol)
		if nil != err {
			return nil, errors.Wrapf(err, "fail to create a new wallet for %s", assetSymbol)
		}
		buf, err := wallet.ToBytes()
		if nil != err {
			return nil, errors.Wrapf(err, "fail to serialize wallet to bytes")
		}
		if err := w.ds.Put([]byte(assetSymbol), buf); nil != err {
			return nil, errors.Wrapf(err, "fail to save wallet to data store, asset symbol:%s", assetSymbol)
		}
		return wallet, nil
	}
	return FromBytes(value)
}
