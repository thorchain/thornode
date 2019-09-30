package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NodeStatus Represent the Node status
type NodeStatus uint8

// As soon as user donate a certain amount of token(defined later)
// their node adddress will be whitelisted
// once we discover their observer had send tx in to statechain , then their status will be standby
// once we rotate them in , then they will be active
const (
	Unknown NodeStatus = iota
	WhiteListed
	Standby
	Nominated
	Ready
	Active
	Queued
	Disabled
)

var nodeStatusStr = map[string]NodeStatus{
	"unknown":     Unknown,
	"whitelisted": WhiteListed,
	"standby":     Standby,
	"nominated":   Nominated,
	"active":      Active,
	"queued":      Queued,
	"disabled":    Disabled,
}

// String implement stringer
func (ps NodeStatus) String() string {
	for key, item := range nodeStatusStr {
		if item == ps {
			return key
		}
	}
	return ""
}

func (ps NodeStatus) Valid() error {
	if ps.String() == "" {
		return fmt.Errorf("invalid node status")
	}
	return nil
}

// MarshalJSON marshal NodeStatus to JSON in string form
func (ps NodeStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(ps.String())
}

// UnmarshalJSON convert string form back to NodeStatus
func (ps *NodeStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); nil != err {
		return err
	}
	*ps = GetNodeStatus(s)
	return nil
}

// GetNodeStatus from string
func GetNodeStatus(ps string) NodeStatus {
	for key, item := range nodeStatusStr {
		if strings.EqualFold(key, ps) {
			return item
		}
	}

	return Unknown
}

// NodeAccount represent node
type NodeAccount struct {
	NodeAddress sdk.AccAddress `json:"node_address"`
	Status      NodeStatus     `json:"status"`
	Accounts    TrustAccount   `json:"accounts"`
	Bond        sdk.Uint       `json:"bond"`
}

// NewNodeAccount create new instance of NodeAccount
func NewNodeAccount(nodeAddress sdk.AccAddress, status NodeStatus, accounts TrustAccount) NodeAccount {
	return NodeAccount{
		NodeAddress: nodeAddress,
		Status:      status,
		Accounts:    accounts,
		Bond:        sdk.ZeroUint(),
	}
}

// IsEmpty decide whether NodeAccount is empty
func (n NodeAccount) IsEmpty() bool {
	return n.NodeAddress.Empty()
}

// IsValid check whether NodeAccount has all necessary values
func (n NodeAccount) IsValid() error {
	if n.NodeAddress.Empty() {
		return errors.New("node bep address is empty")
	}
	return n.Accounts.IsValid()
}

// String implement fmt.Stringer interface
func (n NodeAccount) String() string {
	sb := strings.Builder{}
	sb.WriteString("node:" + n.NodeAddress.String() + "\n")
	sb.WriteString("status:" + n.Status.String() + "\n")
	sb.WriteString("account:" + n.Accounts.String() + "\n")
	sb.WriteString("bond:" + n.Bond.String() + "\n")
	return sb.String()
}

// NodeAccounts just a list of NodeAccount
type NodeAccounts []NodeAccount

// IsTrustAccount validate whether the given account address is an observer address
func (nodeAccounts NodeAccounts) IsTrustAccount(addr sdk.AccAddress) bool {
	for _, na := range nodeAccounts {
		if na.Status == Active && na.Accounts.ObserverBEPAddress.Equals(addr) {
			return true
		}
	}
	return false
}
