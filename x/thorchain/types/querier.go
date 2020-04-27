package types

import (
	"fmt"
	"strings"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

// Query Result Payload for a pools query
type QueryResPools []Pool

// implement fmt.Stringer
func (n QueryResPools) String() string {
	var assets []string
	for _, record := range n {
		assets = append(assets, record.Asset.String())
	}
	return strings.Join(assets, "\n")
}

type QueryResHeights struct {
	Chain            common.Chain `json:"chain"`
	LastChainHeight  int64        `json:"lastobservedin"`
	LastSignedHeight int64        `json:"lastsignedout"`
	Statechain       int64        `json:"statechain"`
}

func (h QueryResHeights) String() string {
	return fmt.Sprintf("Chain: %d, Signed: %d, Statechain: %d", h.LastChainHeight, h.LastSignedHeight, h.Statechain)
}

type ResTxOut struct {
	Height  int64        `json:"height"`
	Hash    common.TxID  `json:"hash"`
	Chain   common.Chain `json:"chain"`
	TxArray []TxOutItem  `json:"tx_array"`
}

type QueryResTxOut struct {
	Chains map[common.Chain]ResTxOut `json:"chains"`
}

type QueryYggdrasilVaults struct {
	Vault      Vault      `json:"vault"`
	Status     NodeStatus `json:"status"`
	Bond       sdk.Uint   `json:"bond"`
	TotalValue sdk.Uint   `json:"total_value"`
}

type QueryNodeAccount struct {
	NodeAddress         sdk.AccAddress   `json:"node_address"`
	Status              NodeStatus       `json:"status"`
	PubKeySet           common.PubKeySet `json:"pub_key_set"`
	ValidatorConsPubKey string           `json:"validator_cons_pub_key"`
	Bond                sdk.Uint         `json:"bond"`
	ActiveBlockHeight   int64            `json:"active_block_height"`
	BondAddress         common.Address   `json:"bond_address"`
	StatusSince         int64            `json:"status_since"`
	SignerMembership    common.PubKeys   `json:"signer_membership"`
	RequestedToLeave    bool             `json:"requested_to_leave"`
	ForcedToLeave       bool             `json:"forced_to_leave"`
	LeaveHeight         int64            `json:"leave_height"`
	IPAddress           string           `json:"ip_address"`
	Version             semver.Version   `json:"version"`
	SlashPoints         int64            `json:"slash_points"`
}

func NewQueryNodeAccount(na NodeAccount) QueryNodeAccount {
	return QueryNodeAccount{
		NodeAddress:         na.NodeAddress,
		Status:              na.Status,
		PubKeySet:           na.PubKeySet,
		ValidatorConsPubKey: na.ValidatorConsPubKey,
		Bond:                na.Bond,
		ActiveBlockHeight:   na.ActiveBlockHeight,
		BondAddress:         na.BondAddress,
		StatusSince:         na.StatusSince,
		SignerMembership:    na.SignerMembership,
		RequestedToLeave:    na.RequestedToLeave,
		ForcedToLeave:       na.ForcedToLeave,
		LeaveHeight:         na.LeaveHeight,
		IPAddress:           na.IPAddress,
		Version:             na.Version,
	}
}
