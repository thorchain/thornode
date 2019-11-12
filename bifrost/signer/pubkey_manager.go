package signer

import (
	"sync"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type PubKeyManager struct {
	rwMutex *sync.RWMutex
	pks     []common.PubKey
}

// NewPubKeyManager create a new instance of PubKeyManager
func NewPubKeyManager() *PubKeyManager {
	return &PubKeyManager{
		rwMutex: &sync.RWMutex{},
		pks:     make([]common.PubKey, 0),
	}
}

func (pkm *PubKeyManager) Add(pk common.PubKey) {
	pkm.rwMutex.Lock()
	defer pkm.rwMutex.Unlock()
	for _, pubkey := range pkm.pks {
		if pk.Equals(pubkey) {
			return
		}
	}
	pkm.pks = append(pkm.pks, pk)
}

// HasKey determinate whether the given key is in the PubKeyManager
func (pkm *PubKeyManager) HasKey(pk common.PubKey) bool {
	for _, item := range pkm.pks {
		if item.Equals(pk) {
			return true
		}
	}
	return false
}

func (pkm *PubKeyManager) Remove(pk common.PubKey) {
	pkm.rwMutex.Lock()
	defer pkm.rwMutex.Unlock()
	for i, pubkey := range pkm.pks {
		if pk.Equals(pubkey) {
			pkm.pks[i] = pkm.pks[len(pkm.pks)-1]         // Copy last element to index i.
			pkm.pks[len(pkm.pks)-1] = common.EmptyPubKey // Erase last element (write zero value).
			pkm.pks = pkm.pks[:len(pkm.pks)-1]           // Truncate slice.
			break
		}
	}
}
