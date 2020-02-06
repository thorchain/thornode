package signer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

const (
	DefaultSignerLevelDBFolder = "signer_data"
	txOutPrefix                = "txout-v1-"
)

type TxOutStoreItem struct {
	TxOutItem types.TxOutItem
	Height    int64
}

func NewTxOutStoreItem(height int64, item types.TxOutItem) TxOutStoreItem {
	return TxOutStoreItem{
		TxOutItem: item,
		Height:    height,
	}
}

type SignerStore struct {
	*blockscanner.LevelDBScannerStorage
	db *leveldb.DB
}

// NewSignerStore create a new instance of SignerStore
func NewSignerStore(levelDbFolder string) (*SignerStore, error) {
	if len(levelDbFolder) == 0 {
		levelDbFolder = DefaultSignerLevelDBFolder
	}
	db, err := leveldb.OpenFile(levelDbFolder, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to open level db %s", levelDbFolder)
	}
	levelDbStorage, err := blockscanner.NewLevelDBScannerStorage(db)
	if err != nil {
		return nil, errors.New("fail to create level db")
	}
	return &SignerStore{
		LevelDBScannerStorage: levelDbStorage,
		db:                    db,
	}, nil
}

func (s *TxOutStoreItem) Key() string {
	buf, _ := json.Marshal(s)
	sha256Bytes := sha256.Sum256(buf)
	return fmt.Sprintf("%s%s", txOutPrefix, hex.EncodeToString(sha256Bytes[:]))
}

func (s *SignerStore) Set(item TxOutStoreItem) error {
	key := item.Key()
	buf, err := json.Marshal(item)
	if err != nil {
		return errors.Wrap(err, "fail to marshal TxOutStoreItem to json")
	}
	if err := s.db.Put([]byte(key), buf, nil); err != nil {
		return errors.Wrap(err, "fail to set txout item")
	}
	return nil
}

func (s *SignerStore) Batch(items []TxOutStoreItem) error {
	batch := new(leveldb.Batch)
	for _, item := range items {
		key := item.Key()
		buf, err := json.Marshal(item)
		if err != nil {
			return errors.Wrap(err, "fail to marshal TxOutStoreItem to json")
		}
		batch.Put([]byte(key), buf)
	}
	return s.db.Write(batch, nil)
}

func (s *SignerStore) Get(key string) (item TxOutStoreItem, err error) {
	ok, err := s.db.Has([]byte(key), nil)
	if !ok || err != nil {
		return
	}
	buf, err := s.db.Get([]byte(key), nil)
	if err := json.Unmarshal(buf, &item); err != nil {
		return item, errors.Wrap(err, "fail to unmarshal to txout store item")
	}
	return
}

func (s *SignerStore) Has(key string) (ok bool) {
	ok, _ = s.db.Has([]byte(key), nil)
	return
}

func (s *SignerStore) Remove(item TxOutStoreItem) error {
	return s.db.Delete([]byte(item.Key()), nil)
}

// GetTxOutsForRetry send back tx out to retry depending on arg failed only
func (s *SignerStore) List() ([]TxOutStoreItem, error) {
	iterator := s.db.NewIterator(util.BytesPrefix([]byte(txOutPrefix)), nil)
	defer iterator.Release()
	var results []TxOutStoreItem
	for iterator.Next() {
		buf := iterator.Value()
		if len(buf) == 0 {
			continue
		}
		var item TxOutStoreItem
		if err := json.Unmarshal(buf, &item); err != nil {
			return nil, errors.Wrap(err, "fail to unmarshal to txout store item")
		}
		results = append(results, item)
	}

	// Ensure that we sort our list by block height (lowest to highest). This
	// makes best efforts to ensure that each node is iterating through their
	// list of items as closely as possible
	sort.SliceStable(results[:], func(i, j int) bool { return results[i].Height < results[j].Height })
	return results, nil
}

// Close underlying db
func (s *SignerStore) Close() error {
	return s.db.Close()
}
