package bitcoin

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

// UnspentTransactionOutputAccessor define methods to access bitcoin unspent transactional output
type UnspentTransactionOutputAccessor interface {
	GetUTXOs() ([]UnspentTransactionOutput, error)
	AddUTXO(UnspentTransactionOutput) error
	RemoveUTXO(key string) error
}

// UTXOAccessor defines struct to hold UnspentTransactionOutput interface
type UTXOAccessor struct {
	*LevelDBUTXOAccessor
	db *leveldb.DB
}

// NewUTXOAccessor creates new utxo object
func NewUTXOAccessor(levelDbFolder string) (*UTXOAccessor, error) {
	var err error
	var db *leveldb.DB
	if len(levelDbFolder) == 0 {
		// no directory given, use in memory store
		storage := storage.NewMemStorage()
		db, err = leveldb.Open(storage, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to in memory open level db")
		}
	} else {
		db, err = leveldb.OpenFile(levelDbFolder, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to open level db %s", levelDbFolder)
		}
	}
	levelDbUTXO, err := NewLevelDBUTXOAccessor(db)
	if err != nil {
		return nil, errors.New("fail to create level db")
	}
	return &UTXOAccessor{
		LevelDBUTXOAccessor: levelDbUTXO,
		db:                  db,
	}, nil
}
