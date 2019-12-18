package types

import "gitlab.com/thorchain/thornode/common"

type Keygen []common.PubKey

type Keygens struct {
	Height  uint64   `json:"height"`
	Keygens []Keygen `json:"keygens"`
}
