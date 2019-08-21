package signer

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"

	"gitlab.com/thorchain/bepswap/observe/x/blockscanner"
)

const DefaultSignerLevelDBFolder = `signer_data`

type StateChanBlockScannerStorage struct {
	*blockscanner.LevelDBScannerStorage
	db *leveldb.DB
}

// NewStateChanBlockScannerStorage create a new instance of StateChanBlockScannerStorage
func NewStateChanBlockScannerStorage(levelDbFolder string) (*StateChanBlockScannerStorage, error) {
	if len(levelDbFolder) == 0 {
		levelDbFolder = DefaultSignerLevelDBFolder
	}
	db, err := leveldb.OpenFile(levelDbFolder, nil)
	if nil != err {
		return nil, errors.Wrapf(err, "fail to open level db %s", levelDbFolder)
	}
	levelDbStorage, err := blockscanner.NewLevelDBScannerStorage(db)
	if nil != err {
		return nil, errors.New("fail to create leven db")
	}
	return &StateChanBlockScannerStorage{
		LevelDBScannerStorage: levelDbStorage,
		db:                    db,
	}, nil
}

// Close underlying db
func (s *StateChanBlockScannerStorage) Close() error {
	return s.db.Close()
}
