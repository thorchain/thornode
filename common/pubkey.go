package common

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/types"
)

// PubKey used in statechain
type PubKey []byte

// EmptyPubKey
var EmptyPubKey PubKey

// NewPubKey create a new instance of PubKey
func NewPubKey(b []byte) PubKey {
	return PubKey(b)
}

// NewPubKeyFromHexString decode
func NewPubKeyFromHexString(key string) (PubKey, error) {
	buf, err := hex.DecodeString(key)
	if nil != err {
		return nil, fmt.Errorf("fail to decode hex string,err:%w", err)
	}
	return PubKey(buf), nil
}

func NewPubKeyFromBech32(key, prefix string) (PubKey, error) {
	buf, err := types.GetFromBech32(key, prefix)
	if nil != err {
		return EmptyPubKey, fmt.Errorf("fail to decode pub key from bech 32")
	}
	return NewPubKey(buf), nil
}

// Equals check whether two are the same
func (pubKey PubKey) Equals(pubKey1 PubKey) bool {
	return bytes.Equal(pubKey, pubKey1)
}

// IsEmpty to check whether it is empty
func (pubKey PubKey) IsEmpty() bool {
	return len(pubKey) == 0
}

// String stringer implementation
func (pubKey PubKey) String() string {
	return hex.EncodeToString(pubKey)
}

// GetAddress will return an address for the given chain
func (pubKey PubKey) GetAddress(chain Chain) (Address, error) {
	if pubKey.IsEmpty() {
		return NoAddress, nil
	}
	chainNetwork := GetCurrentChainNetwork()
	switch chain {
	case BNBChain:
		str, err := ConvertAndEncode(chain.AddressPrefix(chainNetwork), pubKey)
		if nil != err {
			return NoAddress, fmt.Errorf("fail to bech32 encode the address, err:%w", err)
		}
		return NewAddress(str)
	}

	return NoAddress, nil
}

// MarshalJSON to Marshals to JSON using Bech32
func (pubKey PubKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(pubKey.String())
}

// UnmarshalJSON to Unmarshal from JSON assuming Bech32 encoding
func (pubKey *PubKey) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}
	// this it to make our genesis easier, cause we can get the public key from tbnbc
	chainNetWork := GetCurrentChainNetwork()
	if strings.HasPrefix(s, BNBChain.AddressPrefix(chainNetWork)) {
		pKey, err := NewPubKeyFromBech32(s, BNBChain.AddressPrefix(chainNetWork))
		if nil != err {
			return err
		}
		*pubKey = pKey
		return nil
	}
	pKey, err := NewPubKeyFromHexString(s)
	if err != nil {
		return err
	}
	*pubKey = pKey
	return nil
}

// ConvertAndEncode converts from a base64 encoded byte string to base32 encoded byte string and then to bech32
func ConvertAndEncode(hrp string, data []byte) (string, error) {
	converted, err := bech32.ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("encoding bech32 failed,%w", err)
	}
	return bech32.Encode(hrp, converted)
}
