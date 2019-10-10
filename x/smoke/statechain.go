package smoke

import "fmt"

var endpoints = map[string]string{
	"local": "localhost",
	"stage": "testnet-chain.bepswap.io",
	"dev":   "testnet-chain.bepswap.net",
	"prod":  "testnet-chain.bepswap.com",
}

type Statechain struct {
	Env string
}

// NewStatechain : Create a new Statechain instance.
func NewStatechain(env string) Statechain {
	return Statechain{
		Env: env,
	}
}

// PoolURL : Return the Pool URL based on the selected environment.
func (s Statechain) PoolURL() string {
	scheme := "https"

	if s.Env == "local" {
		scheme = "http"
	}

	return fmt.Sprintf("%v://%v/swapservice/pools", scheme, endpoints[s.Env])
}
