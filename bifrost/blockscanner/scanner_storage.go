package blockscanner

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
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

type BlockScannerStorage struct {
	*LevelDBScannerStorage
	db *leveldb.DB
}

func NewBlockScannerStorage(levelDbFolder string) (*BlockScannerStorage, error) {
	if len(levelDbFolder) == 0 {
		return nil, fmt.Errorf("must pass valid directory path to scanner storage")
	}
	db, err := leveldb.OpenFile(levelDbFolder, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to open level db %s", levelDbFolder)
	}
	levelDbStorage, err := NewLevelDBScannerStorage(db)
	if err != nil {
		return nil, errors.New("fail to create leven db")
	}
	return &BlockScannerStorage{
		LevelDBScannerStorage: levelDbStorage,
		db:                    db,
	}, nil
}
