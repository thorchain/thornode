package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	blame "gitlab.com/thorchain/tss/go-tss/blame"

	"gitlab.com/thorchain/thornode/common"
)

// MsgTssPool defines a MsgTssPool message
type MsgTssPool struct {
	ID         string         `json:"id"`
	PoolPubKey common.PubKey  `json:"pool_pub_key"`
	KeygenType KeygenType     `json:"keygen_type"`
	PubKeys    common.PubKeys `json:"pubkeys"`
	Height     int64          `json:"height"`
	Blame      blame.Blame    `json:"blame"`
	Chains     common.Chains  `json:"chains"`
	Signer     sdk.AccAddress `json:"signer"`
}

// NewMsgTssPool is a constructor function for MsgTssPool
func NewMsgTssPool(pks common.PubKeys, poolpk common.PubKey, KeygenType KeygenType, height int64, blame blame.Blame, chains common.Chains, signer sdk.AccAddress) MsgTssPool {
	return MsgTssPool{
		ID:         getTssID(pks, poolpk, height),
		PubKeys:    pks,
		PoolPubKey: poolpk,
		Height:     height,
		KeygenType: KeygenType,
		Blame:      blame,
		Chains:     chains,
		Signer:     signer,
	}
}

// getTssID
func getTssID(members common.PubKeys, poolPk common.PubKey, height int64) string {
	// ensure input pubkeys list is deterministically sorted
	sort.SliceStable(members, func(i, j int) bool {
		return members[i].String() < members[j].String()
	})
	sb := strings.Builder{}
	for _, item := range members {
		sb.WriteString(item.String())
	}
	sb.WriteString(poolPk.String())
	sb.WriteString(fmt.Sprintf("%d", height))
	hash := sha256.New()
	return hex.EncodeToString(hash.Sum([]byte(sb.String())))
}

// Route should return the cmname of the module
func (msg MsgTssPool) Route() string { return RouterKey }

// Type should return the action
func (msg MsgTssPool) Type() string { return "set_tss_pool" }

// ValidateBasic runs stateless checks on the message
func (msg MsgTssPool) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if len(msg.ID) == 0 {
		return sdk.ErrUnknownRequest("ID cannot be blank")
	}
	if len(msg.PubKeys) < 2 {
		return sdk.ErrUnknownRequest("Must have at least 2 pub keys")
	}
	if len(msg.PubKeys) > 100 {
		return sdk.ErrUnknownRequest("Must have no more then 100 pub keys")
	}
	for _, pk := range msg.PubKeys {
		if pk.IsEmpty() {
			return sdk.ErrUnknownRequest("Pubkey cannot be empty")
		}
	}
	// PoolPubKey can't be empty only when keygen success
	if msg.IsSuccess() {
		if msg.PoolPubKey.IsEmpty() {
			return sdk.ErrUnknownRequest("Pool pubkey cannot be empty")
		}
	}
	// ensure pool pubkey is a valid bech32 pubkey
	if _, err := common.NewPubKey(msg.PoolPubKey.String()); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}

	// ensure we have rune chain
	found := false
	for _, chain := range msg.Chains {
		if chain.Equals(common.RuneAsset().Chain) {
			found = true
			break
		}
	}
	if !found {
		return sdk.ErrUnknownRequest("must support rune asset chain")
	}

	// ensure there are no duplicate chains
	chainDup := make(map[common.Chain]int, 0)
	for _, chain := range msg.Chains {
		if _, ok := chainDup[chain]; !ok {
			chainDup[chain] = 0
		}
		chainDup[chain]++
	}
	for k, v := range chainDup {
		if v > 1 {
			return sdk.ErrUnknownRequest(fmt.Sprintf("cannot have duplicate chains (%s)", k.String()))
		}
	}

	return nil
}

func (msg MsgTssPool) IsSuccess() bool {
	return msg.Blame.IsEmpty()
}

// GetSignBytes encodes the message for signing
func (msg MsgTssPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgTssPool) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
