package binance

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
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
	if nil != err {
		return nil, errors.Wrapf(err, "fail to open level db %s", levelDbFolder)
	}
	levelDbStorage, err := blockscanner.NewLevelDBScannerStorage(db)
	if nil != err {
		return nil, errors.New("fail to create leven db")
	}
	return &BinanceBlockScannerStorage{
		LevelDBScannerStorage: levelDbStorage,
		db:                    db,
	}, nil
}

func getTxInStatusKey(blockHeight string) string {
	return fmt.Sprintf("txin-process-status-%s", blockHeight)
}

// SetTxInStatus set the given txin to a status , in the data store
func (ldbss *BinanceBlockScannerStorage) SetTxInStatus(txIn types.TxIn, status types.TxInStatus) error {
	txStatusItem := types.TxInStatusItem{
		TxIn:   txIn,
		Status: status,
	}
	buf, err := json.Marshal(txStatusItem)
	if nil != err {
		return errors.Wrap(err, "fail to marshal TxInStatusItem to json")
	}
	if err := ldbss.db.Put([]byte(getTxInStatusKey(txIn.BlockHeight)), buf, nil); nil != err {
		return errors.Wrap(err, "fail to set tx in status")
	}
	return nil
}

// RemoveTxIn remove the given txin from the store
func (ldbss *BinanceBlockScannerStorage) RemoveTxIn(txin types.TxIn) error {
	return ldbss.db.Delete([]byte(getTxInStatusKey(txin.BlockHeight)), nil)

}

// GetTxInForRetry retrieve all txin that had been failed before to retry
func (ldbss *BinanceBlockScannerStorage) GetTxInForRetry(failedOnly bool) ([]types.TxIn, error) {
	iterator := ldbss.db.NewIterator(util.BytesPrefix([]byte("txin-process-status-")), nil)
	defer iterator.Release()
	var results []types.TxIn
	for iterator.Next() {
		buf := iterator.Value()
		if len(buf) == 0 {
			continue
		}
		var txInStatusItem types.TxInStatusItem
		if err := json.Unmarshal(buf, &txInStatusItem); nil != err {
			return nil, errors.Wrap(err, "fail to unmarshal to txin status item")
		}
		if !failedOnly {
			results = append(results, txInStatusItem.TxIn)
			continue
		}
		if txInStatusItem.Status == types.Failed {
			results = append(results, txInStatusItem.TxIn)
		}
	}
	return results, nil
}

// Close underlying db
func (ldbss *BinanceBlockScannerStorage) Close() error {
	return ldbss.db.Close()
}
