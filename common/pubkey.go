package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
)

// PubKey used in statechain, it should be bech32 encoded string
// thus it will be something like
// thorpub1addwnpepqt7qug8vk9r3saw8n4r803ydj2g3dqwx0mvq5akhnze86fc536xcy2cr8a2
// thorpub1addwnpepqdqvd4r84lq9m54m5kk9sf4k6kdgavvch723pcgadulxd6ey9u70kgjgrwl
type PubKey string

// EmptyPubKey
var EmptyPubKey PubKey

// EmptyPubKeys
var EmptyPubKeys PubKeys

// PubKeys contains two pub keys , secp256k1 and ed25519
type PubKeys struct {
	Secp256k1 PubKey `json:"secp256k1"`
	Ed25519   PubKey `json:"ed25519"`
}

// NewPubKey create a new instance of PubKey
// key is bech32 encoded string
func NewPubKey(key string) (PubKey, error) {
	if len(key) == 0 {
		return EmptyPubKey, nil
	}
	_, err := sdk.GetAccPubKeyBech32(key)
	if nil != err {
		return EmptyPubKey, fmt.Errorf("%s is not bech32 encoded pub key,err : %w", key, err)
	}
	return PubKey(key), nil
}

// NewPubKeyFromCrypto
func NewPubKeyFromCrypto(pk crypto.PubKey) (PubKey, error) {
	s, err := sdk.Bech32ifyAccPub(pk)
	if nil != err {
		return EmptyPubKey, fmt.Errorf("fail to create PubKey from crypto.PubKey,err:%w", err)
	}
	return PubKey(s), nil
}

// Equals check whether two are the same
func (pubKey PubKey) Equals(pubKey1 PubKey) bool {
	return pubKey == pubKey1
}

// IsEmpty to check whether it is empty
func (pubKey PubKey) IsEmpty() bool {
	return len(pubKey) == 0
}

// String stringer implementation
func (pubKey PubKey) String() string {
	return string(pubKey)
}

// GetAddress will return an address for the given chain
func (pubKey PubKey) GetAddress(chain Chain) (Address, error) {
	if pubKey.IsEmpty() {
		return NoAddress, nil
	}
	chainNetwork := GetCurrentChainNetwork()
	switch chain {
	case BNBChain:
		pk, err := sdk.GetAccPubKeyBech32(string(pubKey))
		if nil != err {
			return NoAddress, err
		}
		str, err := ConvertAndEncode(chain.AddressPrefix(chainNetwork), pk.Address().Bytes())
		if nil != err {
			return NoAddress, fmt.Errorf("fail to bech32 encode the address, err:%w", err)
		}
		return NewAddress(str)
	case THORChain:
		pk, err := sdk.GetAccPubKeyBech32(string(pubKey))
		if nil != err {
			return NoAddress, err
		}
		str, err := ConvertAndEncode(chain.AddressPrefix(chainNetwork), pk.Address().Bytes())
		if nil != err {
			return NoAddress, fmt.Errorf("fail to bech32 encode the address, err:%w", err)
		}
		return NewAddress(str)
	}

	return NoAddress, nil
}

func (pubKey PubKey) GetThorAddress() (sdk.AccAddress, error) {
	addr, err := pubKey.GetAddress(THORChain)
	if err != nil {
		return nil, err
	}
	return sdk.AccAddressFromBech32(addr.String())
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
		return err
	}
	if strings.HasPrefix(s, "bnbp") {
		buf, err := sdk.GetFromBech32(s, "bnbp")
		if nil != err {
			return fmt.Errorf("fail to get from bech32 ,err:%w", err)
		}
		pk, err := cryptoAmino.PubKeyFromBytes(buf)
		if nil != err {
			return fmt.Errorf("fail to create pub key from bytes,err:%w", err)
		}
		s, err = sdk.Bech32ifyAccPub(pk)
		if nil != err {
			return fmt.Errorf("fail to bech32 acc pub:%w", err)
		}
	}
	pk, err := NewPubKey(s)
	if nil != err {
		return err
	}
	*pubKey = pk
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

// NewPubKeys create a new instance of PubKeys , which contains two keys
func NewPubKeys(secp256k1, ed25519 PubKey) PubKeys {
	return PubKeys{
		Secp256k1: secp256k1,
		Ed25519:   ed25519,
	}
}

// IsEmpty will determinate whether PubKeys is an empty
func (pks PubKeys) IsEmpty() bool {
	return pks.Secp256k1.IsEmpty() && pks.Ed25519.IsEmpty()
}

// Equals check whether two PubKeys are the same
func (pks PubKeys) Equals(pks1 PubKeys) bool {
	return pks.Ed25519.Equals(pks1.Ed25519) && pks.Secp256k1.Equals(pks1.Secp256k1)
}

func (pks PubKeys) Contains(pk PubKey) bool {
	return pks.Ed25519.Equals(pk) || pks.Secp256k1.Equals(pk)
}

// String implement fmt.Stinger
func (pks PubKeys) String() string {
	return fmt.Sprintf(`
	secp256k1: %s
	ed25519: %s
`, pks.Ed25519.String(), pks.Ed25519.String())
}

// GetAddress
func (pks PubKeys) GetAddress(chain Chain) (Address, error) {
	switch chain.GetSigningAlgo() {
	case keys.Secp256k1:
		return pks.Secp256k1.GetAddress(chain)
	case keys.Ed25519:
		return pks.Ed25519.GetAddress(chain)
	}
	return NoAddress, fmt.Errorf("unknow signing algorithm")
}
