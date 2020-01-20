package thorchain

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"gitlab.com/thorchain/thornode/bifrostv2/blockscanner"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

const DefaultSignerLevelDBFolder = `signer_data`

type BlockScannerStorage struct {
	*blockscanner.LevelDBScannerStorage
	db *leveldb.DB
}

// NewBlockScannerStorage create a new instance of BlockScannerStorage
func NewBlockScannerStorage(levelDbFolder string) (*BlockScannerStorage, error) {
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
	return &BlockScannerStorage{
		LevelDBScannerStorage: levelDbStorage,
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

func (s *BlockScannerStorage) getTxOutKey(height int64) string {
	return fmt.Sprintf("txout-%d", height)
}

// SetTxOutStatus store the txout locally
func (s *BlockScannerStorage) SetTxOutStatus(txOut types.TxOut, status LocalStatus) error {
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
func (s *BlockScannerStorage) RemoveTxOut(txOut types.TxOut) error {
	key := s.getTxOutKey(txOut.Height)
	return s.db.Delete([]byte(key), nil)
}

// GetTxOutsForRetry
func (s *BlockScannerStorage) GetTxOutsForRetry(failedOnly bool) ([]types.TxOut, error) {
	iterator := s.db.NewIterator(util.BytesPrefix([]byte("txout-")), nil)
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

func (s *BlockScannerStorage) SetTxOutItem(toi *types.TxOutItem, height int64) error {
	return s.db.Put([]byte(toi.GetKey(height)), []byte{1}, nil)
}
func (s *BlockScannerStorage) HasTxOutItem(toi *types.TxOutItem, height int64) (bool, error) {
	return s.db.Has([]byte(toi.GetKey(height)), nil)
}

// Close underlying db
func (s *BlockScannerStorage) Close() error {
	return s.db.Close()
}
