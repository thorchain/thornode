package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	common "gitlab.com/thorchain/bepswap/common"
)

type trustAccountStatus string

const (
	ActiveTrustAccount      trustAccountStatus = "active"      // is an active trust account
	InactiveTrustAccount    trustAccountStatus = "inactive"    // is not an active trust account but can become one
	UnavailableTrustAccount trustAccountStatus = "unavailable" // outside of the pool of trust accounts
	BannedTrustAccount      trustAccountStatus = "banned"      // banned from ever becoming a trust account again
)

// TrustAccount represent those accounts we can trust, and can be used to sign tx
type TrustAccount struct {
	Status     trustAccountStatus `json:"status"`
	BepAddress sdk.AccAddress     `json:"bep_address"`
	BnbAddress common.BnbAddress  `json:"bnb_address"`
}

type TrustAccounts []TrustAccount

func NewTrustAccount(address string, bnb common.BnbAddress) (TrustAccount, error) {
	addr, err := sdk.AccAddressFromBech32(address)
	return TrustAccount{
		Status:     UnavailableTrustAccount,
		BepAddress: addr,
		BnbAddress: bnb,
	}, err
}

func (ta TrustAccount) IsActive() bool {
	return ta.Status == ActiveTrustAccount
}

func (ta TrustAccount) IsInactive() bool {
	return ta.Status == InactiveTrustAccount
}

// String implement fmt.Stringer interface
func (ta TrustAccount) String() string {
	sb := strings.Builder{}
	sb.WriteString("status:" + string(ta.Status))
	sb.WriteString("address:" + ta.BepAddress.String())
	sb.WriteString("BNB Address:" + ta.BnbAddress.String())
	return sb.String()
}

func (trusts TrustAccounts) IsActiveTrustAccount(addr sdk.AccAddress) bool {
	for _, trust := range trusts {
		if trust.BepAddress.Equals(addr) && trust.IsActive() {
			return true
		}
	}
	return false
}
