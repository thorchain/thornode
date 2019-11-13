package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

// NodeStatus Represent the Node status
type NodeStatus uint8

// As soon as user donate a certain amount of asset(defined later)
// their node adddress will be whitelisted
// once we discover their observer had send tx in to thorchain , then their status will be standby
// once we rotate them in , then they will be active
const (
	Unknown NodeStatus = iota
	WhiteListed
	Standby
	Ready
	Active
	Disabled
)

var nodeStatusStr = map[string]NodeStatus{
	"unknown":     Unknown,
	"whitelisted": WhiteListed,
	"standby":     Standby,
	"ready":       Ready,
	"active":      Active,
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
	NodeAddress         sdk.AccAddress `json:"node_address"` // Thor address which is an operator address
	Status              NodeStatus     `json:"status"`
	NodePubKey          common.PubKeys `json:"node_pub_keys"`
	ValidatorConsPubKey string         `json:"validator_cons_pub_key"`
	Bond                sdk.Uint       `json:"bond"`
	ActiveBlockHeight   int64          `json:"active_block_height"` // The block height when this node account became active status
	BondAddress         common.Address `json:"bond_address"`        // BNB Address to send bond from. It also indicates the operator address to whilelist and associate.
	SlashPoints         int64          `json:"slash_points"`        // Amount of penalty points the users has incurred.
	// start from which block height this node account is in current status
	// StatusSince field is important , it has been used to sort node account , used for validator rotation
	StatusSince      int64           `json:"status_since"`
	ObserverActive   bool            `json:"observer_active"`
	SignerActive     bool            `json:"signer_active"`
	SignerMembership []common.PubKey `json:"signer_membership"`
	Version          int64           `json:"version"`
}

// NewNodeAccount create new instance of NodeAccount
func NewNodeAccount(nodeAddress sdk.AccAddress, status NodeStatus, nodePubKeys common.PubKeys, validatorConsPubKey string, bond sdk.Uint, bondAddress common.Address, height int64) NodeAccount {
	na := NodeAccount{
		NodeAddress:         nodeAddress,
		NodePubKey:          nodePubKeys,
		ValidatorConsPubKey: validatorConsPubKey,
		Bond:                bond,
		BondAddress:         bondAddress,
	}
	na.UpdateStatus(status, height)
	return na
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
	if n.BondAddress.IsEmpty() {
		return errors.New("bond address is empty")
	}

	return nil
}

// UpdateStatus change the status of node account, in the mean time update StatusSince field
func (n *NodeAccount) UpdateStatus(status NodeStatus, height int64) {
	n.Status = status
	n.StatusSince = height
}

// Equals compare two node account, to see whether they are equal
func (n NodeAccount) Equals(n1 NodeAccount) bool {
	if n.NodeAddress.Equals(n1.NodeAddress) &&
		n.NodePubKey.Equals(n1.NodePubKey) &&
		n.ValidatorConsPubKey == n1.ValidatorConsPubKey &&
		n.BondAddress.Equals(n1.BondAddress) &&
		n.Bond.Equal(n1.Bond) &&
		n.Version == n1.Version {
		return true
	}
	return false
}

// String implement fmt.Stringer interface
func (n NodeAccount) String() string {
	sb := strings.Builder{}
	sb.WriteString("node:" + n.NodeAddress.String() + "\n")
	sb.WriteString("status:" + n.Status.String() + "\n")
	sb.WriteString("node pubkeys:" + n.NodePubKey.String() + "\n")
	sb.WriteString("validator consensus pub key:" + n.ValidatorConsPubKey + "\n")
	sb.WriteString("bond:" + n.Bond.String() + "\n")
	sb.WriteString("version:" + strconv.FormatInt(n.Version, 10) + "\n")
	sb.WriteString("bond address:" + n.BondAddress.String() + "\n")
	return sb.String()
}

// AddSignerPubKey add a key to node account
func (n *NodeAccount) TryAddSignerPubKey(key common.PubKey) {
	if key.IsEmpty() {
		return
	}
	for _, item := range n.SignerMembership {
		if item.Equals(key) {
			return
		}
	}
	n.SignerMembership = append(n.SignerMembership, key)
}

// TryRemoveSignerPubKey remove the given pubkey from
func (n *NodeAccount) TryRemoveSignerPubKey(key common.PubKey) {
	if key.IsEmpty() {
		return
	}
	idxToDelete := -1
	for idx, item := range n.SignerMembership {
		if item.Equals(key) {
			idxToDelete = idx
		}
	}
	if idxToDelete != -1 {
		n.SignerMembership = append(n.SignerMembership[:idxToDelete], n.SignerMembership[idxToDelete+1:]...)
	}
}

// NodeAccounts just a list of NodeAccount
type NodeAccounts []NodeAccount

// IsTrustAccount validate whether the given account address belongs to an currently active validator
func (nodeAccounts NodeAccounts) IsTrustAccount(addr sdk.AccAddress) bool {
	for _, na := range nodeAccounts {
		if na.Status == Active && addr.Equals(na.NodeAddress) {
			return true
		}
	}
	return false
}

// NodeAccount sort interface , it will sort by StatusSince field, and then by SignerBNBAddress
func (nodeAccounts NodeAccounts) Less(i, j int) bool {
	if nodeAccounts[i].StatusSince < nodeAccounts[j].StatusSince {
		return true
	}
	if nodeAccounts[i].StatusSince > nodeAccounts[j].StatusSince {
		return false
	}
	return nodeAccounts[i].NodeAddress.String() < nodeAccounts[j].NodeAddress.String()
}
func (nodeAccounts NodeAccounts) Len() int { return len(nodeAccounts) }
func (nodeAccounts NodeAccounts) Swap(i, j int) {
	nodeAccounts[i], nodeAccounts[j] = nodeAccounts[j], nodeAccounts[i]
}

// First return the first item in the slice
func (nodeAccounts NodeAccounts) First() NodeAccount {
	if len(nodeAccounts) > 0 {
		return nodeAccounts[0]
	}
	return NodeAccount{}
}
