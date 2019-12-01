package types

import "gitlab.com/thorchain/thornode/common"

// PoolAddresses is a struct to represent the address of pool also facilitate pool rotation
type PoolAddresses struct {
	Previous           common.PoolPubKeys `json:"previous"`
	Current            common.PoolPubKeys `json:"current"`
	Next               common.PoolPubKeys `json:"next"`
	RotateAt           int64              `json:"rotate_at"`
	RotateWindowOpenAt int64              `json:"rotate_window_open_at"`
}

// NewPoolAddresses create a new instance of PoolAddress
func NewPoolAddresses(previous, current, next common.PoolPubKeys, rotateAt, rotateWindowOpenAt int64) *PoolAddresses {
	return &PoolAddresses{
		Previous:           previous,
		Current:            current,
		Next:               next,
		RotateAt:           rotateAt,
		RotateWindowOpenAt: rotateWindowOpenAt,
	}
}

// IsEmpty check whether PoolAddress is empty
func (pa PoolAddresses) IsEmpty() bool {
	// when current pool address is empty then THORNode think it is empty , even the others are not, that will not matter
	return len(pa.Current) == 0
}

var EmptyPoolAddresses PoolAddresses
