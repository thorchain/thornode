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

const NoSigningAlgo = keys.SigningAlgo("")

// Chain is the
type Chain string
type Chains []Chain

// NewChain create a new Chain and default the siging_algo to Secp256k1
func NewChain(chain string) (Chain, error) {
	noChain := Chain("")
	if len(chain) < 3 {
		return noChain, fmt.Errorf("Chain Error: Not enough characters")
	}

	return Chain(strings.ToUpper(chain)), nil
}

func (c Chain) Equals(c2 Chain) bool {
	return strings.EqualFold(c.String(), c2.String())
}

// IsEmpty is to determinate whether the chain is empty
func (c Chain) IsEmpty() bool {
	return strings.TrimSpace(c.String()) == ""
}

func (c Chain) String() string {
	// uppercasing again just incase someon created a ticker via Chain("rune")
	return strings.ToUpper(string(c))
}

func (c Chain) IsBNB() bool {
	return c.Equals(BNBChain)
}

func (c Chain) GetSigningAlgo() keys.SigningAlgo {
	switch c {
	case BNBChain, ETHChain, BTCChain, THORChain:
		return keys.Secp256k1
	}
	return keys.Secp256k1
}

// AddressPrefix return the address prefix used by the given network (testnet/mainnet)
func (c Chain) AddressPrefix(cn ChainNetwork) string {
	switch cn {
	case TestNet:
		switch c {
		case ETHChain:
			// TODO add support
		case BTCChain:
			// TODO add support
		case BNBChain:
			return btypes.TestNetwork.Bech32Prefixes()
		case THORChain:
			// TODO update this to use testnet address prefix
			return types.GetConfig().GetBech32AccountAddrPrefix()
		}
	case MainNet:
		switch c {
		case BTCChain:
			// TODO add support
		case ETHChain:
			// TODO Add support
		case BNBChain:
			return btypes.ProdNetwork.Bech32Prefixes()
		case THORChain:
			return types.GetConfig().GetBech32AccountAddrPrefix()
		}
	}
	return ""
}

func (chains Chains) Has(c Chain) bool {
	for _, ch := range chains {
		if ch.Equals(c) {
			return true
		}
	}
	return false
}

func (chains Chains) Uniquify() Chains {
	var newChains Chains
	for _, chain := range chains {
		if !newChains.Has(chain) {
			newChains = append(newChains, chain)
		}
	}
	return newChains
}
