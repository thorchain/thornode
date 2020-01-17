package types

import (
	"encoding/json"
)

/// AccountResp the response from thorchain
type AccountResp struct {
	Height string          `json:"height"`
	Result json.RawMessage `json:"result"`
}
