package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	common "gitlab.com/thorchain/bepswap/common"
)

// TrustAccount represent those accounts we can trust, and can be used to sign tx
type TrustAccount struct {
	BepAddress sdk.AccAddress    `json:"bep_address"`
	BnbAddress common.BnbAddress `json:"bnb_address"`
}

type TrustAccounts []TrustAccount

func NewTrustAccount(name string, address string, bnb common.BnbAddress) (TrustAccount, error) {
	addr, err := sdk.AccAddressFromBech32(address)
	return TrustAccount{
		BepAddress: addr,
		BnbAddress: bnb,
	}, err
}

// String implement fmt.Stringer interface
func (ta TrustAccount) String() string {
	sb := strings.Builder{}
	sb.WriteString("address:" + ta.BepAddress.String())
	sb.WriteString("BNB Address:" + ta.BnbAddress.String())
	return sb.String()
}

func (trusts TrustAccounts) IsTrustAccount(addr sdk.AccAddress) bool {
	for _, trust := range trusts {
		if trust.BepAddress.Equals(addr) {
			return true
		}
	}
	return false
}
