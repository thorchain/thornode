package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/binance-chain/go-sdk/common/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/syndtr/goleveldb/leveldb"
)

type DataStore struct {
	folder       string
	filePathName string
	nodeName     string
	db           *leveldb.DB
	logger       zerolog.Logger
}

const fileName = `.bebswap.leveldb`
const nodeNameKey = `node-name`

// NewDataStore create a new instance of DataStore
// all level db related logic here
func NewDataStore(folder string, logger zerolog.Logger) (*DataStore, error) {
	if len(folder) == 0 {
		f, err := os.Getwd()
		if nil != err {
			return nil, errors.Wrap(err, "fail to get current working folder")
		}
		folder = f
	}
	filePathName := filepath.Join(folder, fileName)
	db, err := leveldb.OpenFile(filePathName, nil)
	if nil != err {
		return nil, errors.Wrap(err, "fail to open db")
	}
	nodeName, err := getOrCreateNodeName(db)
	if nil != err {
		return nil, errors.Wrap(err, "fail to get node name , we can't proceed.")
	}
	return &DataStore{
		folder:       folder,
		filePathName: filePathName,
		nodeName:     nodeName,
		db:           db,
		logger:       logger.With().Str("module", "datastore").Logger(),
	}, nil
}

func getOrCreateNodeName(db *leveldb.DB) (string, error) {
	// Load the node's name, or generate a new one
	n, err := db.Get([]byte(nodeNameKey), nil)
	if nil != err {
		if err != leveldb.ErrNotFound {
			return "", errors.Wrap(err, "fail to get node name from db")
		}
		id, err := uuid.NewV4()
		if nil != err {
			return "", errors.Wrap(err, "fail to generate v4 uuid")
		}
		nodeName := fmt.Sprintf("bep2swap-%s", id)
		if err := db.Put([]byte(nodeNameKey), []byte(nodeName), nil); nil != err {
			return "", errors.Wrap(err, "fail to write node name to db")
		}
		return nodeName, nil
	}
	return string(n), nil
}

// Get from the data store
func (ds *DataStore) Get(key []byte) (value []byte, err error) {
	return ds.db.Get(key, nil)
}

// Put save something to the datastore
func (ds *DataStore) Put(key, value []byte) error {
	return ds.db.Put(key, value, nil)
}

// Close DataStore instance
func (ds *DataStore) Close() error {
	if err := ds.db.Close(); nil != err {
		return errors.Wrap(err, "fail to close db")
	}
	return nil
}
