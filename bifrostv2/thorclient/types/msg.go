package types

import (
	stypes "gitlab.com/thorchain/thornode/x/thorchain/types"
)

type Msg struct {
	Type  string            `json:"type"`
	Value stypes.MsgSetTxIn `json:"value"`
}
