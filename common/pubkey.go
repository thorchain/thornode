package common

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcutil/bech32"
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
	switch chain {
	case BNBChain:
		str, err := bech32.Encode(chain.AddressPrefix(TestNetwork), pubKey)
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
	pKey, err := NewPubKeyFromHexString(s)
	if err != nil {
		return err
	}
	*pubKey = pKey
	return nil
}
