package blockscanner

import (
	"errors"
	"fmt"
	"io"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

// ScannerStorage define the method need to be used by scanner
type ScannerStorage interface {
	GetScanPos() (int64, error)
	SetScanPos(block int64) error

	SetBlockScanStatus(block Block, status BlockScanStatus) error
	RemoveBlockStatus(block int64) error

	GetBlocksForRetry(failedOnly bool) ([]Block, error)
	io.Closer
}

// BlockScannerStorage
type BlockScannerStorage struct {
	*LevelDBScannerStorage
	db *leveldb.DB
}

func NewBlockScannerStorage(levelDbFolder string) (*BlockScannerStorage, error) {
	var err error
	var db *leveldb.DB
	if len(levelDbFolder) == 0 {
		// no directory given, use in memory store
		storage := storage.NewMemStorage()
		db, err = leveldb.Open(storage, nil)
		if err != nil {
			return nil, fmt.Errorf("fail to in memory open level db: %w", err)
		}
	} else {
		db, err = leveldb.OpenFile(levelDbFolder, nil)
		if err != nil {
			return nil, fmt.Errorf("fail to open level db %s: %w", levelDbFolder, err)
		}
	}
	levelDbStorage, err := NewLevelDBScannerStorage(db)
	if err != nil {
		return nil, errors.New("fail to create level db")
	}
	return &BlockScannerStorage{
		LevelDBScannerStorage: levelDbStorage,
		db:                    db,
	}, nil
}

func (s *BlockScannerStorage) GetInternalDb() *leveldb.DB {
	return s.db
}
