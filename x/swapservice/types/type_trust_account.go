package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	common "gitlab.com/thorchain/bepswap/common"
)

// TrustAccount represent those accounts we can trust, and can be used to sign tx
type TrustAccount struct {
	AdminAddress    common.BnbAddress `json:"admin_address"`
	ObserverAddress sdk.AccAddress    `json:"observer_address"`
	SignerAddress   common.BnbAddress `json:"signer_address"`
}

type TrustAccounts []TrustAccount

func NewTrustAccount(admin, signer common.BnbAddress, ob sdk.AccAddress) TrustAccount {
	return TrustAccount{
		AdminAddress:    admin,
		SignerAddress:   signer,
		ObserverAddress: ob,
	}
}

// String implement fmt.Stringer interface
func (ta TrustAccount) String() string {
	sb := strings.Builder{}
	sb.WriteString("admin:" + ta.AdminAddress.String())
	sb.WriteString("signer:" + ta.SignerAddress.String())
	sb.WriteString("observer:" + ta.ObserverAddress.String())
	return sb.String()
}

func (trusts TrustAccounts) IsTrustAccount(addr sdk.AccAddress) bool {
	for _, trust := range trusts {
		if trust.ObserverAddress.Equals(addr) {
			return true
		}
	}
	return false
}
