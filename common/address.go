package common

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/bech32"
)

type Address string

var (
	NoAddress    Address = Address("")
	RagnarokAddr Address = Address("Ragnarok")
)

// NewAddress create a new Address
// Sample: bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6
func NewAddress(address string) (Address, error) {
	if len(address) == 0 {
		return NoAddress, nil
	}

	// Check is eth address
	if strings.HasPrefix(address, "0x") {
		if len(address) != 42 {
			return NoAddress, fmt.Errorf("0x address must be 42 characters (%d/42)", len(address))
		}
		return Address(address), nil
	}

	_, _, err := bech32.Decode(address)
	if err != nil {
		return NoAddress, err
	}

	return Address(address), nil
}

func (addr Address) IsChain(chain Chain) bool {
	switch chain {
	case ETHChain:
		return strings.HasPrefix(addr.String(), "0x")
	case BNBChain:
		prefix, _, _ := bech32.Decode(addr.String())
		return prefix == "bnb" || prefix == "tbnb"
	case THORChain:
		prefix, _, _ := bech32.Decode(addr.String())
		return prefix == "thor" || prefix == "tthor"
	default:
		return true // if THORNode don't specifically check a chain yet, assume its ok.
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
