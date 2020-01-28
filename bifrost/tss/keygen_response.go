package tss

import "gitlab.com/thorchain/thornode/common"

// KeyGenResponse
type KeyGenResp struct {
	PubKey     string       `json:"pub_key"`
	BNBAddress string       `json:"pool_address"`
	Status     int          `json:"status"`
	Blame      common.Blame `json:"blame"`
}
