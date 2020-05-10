package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/tss/go-tss/blame"
)

// MsgTssKeysignFail means TSS keysign failed
type MsgTssKeysignFail struct {
	ID     string         `json:"id"`
	Height int64          `json:"height"`
	Blame  blame.Blame    `json:"blame"`
	Memo   string         `json:"memo"`
	Coins  common.Coins   `json:"coins"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgTssKeysignFail create a new instance of MsgTssKeysignFail message
func NewMsgTssKeysignFail(height int64, blame blame.Blame, memo string, coins common.Coins, signer sdk.AccAddress) MsgTssKeysignFail {
	return MsgTssKeysignFail{
		ID:     getMsgTssKeysignFailID(blame.BlameNodes, height, memo, coins),
		Height: height,
		Blame:  blame,
		Memo:   memo,
		Coins:  coins,
		Signer: signer,
	}
}

// getTssKeysignFailID this method will use all the members that caused the tss
// keysign failure , as well as the block height of the txout item to generate
// a hash, given that , if the same party keep failing the same txout item ,
// then we will only slash it once.
func getMsgTssKeysignFailID(members []blame.Node, height int64, memo string, coins common.Coins) string {
	// ensure input pubkeys list is deterministically sorted
	sort.SliceStable(members, func(i, j int) bool {
		return members[i].Pubkey < members[j].Pubkey
	})
	sb := strings.Builder{}
	for _, item := range members {
		sb.WriteString(item.Pubkey)
	}
	sb.WriteString(fmt.Sprintf("%d", height))
	sb.WriteString(memo)
	for _, c := range coins {
		sb.WriteString(c.String())
	}
	hash := sha256.New()
	return hex.EncodeToString(hash.Sum([]byte(sb.String())))
}

// Route should return the cmname of the module
func (msg MsgTssKeysignFail) Route() string { return RouterKey }

// Type should return the action
func (msg MsgTssKeysignFail) Type() string { return "set_tss_keysign_fail" }

// ValidateBasic runs stateless checks on the message
func (msg MsgTssKeysignFail) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if len(msg.ID) == 0 {
		return sdk.ErrUnknownRequest("ID cannot be blank")
	}
	if len(msg.Coins) == 0 {
		return sdk.ErrUnknownRequest("no coins")
	}
	for _, c := range msg.Coins {
		if err := c.IsValid(); err != nil {
			return sdk.ErrInvalidCoins(err.Error())
		}
	}
	if msg.Blame.IsEmpty() {
		return sdk.ErrUnknownRequest("tss blame is empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgTssKeysignFail) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgTssKeysignFail) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
