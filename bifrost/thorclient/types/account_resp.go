package types

import (
	"encoding/json"
)

/// AccountResp the response from thorclient
type AccountResp struct {
	Height string          `json:"height"`
	Result json.RawMessage `json:"result"`
}
