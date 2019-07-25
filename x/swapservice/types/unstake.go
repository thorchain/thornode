package types

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgSetUnStake struct {
	PublicAddress string         `json:"public_address"` // it should be the public address
	Percentage    string         `json:"percentage"`     // unstake percentage
	Ticker        string         `json:"ticker"`         // ticker token symbol
	Owner         sdk.AccAddress `json:"owner"`
}

// NewMsgSetUnStake is a constructor function for MsgSetPoolData
func NewMsgSetUnStake(publicAddress, percentage, ticker string, owner sdk.AccAddress) MsgSetUnStake {
	return MsgSetUnStake{
		PublicAddress: publicAddress,
		Percentage:    percentage,
		Ticker:        strings.ToUpper(ticker),
		Owner:         owner,
	}
}

// Route should return the pooldata of the module
func (msg MsgSetUnStake) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetUnStake) Type() string { return "set_unstake" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetUnStake) ValidateBasic() sdk.Error {
	if len(msg.Ticker) == 0 {
		return sdk.ErrUnknownRequest("Pool Ticker cannot be empty")
	}
	if len(msg.Percentage) == 0 {
		return sdk.ErrUnknownRequest("Percentage cannot be empty")
	}
	_, err := strconv.ParseFloat(msg.Percentage, 64)
	if nil != err {
		return sdk.ErrUnknownRequest("invalid percentage value")
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetUnStake) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetUnStake) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}
