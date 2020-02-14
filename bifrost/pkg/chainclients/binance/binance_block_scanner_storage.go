package binance

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
)

const DefaultObserverLevelDBFolder = `observer_data`

type BinanceBlockScannerStorage struct {
	*blockscanner.LevelDBScannerStorage
	db *leveldb.DB
}

func NewBinanceBlockScannerStorage(levelDbFolder string) (*BinanceBlockScannerStorage, error) {
	if len(levelDbFolder) == 0 {
		levelDbFolder = DefaultObserverLevelDBFolder
	}
	db, err := leveldb.OpenFile(levelDbFolder, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to open level db %s", levelDbFolder)
	}
	levelDbStorage, err := blockscanner.NewLevelDBScannerStorage(db)
	if err != nil {
		return nil, errors.New("fail to create leven db")
	}
	return &BinanceBlockScannerStorage{
		LevelDBScannerStorage: levelDbStorage,
		db:                    db,
	}, nil
}

// Close underlying db
func (ldbss *BinanceBlockScannerStorage) Close() error {
	return ldbss.db.Close()
}
