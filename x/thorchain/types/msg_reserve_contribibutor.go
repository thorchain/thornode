package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// MsgReserveContributor defines a MsgReserveContributor message
type MsgReserveContributor struct {
	Tx          common.Tx          `json:"tx"`
	Contributor ReserveContributor `json:"contributor"`
	Signer      sdk.AccAddress     `json:"signer"`
}

// NewMsgReserveContributor is a constructor function for MsgReserveContributor
func NewMsgReserveContributor(tx common.Tx, contrib ReserveContributor, signer sdk.AccAddress) MsgReserveContributor {
	return MsgReserveContributor{
		Tx:          tx,
		Contributor: contrib,
		Signer:      signer,
	}
}

func (msg MsgReserveContributor) Route() string { return RouterKey }

func (msg MsgReserveContributor) Type() string { return "set_reserve_contributor" }

// ValidateBasic runs stateless checks on the message
func (msg MsgReserveContributor) ValidateBasic() sdk.Error {
	if err := msg.Tx.IsValid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if err := msg.Contributor.IsValid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgReserveContributor) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgReserveContributor) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
