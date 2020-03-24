package binance

import (
	"sync"

	"gitlab.com/thorchain/thornode/common"
)

type BinanceMetadata struct {
	AccountNumber int64
	SeqNumber     int64
	BlockHeight   int64
}

type BinanceMetaDataStore struct {
	lock  *sync.Mutex
	accts map[common.PubKey]BinanceMetadata
}

func NewBinanceMetaDataStore() *BinanceMetaDataStore {
	return &BinanceMetaDataStore{
		lock:  &sync.Mutex{},
		accts: make(map[common.PubKey]BinanceMetadata, 0),
	}
}

func (b *BinanceMetaDataStore) Get(pk common.PubKey) BinanceMetadata {
	b.lock.Lock()
	defer b.lock.Unlock()
	if val, ok := b.accts[pk]; ok {
		return val
	}
	return BinanceMetadata{}
}

func (b *BinanceMetaDataStore) GetByAccount(acct int64) BinanceMetadata {
	b.lock.Lock()
	defer b.lock.Unlock()
	for _, meta := range b.accts {
		if meta.AccountNumber == acct {
			return meta
		}
	}
	return BinanceMetadata{}
}

func (b *BinanceMetaDataStore) Set(pk common.PubKey, meta BinanceMetadata) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.accts[pk] = meta
}

func (b *BinanceMetaDataStore) SeqInc(pk common.PubKey) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if meta, ok := b.accts[pk]; ok {
		meta.SeqNumber += 1
		b.accts[pk] = meta
	}
}
