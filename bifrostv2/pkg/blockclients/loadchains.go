package blockclients

import (
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/pkg/blockclients/bnb"
	"gitlab.com/thorchain/thornode/bifrostv2/pkg/blockclients/btc"
	"gitlab.com/thorchain/thornode/bifrostv2/pkg/blockclients/eth"
)

// LoadChains returns block chain clients from chain configurations
func LoadChains(cfgChains []config.ChainConfigurations) []BlockChainClient {
	var chains []BlockChainClient

	for _, chain := range cfgChains {
		if chain.Name == "bnb" && chain.Enabled {
			bnb := loadBnbClient(chain)
			if bnb != nil {
				chains = append(chains, bnb)
			}
		}

		if chain.Name == "eth" && chain.Enabled {
			eth := loadEthClient(chain)
			if eth != nil {
				chains = append(chains, eth)
			}
		}

		if chain.Name == "btc" && chain.Enabled {
			btc := loadBtcClient(chain)
			if btc != nil {
				chains = append(chains, btc)
			}
		}
	}
	return chains
}

func loadBtcClient(cfg config.ChainConfigurations) BlockChainClient {
	btcClient, err := btc.NewClient(cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to load btcClient")
		return nil
	}
	log.Debug().Msg("loadBTCClient")
	return btcClient
}

func loadBnbClient(cfg config.ChainConfigurations) BlockChainClient {
	bnbClient, err := bnb.NewClient(cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to load bnbClient")
		return nil
	}
	log.Debug().Msg("loadBNBClient")
	return bnbClient
}

func loadEthClient(cfg config.ChainConfigurations) BlockChainClient {
	ethClient, err := eth.NewClient(cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to load ethClient")
		return nil
	}

	log.Debug().Msg("loadETHClient")
	return ethClient
}
