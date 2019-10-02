package smoke

import "fmt"

var endpoints map[string]string

type Statechain struct{
	Env string
}

func init() {
	endpoints["stage"] = "testnet-chain.bepswap.io"
	endpoints["dev"] = "testnet-chain.bepswap.net"
	endpoints["prod"] = "testnet-chain.bepswap.com"
}

func NewStatechain(env string) Statechain {
	return Statechain{
		Env: env,
	}
}

func (s Statechain) PoolURL(stage string) string {
	return fmt.Sprintf("https://%v/swapservice/pools", endpoints[s.Env])
}
