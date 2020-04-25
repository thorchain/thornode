package bitcoin

import (
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"gitlab.com/thorchain/thornode/common"
)

// PrefixUTXOStorage declares prefix to use in leveldb to avoid conflicts
const (
	PrefixUTXOStorage = "utxo-"
	TransactionFeeKey = "transactionfee"
)

// LevelDBUTXOAccessor struct
type LevelDBUTXOAccessor struct {
	db *leveldb.DB
}

// NewLevelDBUTXOAccessor creates a new level db utxo accessor
func NewLevelDBUTXOAccessor(db *leveldb.DB) (*LevelDBUTXOAccessor, error) {
	return &LevelDBUTXOAccessor{db: db}, nil
}

// GetUTXOs retrieves all utxo from level db storage
func (t *LevelDBUTXOAccessor) GetUTXOs(pubKey common.PubKey) ([]UnspentTransactionOutput, error) {
	iterator := t.db.NewIterator(util.BytesPrefix([]byte(PrefixUTXOStorage)), nil)
	defer iterator.Release()
	var results []UnspentTransactionOutput
	for iterator.Next() {
		buf := iterator.Value()
		if len(buf) == 0 {
			continue
		}
		var utxo UnspentTransactionOutput
		if err := json.Unmarshal(buf, &utxo); err != nil {
			return nil, fmt.Errorf("fail to unmarshal to utxo: %w", err)
		}
		if !pubKey.Equals(utxo.VaultPubKey) {
			continue
		}
		results = append(results, utxo)
	}
	return results, nil
}

// AddUTXO adds a utxo to level db storage
func (t *LevelDBUTXOAccessor) AddUTXO(u UnspentTransactionOutput) error {
	buf, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("fail to marshal utxo to json: %w", err)
	}
	key := fmt.Sprintf("%s%s", PrefixUTXOStorage, u.GetKey())
	if err := t.db.Put([]byte(key), buf, nil); err != nil {
		return fmt.Errorf("fail to add utxo to level db storage: %w", err)
	}
	return nil
}

// RemoveUTXO removes utxo from level db storage by key
func (t *LevelDBUTXOAccessor) RemoveUTXO(key string) error {
	key = fmt.Sprintf("%s%s", PrefixUTXOStorage, key)
	return t.db.Delete([]byte(key), nil)
}

// UpsertTransactionFee update the transaction fee in storage
func (t *LevelDBUTXOAccessor) UpsertTransactionFee(fee float64, vSize int32) error {
	transactionFee := TransactionFee{
		Fee:   fee,
		VSize: vSize,
	}
	buf, err := json.Marshal(transactionFee)
	if err != nil {
		return fmt.Errorf("fail to marshal transaction fee struct to json: %w", err)
	}
	return t.db.Put([]byte(TransactionFeeKey), buf, nil)
}

// GetTransactionFee from db
func (t *LevelDBUTXOAccessor) GetTransactionFee() (float64, int32, error) {
	buf, err := t.db.Get([]byte(TransactionFeeKey), nil)
	if err != nil {
		return 0.0, 0, fmt.Errorf("fail to get transaction fee from storage: %w", err)
	}
	var transactionFee TransactionFee
	if err := json.Unmarshal(buf, &transactionFee); err != nil {
		return 0.0, 0, fmt.Errorf("fail to unmarshal transaction fee: %w", err)
	}
	return transactionFee.Fee, transactionFee.VSize, nil
}
