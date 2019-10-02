package smoke

import "fmt"

var endpoints = map[string]string{
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
	return fmt.Sprintf("https://%v/swapservice/pools", endpoints[s.Env])
}
