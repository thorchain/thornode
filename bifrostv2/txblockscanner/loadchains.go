package txblockscanner

import (
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/pkg/blockclients/bnb"
	"gitlab.com/thorchain/thornode/bifrostv2/pkg/blockclients/btc"
	"gitlab.com/thorchain/thornode/bifrostv2/pkg/blockclients/eth"
)

func loadChains(cfg config.TxScannerConfigurations) []BlockChainClients {
	var chains []BlockChainClients

	for _, chain := range cfg.BlockChains {
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

func loadBtcClient(cfg config.ChainConfigurations) BlockChainClients {
	btcClient, err := btc.NewClient(cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to load btcClient")
		return nil
	}
	log.Debug().Msg("loadBTCClient")
	return btcClient
}

func loadBnbClient(cfg config.ChainConfigurations) BlockChainClients {
	bnbClient, err := bnb.NewClient(cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to load bnbClient")
		return nil
	}
	log.Debug().Msg("loadBNBClient")
	return bnbClient
}

func loadEthClient(cfg config.ChainConfigurations) BlockChainClients {
	ethClient, err := eth.NewClient(cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to load ethClient")
		return nil
	}

	log.Debug().Msg("loadETHClient")
	return ethClient
}
