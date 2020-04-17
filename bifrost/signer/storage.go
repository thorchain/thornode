package signer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
)

const (
	DefaultSignerLevelDBFolder = "signer_data"
	txOutPrefix                = "txout-v1-"
)

type TxStatus int

const (
	TxUnknown TxStatus = iota
	TxAvailable
	TxUnavailable
	TxSpent
)

type TxOutStoreItem struct {
	TxOutItem types.TxOutItem
	Status    TxStatus
	Height    int64
}

func NewTxOutStoreItem(height int64, item types.TxOutItem) TxOutStoreItem {
	return TxOutStoreItem{
		TxOutItem: item,
		Height:    height,
		Status:    TxAvailable,
	}
}

func (s *TxOutStoreItem) Key() string {
	buf, _ := json.Marshal(struct {
		TxOutItem types.TxOutItem
		Height    int64
	}{
		s.TxOutItem,
		s.Height,
	})
	sha256Bytes := sha256.Sum256(buf)
	return fmt.Sprintf("%s%s", txOutPrefix, hex.EncodeToString(sha256Bytes[:]))
}

type SignerStorage interface {
	Set(item TxOutStoreItem) error
	Batch(items []TxOutStoreItem) error
	Get(key string) (TxOutStoreItem, error)
	Has(key string) bool
	Remove(item TxOutStoreItem) error
	List() []TxOutStoreItem
	OrderedLists() map[string][]TxOutStoreItem
	Close() error
}

type SignerStore struct {
	*blockscanner.LevelDBScannerStorage
	logger     zerolog.Logger
	db         *leveldb.DB
	passphrase string
}

// NewSignerStore create a new instance of SignerStore. If no folder is given,
// an in memory implementation is used.
func NewSignerStore(levelDbFolder, passphrase string) (*SignerStore, error) {
	var db *leveldb.DB
	var err error
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
	levelDbStorage, err := blockscanner.NewLevelDBScannerStorage(db)
	if err != nil {
		return nil, errors.New("fail to create level db")
	}
	return &SignerStore{
		LevelDBScannerStorage: levelDbStorage,
		logger:                log.With().Str("module", "signer-storage").Logger(),
		db:                    db,
		passphrase:            passphrase,
	}, nil
}

func (s *SignerStore) Set(item TxOutStoreItem) error {
	key := item.Key()
	buf, err := json.Marshal(item)
	if err != nil {
		s.logger.Error().Err(err).Msg("fail to marshal to txout store item")
		return err
	}
	if len(s.passphrase) > 0 {
		buf, err = common.Encrypt(buf, s.passphrase)
		if err != nil {
			s.logger.Error().Err(err).Msg("fail to encrypt txout item")
			return err
		}
	}
	if err := s.db.Put([]byte(key), buf, nil); err != nil {
		s.logger.Error().Err(err).Msg("fail to set txout item")
		return err
	}
	return nil
}

func (s *SignerStore) Batch(items []TxOutStoreItem) error {
	batch := new(leveldb.Batch)
	for _, item := range items {
		key := item.Key()
		buf, err := json.Marshal(item)
		if err != nil {
			s.logger.Error().Err(err).Msg("fail to marshal to txout store item")
			return err
		}
		if len(s.passphrase) > 0 {
			buf, err = common.Encrypt(buf, s.passphrase)
			if err != nil {
				s.logger.Error().Err(err).Msg("fail to encrypt txout item")
				return err
			}
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
	if len(s.passphrase) > 0 {
		buf, err = common.Decrypt(buf, s.passphrase)
		if err != nil {
			s.logger.Error().Err(err).Msg("fail to decrypt txout item")
			return item, err
		}
	}
	if err := json.Unmarshal(buf, &item); err != nil {
		s.logger.Error().Err(err).Msg("fail to unmarshal to txout store item")
		return item, err
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
func (s *SignerStore) List() []TxOutStoreItem {
	iterator := s.db.NewIterator(util.BytesPrefix([]byte(txOutPrefix)), nil)
	defer iterator.Release()
	var results []TxOutStoreItem
	for iterator.Next() {
		var err error
		buf := iterator.Value()
		if len(buf) == 0 {
			continue
		}

		if len(s.passphrase) > 0 {
			buf, err = common.Decrypt(buf, s.passphrase)
			if err != nil {
				s.logger.Error().Err(err).Msg("fail to decrypt txout item")
				continue
			}
		}

		var item TxOutStoreItem
		if err := json.Unmarshal(buf, &item); err != nil {
			s.logger.Error().Err(err).Msg("fail to unmarshal to txout store item")
			continue
		}

		// ignore already spent items
		if item.Status == TxSpent {
			continue
		}

		results = append(results, item)
	}

	// Ensure that we sort our list by block height (lowest to highest), then
	// by Hash. This makes best efforts to ensure that each node is iterating
	// through their list of items as closely as possible
	sort.SliceStable(results, func(i, j int) bool { return results[i].TxOutItem.Hash() < results[j].TxOutItem.Hash() })
	sort.SliceStable(results, func(i, j int) bool { return results[i].Height < results[j].Height })
	return results
}

// OrderedLists
func (s *SignerStore) OrderedLists() map[string][]TxOutStoreItem {
	lists := make(map[string][]TxOutStoreItem, 0)
	for _, item := range s.List() {
		key := fmt.Sprintf("%s-%s", item.TxOutItem.Chain.String(), item.TxOutItem.VaultPubKey.String())
		if _, ok := lists[key]; !ok {
			lists[key] = make([]TxOutStoreItem, 0)
		}
		lists[key] = append(lists[key], item)
	}
	return lists
}

// Close underlying db
func (s *SignerStore) Close() error {
	return s.db.Close()
}
