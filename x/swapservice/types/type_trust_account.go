package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/common"
)

// TrustAccount represent those accounts we can trust, and can be used to sign tx
type TrustAccount struct {
	SignerBNBAddress       common.Address `json:"bnb_signer_acc"`
	ObserverBEPAddress     sdk.AccAddress `json:"bep_observer_acc"`
	ValidatorBEPConsPubKey string         `json:"bepv_validator_acc"`
}

// TrustAccounts just a list of trust account
type TrustAccounts []TrustAccount

// NewTrustAccount create a new instance of trust account
func NewTrustAccount(signerBNBAddress common.Address, observerBepAddress sdk.AccAddress, validatorConsPubKey string) TrustAccount {
	return TrustAccount{
		SignerBNBAddress:       signerBNBAddress,
		ObserverBEPAddress:     observerBepAddress,
		ValidatorBEPConsPubKey: validatorConsPubKey,
	}
}

// IsValid do some basic check make sure all the field has legit value
func (ta TrustAccount) IsValid() error {
	if ta.ObserverBEPAddress.Empty() {
		return errors.New("Observer BEP address cannot be empty")
	}
	if len(ta.ValidatorBEPConsPubKey) == 0 {
		return errors.New("Validator BEP consensus public key cannot be empty")
	}

	return nil
}

// Equals is used to check whether one trust account equals to another
func (ta TrustAccount) Equals(ta1 TrustAccount) bool {
	if strings.EqualFold(ta.ValidatorBEPConsPubKey, ta1.ValidatorBEPConsPubKey) &&
		ta.SignerBNBAddress.Equals(ta1.SignerBNBAddress) &&
		ta.ObserverBEPAddress.Equals(ta1.ObserverBEPAddress) {
		return true
	}
	return false
}

// String implement fmt.Stringer interface
func (ta TrustAccount) String() string {
	sb := strings.Builder{}
	sb.WriteString("signer_bnb_address:" + ta.SignerBNBAddress.String() + "\n")
	sb.WriteString("observer_bep_address:" + ta.ObserverBEPAddress.String() + "\n")
	sb.WriteString("validator_bep_consensus_public_key:" + ta.ValidatorBEPConsPubKey + "\n")
	return sb.String()
}
