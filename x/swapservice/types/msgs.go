package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const RouterKey = ModuleName // this was defined in your key.go file

// MsgSetAccData defines a SetAccData message
// TODO: given we don't hold asset ,thus we don't need to set Account Data
type MsgSetAccData struct {
	AccID  string         `json:"acc_id"`
	Name   string         `json:"name"`
	Ticker string         `json:"ticker"`
	Amount string         `json:"token"`
	Owner  sdk.AccAddress `json:"owner"`
}

// NewMsgSetAccData is a constructor function for MsgSetAccData
func NewMsgSetAccData(name, ticker, amount string, owner sdk.AccAddress) MsgSetAccData {
	return MsgSetAccData{
		AccID:  fmt.Sprintf("acc-%s", strings.ToLower(name)),
		Name:   strings.ToLower(name),
		Ticker: ticker,
		Amount: amount,
		Owner:  owner,
	}
}

// Route should return the pooldata of the module
func (msg MsgSetAccData) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetAccData) Type() string { return "set_accdata" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetAccData) ValidateBasic() sdk.Error {
	if len(msg.Name) == 0 {
		return sdk.ErrUnknownRequest("Account Name cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetAccData) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetAccData) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

// MsgStake defines a MsgStake message
// TODO don't need it,and it is not used either , for remove
type MsgStake struct {
	Name        string         `json:"name"`
	Ticker      string         `json:"ticker"`
	AtomAmount  string         `json:"atom_amount"`
	TokenAmount string         `json:"token_amount"`
	Owner       sdk.AccAddress `json:"owner"`
}
