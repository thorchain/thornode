package chainclients

import (
	"errors"
	"strings"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/chainclients/binance"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
)

var (
	NotSupported   = errors.New("not supported")
	NotImplemented = errors.New("not implemented")
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

func NewBlockScannerStorage(observerDbPath, chain string) (BlockScannerStorage, error) {
	switch chain {
	case "bnb":
		return binance.NewBinanceBlockScannerStorage(observerDbPath)
	default:
		return nil, NotSupported
	}
}

func NewBlockScanner(cfg config.BlockScannerConfiguration, scanStorage blockscanner.ScannerStorage, chain string, isTestNet bool, pkmgr pubkeymanager.PubKeyValidator, m *metrics.Metrics) (BlockScanner, error) {
	switch chain {
	case "bnb":
		return binance.NewBinanceBlockScanner(cfg, scanStorage, isTestNet, pkmgr, m)
	default:
		return nil, NotSupported
	}
}

func loadBTCClient(cfg config.ChainConfigurations) (ChainClient, error) {
	return nil, NotImplemented
}

func loadBNBClient(cfg config.ChainConfigurations) (ChainClient, error) {
	return nil, NotImplemented
}

func loadETHClient(cfg config.ChainConfigurations) (ChainClient, error) {
	return nil, NotImplemented
}
