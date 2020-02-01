package signer

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

const DefaultSignerLevelDBFolder = `signer_data`

type ThorchainBlockScannerStorage struct {
	*blockscanner.LevelDBScannerStorage
	mutex *sync.RWMutex
	db    *leveldb.DB
}

// NewThorchainBlockScannerStorage create a new instance of ThorchainBlockScannerStorage
func NewThorchainBlockScannerStorage(levelDbFolder string) (*ThorchainBlockScannerStorage, error) {
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
	return &ThorchainBlockScannerStorage{
		LevelDBScannerStorage: levelDbStorage,
		mutex:                 &sync.RWMutex{},
		db:                    db,
	}, nil
}

type LocalStatus byte

const (
	Processing LocalStatus = iota
	Failed
)

// TxOutLocalItem for local storage
type TxOutLocalItem struct {
	TxOut  types.TxOut `json:"tx_out"`
	Status LocalStatus `json:"status"`
}

func (s *ThorchainBlockScannerStorage) getTxOutKey(height int64) string {
	return fmt.Sprintf("txout-%d", height)
}

// SetTxOutStatus store the txout locally
func (s *ThorchainBlockScannerStorage) SetTxOutStatus(txOut types.TxOut, status LocalStatus) error {
	txOutLocalItem := TxOutLocalItem{
		TxOut:  txOut,
		Status: status,
	}
	buf, err := json.Marshal(txOutLocalItem)
	if err != nil {
		return errors.Wrap(err, "fail to marshal TxOutLocalItem to json")
	}
	if err := s.db.Put([]byte(s.getTxOutKey(txOut.Height)), buf, nil); err != nil {
		return errors.Wrap(err, "fail to set txout local item status")
	}
	return nil
}

// RemoveTxOut delete the given txout from data store
func (s *ThorchainBlockScannerStorage) RemoveTxOut(txOut types.TxOut) error {
	key := s.getTxOutKey(txOut.Height)
	return s.db.Delete([]byte(key), nil)
}

// GetTxOutsForRetry send back tx out to retry depending on arg failed only
func (s *ThorchainBlockScannerStorage) GetTxOutsForRetry(failedOnly bool) ([]types.TxOut, error) {
	iterator := s.db.NewIterator(util.BytesPrefix([]byte("txout-")), nil)
	defer iterator.Release()
	var results []types.TxOut
	for iterator.Next() {
		buf := iterator.Value()
		if len(buf) == 0 {
			continue
		}
		var txOutLocalItem TxOutLocalItem
		if err := json.Unmarshal(buf, &txOutLocalItem); err != nil {
			return nil, errors.Wrap(err, "fail to unmarshal to txout item")
		}
		if !failedOnly {
			results = append(results, txOutLocalItem.TxOut)
			continue
		}
		if txOutLocalItem.Status == Failed {
			results = append(results, txOutLocalItem.TxOut)
		}
	}
	return results, nil
}

func (s *ThorchainBlockScannerStorage) SuccessTxOutItem(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.db.Put([]byte(key), []byte{0}, nil)
}

func (s *ThorchainBlockScannerStorage) ClearTxOutItem(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.db.Delete([]byte(key), nil)
}

func (s *ThorchainBlockScannerStorage) HasTxOutItem(key string) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ok, err := s.db.Has([]byte(key), nil)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	// mark as pending (2)
	return false, s.db.Put([]byte(key), []byte{2}, nil)
}

// Close underlying db
func (s *ThorchainBlockScannerStorage) Close() error {
	return s.db.Close()
}
