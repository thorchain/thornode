package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TrustAccount represent those accounts we can trust, and can be used to sign tx
type TrustAccount struct {
	Name        string         `json:"name"`
	RuneAddress sdk.AccAddress `json:"rune_address"`
	BnbAddress  BnbAddress     `json:"bnb_address"`
}

func NewTrustAccount(name string, address string, bnb BnbAddress) (TrustAccount, error) {
	addr, err := sdk.AccAddressFromHex(address)
	return TrustAccount{
		Name:        name,
		RuneAddress: addr,
		BnbAddress:  bnb,
	}, err
}

// String implement fmt.Stringer interface
func (ta TrustAccount) String() string {
	sb := strings.Builder{}
	sb.WriteString("name:" + ta.Name)
	sb.WriteString("address:" + ta.RuneAddress.String())
	sb.WriteString("BNB Address:" + ta.BnbAddress.String())
	return sb.String()
}
