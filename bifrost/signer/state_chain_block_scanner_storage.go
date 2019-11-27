package signer

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
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

type TxOutLocalStatus byte

const (
	Processing TxOutLocalStatus = iota
	Failed
)

// TxOutLocalItem for local storage
type TxOutLocalItem struct {
	TxOut  types.TxOut      `json:"tx_out"`
	Status TxOutLocalStatus `json:"status"`
}

func (s *StateChanBlockScannerStorage) getTxOutKey(height string) string {
	return fmt.Sprintf("txout-%s", height)
}

// SetTxOutStatus store the txout locally
func (s *StateChanBlockScannerStorage) SetTxOutStatus(txOut types.TxOut, status TxOutLocalStatus) error {
	txOutLocalItem := TxOutLocalItem{
		TxOut:  txOut,
		Status: status,
	}
	buf, err := json.Marshal(txOutLocalItem)
	if nil != err {
		return errors.Wrap(err, "fail to marshal TxOutLocalItem to json")
	}
	if err := s.db.Put([]byte(s.getTxOutKey(txOut.Height)), buf, nil); nil != err {
		return errors.Wrap(err, "fail to set txout local item status")
	}
	return nil
}

// RemoveTxOut delete the given txout from data store
func (s *StateChanBlockScannerStorage) RemoveTxOut(txOut types.TxOut) error {
	key := s.getTxOutKey(txOut.Height)
	return s.db.Delete([]byte(key), nil)
}

// GetFailedBlocksForRetry
func (ldbss *StateChanBlockScannerStorage) GetTxOutsForRetry(failedOnly bool) ([]types.TxOut, error) {
	iterator := ldbss.db.NewIterator(util.BytesPrefix([]byte("txout-")), nil)
	defer iterator.Release()
	var results []types.TxOut
	for iterator.Next() {
		buf := iterator.Value()
		if len(buf) == 0 {
			continue
		}
		var txOutLocalItem TxOutLocalItem
		if err := json.Unmarshal(buf, &txOutLocalItem); nil != err {
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

// Close underlying db
func (s *StateChanBlockScannerStorage) Close() error {
	return s.db.Close()
}
