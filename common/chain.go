package common

import (
	"fmt"
	"strings"

	btypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/types"
)

var (
	BNBChain   = Chain("BNB")
	ETHChain   = Chain("ETH")
	BTCChain   = Chain("BTC")
	THORChain  = Chain("THOR")
	EmptyChain = Chain("")
)

// NoSigningAlgo empty signing algorithm
const NoSigningAlgo = keys.SigningAlgo("")

// Chain is the
type Chain string

// Chains represent a slice of Chain
type Chains []Chain

// NewChain create a new Chain and default the siging_algo to Secp256k1
func NewChain(chain string) (Chain, error) {

	if len(chain) < 3 {
		return EmptyChain, fmt.Errorf("chain error: not enough characters")
	}

	return Chain(strings.ToUpper(chain)), nil
}

// Equals compare two chain to see whether they represent the same chain
func (c Chain) Equals(c2 Chain) bool {
	return strings.EqualFold(c.String(), c2.String())
}

// IsEmpty is to determinate whether the chain is empty
func (c Chain) IsEmpty() bool {
	return strings.TrimSpace(c.String()) == ""
}

// String implement fmt.Stringer
func (c Chain) String() string {
	// convert it to upper case again just in case someone created a ticker via Chain("rune")
	return strings.ToUpper(string(c))
}

// IsBNB determinate whether it is BNBChain
func (c Chain) IsBNB() bool {
	return c.Equals(BNBChain)
}

// GetSigningAlgo get the signing algorithm for the given chain
func (c Chain) GetSigningAlgo() keys.SigningAlgo {
	switch c {
	case BNBChain, ETHChain, BTCChain, THORChain:
		return keys.Secp256k1
	}
	return keys.Secp256k1
}

// GetGasAsset chain's base asset
func (c Chain) GetGasAsset() Asset {
	switch c {
	case BNBChain:
		return BNBAsset
	case BTCChain:
		return BTCAsset
	case ETHChain:
		return ETHAsset
	default:
		return EmptyAsset
	}
}

// AddressPrefix return the address prefix used by the given network (testnet/mainnet)
func (c Chain) AddressPrefix(cn ChainNetwork) string {
	switch cn {
	case TestNet:
		switch c {
		case BNBChain:
			return btypes.TestNetwork.Bech32Prefixes()
		case THORChain:
			// TODO update this to use testnet address prefix
			return types.GetConfig().GetBech32AccountAddrPrefix()
		}
	case MainNet:
		switch c {
		case BNBChain:
			return btypes.ProdNetwork.Bech32Prefixes()
		case THORChain:
			return types.GetConfig().GetBech32AccountAddrPrefix()
		}
	}
	return ""
}

// Has check whether chain c is in the list
func (chains Chains) Has(c Chain) bool {
	for _, ch := range chains {
		if ch.Equals(c) {
			return true
		}
	}
	return false
}

// Distinct return a distinct set of chains , no duplicate
func (chains Chains) Distinct() Chains {
	var newChains Chains
	for _, chain := range chains {
		if !newChains.Has(chain) {
			newChains = append(newChains, chain)
		}
	}
	return newChains
}
