package common

import (
	"strings"

	"github.com/btcsuite/btcutil/bech32"
)

type Address string

var NoAddress Address = Address("")

// NewAddress create a new Address
// Sample: bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6
func NewAddress(address string) (Address, error) {
	if len(address) == 0 {
		return NoAddress, nil
	}

	_, _, err := bech32.Decode(address)
	if err != nil {
		return NoAddress, err
	}

	return Address(address), nil
}

func (addr Address) IsChain(chain Chain) bool {
	switch chain {
	case BNBChain:
		prefix, _, _ := bech32.Decode(addr.String())
		return prefix == "bnb" || prefix == "tbnb"
	default:
		return true // if we don't specifically check a chain yet, assume its ok.
	}
}

func (addr Address) Equals(addr2 Address) bool {
	return strings.EqualFold(addr.String(), addr2.String())
}

func (addr Address) IsEmpty() bool {
	return strings.TrimSpace(addr.String()) == ""
}

func (addr Address) String() string {
	return string(addr)
}
