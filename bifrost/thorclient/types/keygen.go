package types

import "gitlab.com/thorchain/thornode/common"

type Keygens struct {
	Height  string           `json:"height"`
	Keygens []common.PubKeys `json:"keygens"`
}
