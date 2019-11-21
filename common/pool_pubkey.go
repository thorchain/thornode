package common

import (
	"strconv"
	"strings"
)

// PoolPubKey is the pub key only related to a specific chain
type PoolPubKey struct {
	Chain   Chain   `json:"chain"`
	SeqNo   uint64  `json:"seq_no"`
	PubKey  PubKey  `json:"pub_key"`
	Address Address `json:"address"`
}

// PoolPubKeys
type PoolPubKeys []*PoolPubKey

// EmptyPoolPubKeys
var EmptyPoolPubKeys PoolPubKeys

// NewPoolPubKey create a new instance of PoolPubKey
func NewPoolPubKey(chain Chain, seqNo uint64, pubkey PubKey) (*PoolPubKey, error) {
	var err error
	pk := &PoolPubKey{
		Chain:  chain,
		SeqNo:  seqNo,
		PubKey: pubkey,
	}
	pk.Address, err = pk.GetAddress()
	return pk, err
}

// Equals compare two PoolPubKey to determinate whether they are representing the same address
func (ppk *PoolPubKey) Equals(ppk1 *PoolPubKey) bool {
	return ppk.Chain.Equals(ppk1.Chain) && ppk.PubKey.Equals(ppk1.PubKey)
}

// IsEmpty check whether the given pool pub key is empty
func (ppk PoolPubKey) IsEmpty() bool {
	return ppk.Chain.IsEmpty() && ppk.SeqNo == 0 && ppk.PubKey.IsEmpty()
}

// Stringer implementation
func (ppk PoolPubKey) String() string {
	sb := strings.Builder{}
	sb.WriteString("chain:" + ppk.Chain.String() + "\n")
	sb.WriteString("seqNo:" + strconv.FormatUint(ppk.SeqNo, 10) + "\n")
	sb.WriteString("pubkey:" + ppk.PubKey.String() + "\n")
	return sb.String()
}
func (ppk PoolPubKey) GetAddress() (Address, error) {
	return ppk.PubKey.GetAddress(ppk.Chain)
}

// GetSeqNo
func (ppk *PoolPubKey) GetSeqNo() uint64 {
	current := ppk.SeqNo
	ppk.SeqNo++
	return current
}

// Stringer implementation
func (poolPubKeys PoolPubKeys) String() string {
	sb := strings.Builder{}
	for _, ppk := range poolPubKeys {
		sb.WriteString(ppk.String())
	}
	return sb.String()
}

// GetByChain get PoolPubKey by chain,
func (poolPubKeys PoolPubKeys) GetByChain(c Chain) *PoolPubKey {
	for _, item := range poolPubKeys {
		if item.Chain.Equals(c) {
			return item
		}
	}
	return nil
}

// IsEmpty when there is no item in the list
func (poolPubKeys PoolPubKeys) IsEmpty() bool {
	return len(poolPubKeys) == 0
}

// TryAddKey trying to add the given pool pubkey into the list,if it already exist ,then we just return the list
func (poolPubKeys PoolPubKeys) TryAddKey(k *PoolPubKey) PoolPubKeys {
	if nil == k {
		return poolPubKeys
	}
	for _, item := range poolPubKeys {
		// we should only have one pub address per chain
		if item.Equals(k) || item.Chain.Equals(k.Chain) {
			// already exist
			return poolPubKeys
		}
	}
	return append(poolPubKeys, k)
}

// TryRemoveKey trying to remove the given key from the list
func (poolPubKeys PoolPubKeys) TryRemoveKey(k *PoolPubKey) PoolPubKeys {
	if nil == k {
		return poolPubKeys
	}
	idxToDelete := -1
	for idx, item := range poolPubKeys {
		if item.Equals(k) {
			idxToDelete = idx
			break
		}
	}
	// not found
	if idxToDelete == -1 {
		return poolPubKeys
	}
	return append(poolPubKeys[:idxToDelete], poolPubKeys[idxToDelete+1:]...)

}
