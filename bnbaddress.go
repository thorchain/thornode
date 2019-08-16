package common

import (
	"fmt"
	"strings"
)

type BnbAddress string

var NoBnbAddress BnbAddress = BnbAddress("")

// NewBnbAddress create a new BnbAddress
// Sample: bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6
func NewBnbAddress(address string) (BnbAddress, error) {
	if len(address) == 0 {
		return NoBnbAddress, nil
	}
	prefixes := []string{"bnb", "tbnb"}

	// check if our address has one of the prefixes above
	hasPrefix := false
	for _, pref := range prefixes {
		if strings.HasPrefix(address, pref) {
			hasPrefix = true
			break
		}
	}
	if !hasPrefix {
		return "", fmt.Errorf("Address prefix is not supported")
	}

	// trim the prefix from our address
	var suffix string
	for _, pref := range prefixes {
		if strings.HasPrefix(address, pref) {
			suffix = address[len(pref):]
			break
		}
	}

	// check address length is valid
	if len(suffix) != 39 {
		return "", fmt.Errorf("Address length is not correct: %s (%d != 39)", suffix, len(suffix))
	}

	return BnbAddress(address), nil
}

func (bnb BnbAddress) Equals(bnb2 BnbAddress) bool {
	return strings.EqualFold(bnb.String(), bnb2.String())
}

func (bnb BnbAddress) IsEmpty() bool {
	return strings.TrimSpace(bnb.String()) == ""
}

func (bnb BnbAddress) String() string {
	return string(bnb)
}
