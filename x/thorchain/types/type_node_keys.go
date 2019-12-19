package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/common"
)

// NodeKeys represent those accounts THORNode can trust, and can be used to sign tx
type NodeKeys struct {
	SignerBNBAddress       common.Address `json:"bnb_signer_acc"`     // Is used by the signer to sign outbound tx.
	ObserverBEPAddress     sdk.AccAddress `json:"bep_observer_acc"`   // Thor address used to relay chain observations
	ValidatorBEPConsPubKey string         `json:"bepv_validator_acc"` // Used to sign tendermint blocks. Can be fetched via `thord tendermint show-validator`
}

// NodeKeyss just a list of node keys
type NodeKeyss []NodeKeys

// NewNodeKeys create a new instance of node keys
func NewNodeKeys(signerBNBAddress common.Address, observerBepAddress sdk.AccAddress, validatorConsPubKey string) NodeKeys {
	return NodeKeys{
		SignerBNBAddress:       signerBNBAddress,
		ObserverBEPAddress:     observerBepAddress,
		ValidatorBEPConsPubKey: validatorConsPubKey,
	}
}

// IsValid do some basic check make sure all the field has legit value
func (nk NodeKeys) IsValid() error {
	if nk.ObserverBEPAddress.Empty() {
		return errors.New("Observer BEP address cannot be empty")
	}
	if len(nk.ValidatorBEPConsPubKey) == 0 {
		return errors.New("Validator BEP consensus public key cannot be empty")
	}

	return nil
}

// Equals is used to check whether one node keys equals to another
func (nk NodeKeys) Equals(nk1 NodeKeys) bool {
	if strings.EqualFold(nk.ValidatorBEPConsPubKey, nk1.ValidatorBEPConsPubKey) &&
		nk.SignerBNBAddress.Equals(nk1.SignerBNBAddress) &&
		nk.ObserverBEPAddress.Equals(nk1.ObserverBEPAddress) {
		return true
	}
	return false
}

// String implement fmt.Stringer interface
func (nk NodeKeys) String() string {
	sb := strings.Builder{}
	sb.WriteString("signer_bnb_address:" + nk.SignerBNBAddress.String() + "\n")
	sb.WriteString("observer_bep_address:" + nk.ObserverBEPAddress.String() + "\n")
	sb.WriteString("validator_bep_consensus_public_key:" + nk.ValidatorBEPConsPubKey + "\n")
	return sb.String()
}
