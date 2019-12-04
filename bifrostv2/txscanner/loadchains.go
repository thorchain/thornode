package txscanner

import (
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/pkg/blockclients/btc"
	"gitlab.com/thorchain/thornode/bifrostv2/pkg/blockclients/eth"
)

func loadChains(cfg config.TxScannerConfigurations) []BlockChainClients {
	var chains []BlockChainClients
	if cfg.BlockChains.BNB.Enabled {
		// bnb := loadBnbClient(cfg.BlockChains.BNB)
		// if bnb != nil {
		// 	chains = append(chains, bnb)
		// }
	}
	if cfg.BlockChains.ETH.Enabled {
		eth := loadEthClient(cfg.BlockChains.ETH)
		if eth != nil {
			chains = append(chains, eth)
		}
	}
	if cfg.BlockChains.BTC.Enabled {
		btc := loadBitcoinClient(cfg.BlockChains.BTC)
		if btc != nil {
			chains = append(chains, btc)
		}
	}
	return chains
}

func loadBitcoinClient(cfg config.BTCConfiguration) BlockChainClients {
	bitcoinClient, err := btc.NewClient(cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to load bitcoinClient")
		return nil
	}
	log.Debug().Msg("loadBitcoinClient")
	return bitcoinClient
}

// func loadBnbClient(cfg config.BNBConfiguration) BlockChainClients {
// 	binanceClient, err := binanceChain.NewClient(cfg)
// 	if err != nil {
// 		log.Error().Err(err).Msg("failed to load binacneClient")
// 		return nil
// 	}
// 	return binanceClient
// }

func loadEthClient(cfg config.ETHConfiguration) BlockChainClients {
	etherumClient, err := eth.NewClient(cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to load etherumClient")
		return nil
	}

	log.Debug().Msg("loadEthClient")
	return etherumClient
}
