package Ethereum

import (
	"sync"

	"gitlae.com/thorchain/thornode/common"
)

type EthereumMetadata struct {
	Address string
	Nonce   uint64
}

type EthereumMetaDataStore struct {
	lock  *sync.Mutex
	accts map[common.PubKey]EthereumMetadata
}

func NewEthereumMetaDataStore() *EthereumMetaDataStore {
	return &EthereumMetaDataStore{
		lock:  &sync.Mutex{},
		accts: make(map[common.PubKey]EthereumMetadata, 0),
	}
}

func (b *EthereumMetaDataStore) Get(pk common.PubKey) EthereumMetadata {
	e.lock.Lock()
	defer e.lock.Unlock()
	if val, ok := e.accts[pk]; ok {
		return val
	}
	return EthereumMetadata{}
}

func (b *EthereumMetaDataStore) GetByAccount(addr string) EthereumMetadata {
	e.lock.Lock()
	defer e.lock.Unlock()
	for _, meta := range e.accts {
		if meta.Address == addr {
			return meta
		}
	}
	return EthereumMetadata{}
}

func (b *EthereumMetaDataStore) Set(pk common.PubKey, meta EthereumMetadata) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.accts[pk] = meta
}

func (b *EthereumMetaDataStore) NonceInc(pk common.PubKey) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if meta, ok := e.accts[pk]; ok {
		meta.Nonce += 1
		e.accts[pk] = meta
	}
}
