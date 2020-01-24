package blockclients

import (
	"errors"
	"strings"

	"gitlab.com/thorchain/thornode/bifrost/config"
)

// LoadChains returns block chain clients from chain configurations
func LoadChains(cfgChains []config.ChainConfigurations) []BlockChainClient {
	var chains []BlockChainClient

	for _, chain := range cfgChains {
		if !chain.Enabled {
			continue
		}

		switch strings.ToLower(chain.Name) {
		case "bnb":
			bnb, err := loadBnbClient(chain)
			if err == nil {
				chains = append(chains, bnb)
			}
		case "eth":
			eth, err := loadEthClient(chain)
			if err == nil {
				chains = append(chains, eth)
			}
		case "btc":
			btc, err := loadBtcClient(chain)
			if err == nil {
				chains = append(chains, btc)
			}
		}
	}

	return chains
}

func loadBtcClient(cfg config.ChainConfigurations) (BlockChainClient, error) {
	return nil, errors.New("not implemented")
}

func loadBnbClient(cfg config.ChainConfigurations) (BlockChainClient, error) {
	return nil, errors.New("not implemented")
}

func loadEthClient(cfg config.ChainConfigurations) (BlockChainClient, error) {
	return nil, errors.New("not implemented")
}
