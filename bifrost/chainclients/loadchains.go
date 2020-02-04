package chainclients

import (
	"errors"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/chainclients/binance"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/common"
)

var (
	NotFound       = errors.New("not found")
	NotImplemented = errors.New("not implemented")
	NotSupported   = errors.New("not supported")
)

// LoadChains returns chain clients from chain configurations
func LoadChains(thorKeys *thorclient.Keys, cfg []config.ChainConfigurations, tss config.TSSConfiguration, thorchainBridge *thorclient.ThorchainBridge) map[common.Chain]ChainClient {
	chains := make(map[common.Chain]ChainClient, 0)

	for _, chain := range cfg {
		if !chain.Enabled {
			continue
		}

		switch chain.Name {
		case "BNB":
			bnb, err := loadBNBClient(thorKeys, chain, tss, thorchainBridge)
			if err == nil {
				chains[common.BNBChain] = bnb
			}
		case "ETH":
			eth, err := loadETHClient(thorKeys, chain, tss)
			if err == nil {
				chains[common.ETHChain] = eth
			}
		case "BTC":
			btc, err := loadBTCClient(thorKeys, chain, tss)
			if err == nil {
				chains[common.BTCChain] = btc
			}
		}
	}

	return chains
}

func NewBlockScannerStorage(observerDbPath string, chain ChainClient) (BlockScannerStorage, error) {
	switch chain.GetChain() {
	case common.BNBChain:
		return binance.NewBinanceBlockScannerStorage(observerDbPath)
	default:
		return nil, NotSupported
	}
}

func NewBlockScanner(cfg config.BlockScannerConfiguration, scanStorage blockscanner.ScannerStorage, chain ChainClient, isTestNet bool, pkmgr pubkeymanager.PubKeyValidator, m *metrics.Metrics) (BlockScanner, error) {
	switch chain.GetChain() {
	case common.BNBChain:
		return binance.NewBinanceBlockScanner(cfg, scanStorage, isTestNet, pkmgr, m)
	default:
		return nil, NotSupported
	}
}

func loadBTCClient(thorKeys *thorclient.Keys, chain config.ChainConfigurations, tss config.TSSConfiguration) (ChainClient, error) {
	return nil, NotImplemented
}

func loadBNBClient(thorKeys *thorclient.Keys, chain config.ChainConfigurations, tss config.TSSConfiguration, thorchainBridge *thorclient.ThorchainBridge) (ChainClient, error) {
	return binance.NewBinance(thorKeys, chain.RPCHost, tss, thorchainBridge)
}

func loadETHClient(thorKeys *thorclient.Keys, chain config.ChainConfigurations, tss config.TSSConfiguration) (ChainClient, error) {
	return nil, NotImplemented
}
