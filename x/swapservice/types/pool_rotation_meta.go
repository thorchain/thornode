package types

import "gitlab.com/thorchain/bepswap/thornode/common"

// PoolRotationMetaData is a structure used to help with pool rotation
type PoolRotationMetaData struct {
	ObservedNextPoolPubKey common.PubKey         `json:"observed_next_pool_pub_key"`
	ConfirmedChain         map[common.Chain]bool `json:"confirmed_chain"`
}
