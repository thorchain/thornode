package smoke

import "fmt"

var endpoints = map[string]string{
	"local":      "localhost:1317",
	"staging":    "testnet-chain.bepswap.io",
	"develop":    "testnet-chain.bepswap.net",
	"production": "testnet-chain.bepswap.com",
}

type Thorchain struct {
	Env string
}

// NewStatechain : Create a new Statechain instance.
func NewThorchain(env string) Thorchain {
	return Thorchain{
		Env: env,
	}
}

// Scheme : SSL or not.
func (t Thorchain) scheme() string {
	scheme := "https"

	if t.Env == "local" {
		scheme = "http"
	}

	return scheme
}

// PoolURL : Return the Pool URL based on the selected environment.
func (t Thorchain) PoolURL() string {
	return fmt.Sprintf("%v://%v/thorchain/pools", t.scheme(), endpoints[t.Env])
}

// StakerURL  : Return the Staker URL based on the selected environment.
func (t Thorchain) StakerURL(staker string) string {
	return fmt.Sprintf("%v://%v/thorchain/staker/%v", t.scheme(), endpoints[t.Env], staker)
}
