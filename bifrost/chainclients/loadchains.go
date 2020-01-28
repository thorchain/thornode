package chainclients

import (
	"errors"
	"strings"

	"gitlab.com/thorchain/thornode/bifrost/config"
)

// LoadChains returns chain clients from chain configurations
func LoadChains(cfgChains []config.ChainConfigurations) []ChainClient {
	var chains []ChainClient

	for _, chain := range cfgChains {
		if !chain.Enabled {
			continue
		}

		switch strings.ToLower(chain.Name) {
		case "bnb":
			bnb, err := loadBNBClient(chain)
			if err == nil {
				chains = append(chains, bnb)
			}
		case "eth":
			eth, err := loadETHClient(chain)
			if err == nil {
				chains = append(chains, eth)
			}
		case "btc":
			btc, err := loadBTCClient(chain)
			if err == nil {
				chains = append(chains, btc)
			}
		}
	}

	return chains
}

func loadBTCClient(cfg config.ChainConfigurations) (ChainClient, error) {
	return nil, errors.New("not implemented")
}

func loadBNBClient(cfg config.ChainConfigurations) (ChainClient, error) {
	return nil, errors.New("not implemented")
}

func loadETHClient(cfg config.ChainConfigurations) (ChainClient, error) {
	return nil, errors.New("not implemented")
}
