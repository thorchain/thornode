package bitcoin

import (
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// PrefixUTXOStorage declares prefix to use in leveldb to avoid conflicts
const PrefixUTXOStorage = "utxo-"

// LevelDBUTXOAccessor struct
type LevelDBUTXOAccessor struct {
	db *leveldb.DB
}

// NewLevelDBUTXOAccessor creates a new level db utxo accessor
func NewLevelDBUTXOAccessor(db *leveldb.DB) (*LevelDBUTXOAccessor, error) {
	return &LevelDBUTXOAccessor{db: db}, nil
}

// GetUTXOs retrieves all utxo from level db storage
func (t *LevelDBUTXOAccessor) GetUTXOs() ([]UnspentTransactionOutput, error) {
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
