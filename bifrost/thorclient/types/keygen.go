package types

import "gitlab.com/thorchain/thornode/common"

// Keygen is a slice of pub keys
type Keygen []common.PubKey

// Keygens is a struct of key gen information
type Keygens struct {
	Height  uint64   `json:"height"`
	Keygens []Keygen `json:"keygens"`
}
