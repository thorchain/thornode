package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TrustAccountPrefix all trust account will have this prefix
const TrustAccountPrefix = `trustaccount-`

// TrustAccount represent those accounts we can trust, and can be used to sign tx
type TrustAccount struct {
	Name    string         `json:"name"`
	Address sdk.AccAddress `json:"address"`
}

// String implement fmt.Stringer interface
func (ta TrustAccount) String() string {
	sb := strings.Builder{}
	sb.WriteString("name:" + ta.Name)
	sb.WriteString("address:" + ta.Address.String())
	return sb.String()
}
