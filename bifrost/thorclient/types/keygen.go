package types

import "gitlab.com/thorchain/thornode/common"

type Keygens struct {
	Height  int64            `json:"height,string"`
	Keygens []common.PubKeys `json:"keygens"`
}
