package types

// ValidatorsResp response data we send back to client when they ask for validator information
type ValidatorsResp struct {
	ActiveNodes        NodeAccounts `json:"active_nodes"`
	Nominated          *NodeAccount `json:"nominated"`
	Queued             *NodeAccount `json:"queued"`
	RotateAt           uint64       `json:"rotate_at"`
	RotateWindowOpenAt uint64       `json:"rotate_window_open_at"`
}
