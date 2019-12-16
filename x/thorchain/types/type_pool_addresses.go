package types

import "gitlab.com/thorchain/thornode/common"

// PoolAddresses is a struct to represent the address of pool also facilitate pool rotation
type PoolAddresses struct {
	Previous common.PoolPubKeys `json:"previous"`
	Current  common.PoolPubKeys `json:"current"`
	Next     common.PoolPubKeys `json:"next"`
}

// NewPoolAddresses create a new instance of PoolAddress
func NewPoolAddresses(previous, current, next common.PoolPubKeys) *PoolAddresses {
	return &PoolAddresses{
		Previous: previous,
		Current:  current,
		Next:     next,
	}
}

// IsEmpty check whether PoolAddress is empty
func (pa PoolAddresses) IsEmpty() bool {
	// when current pool address is empty then THORNode think it is empty , even the others are not, that will not matter
	return len(pa.Current) == 0
}

var EmptyPoolAddresses PoolAddresses
