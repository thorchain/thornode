package types

import (
	stypes "gitlab.com/thorchain/bepswap/thornode/x/swapservice/types"
)

type Msg struct {
	Type  string            `json:"type"`
	Value stypes.MsgSetTxIn `json:"value"`
}
