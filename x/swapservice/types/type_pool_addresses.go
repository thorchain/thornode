package types

import "gitlab.com/thorchain/bepswap/thornode/common"

type PoolAddresses struct {
	Previous common.PubKey `json:"previous"`
	Current  common.PubKey `json:"current"`
	Next     common.PubKey `json:"next"`
	RotateAt int64         `json:"rotate_at"`
}

// NewPoolAddresses create a new instance of PoolAddress
func NewPoolAddresses(previous, current, next common.PubKey, rotateAt int64) PoolAddresses {
	return PoolAddresses{
		Previous: previous,
		Current:  current,
		Next:     next,
		RotateAt: rotateAt,
	}
}

// IsEmpty check whether PoolAddress is empty
func (pa PoolAddresses) IsEmpty() bool {
	// when current pool address is empty then we think it is empty , even the others are not, that will not matter
	return len(pa.Current) == 0
}
