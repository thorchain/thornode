package chainclients

import (
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients/binance"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients/ethereum"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/tss/go-tss/tss"
)

// LoadChains returns chain clients from chain configuration
func LoadChains(thorKeys *thorclient.Keys, cfg []config.ChainConfiguration, server *tss.TssServer, thorchainBridge *thorclient.ThorchainBridge, m *metrics.Metrics) map[common.Chain]ChainClient {
	chains := make(map[common.Chain]ChainClient, 0)

	for _, chain := range cfg {
		switch chain.ChainID {
		case common.BNBChain:
			bnb, err := binance.NewBinance(thorKeys, chain, server, thorchainBridge, m)
			if err != nil {
				continue
			}

			chains[common.BNBChain] = bnb
		case common.ETHChain:
			eth, err := ethereum.NewClient(thorKeys, chain, server, thorchainBridge, m)
			if err == nil {
				chains[common.ETHChain] = eth
			}
		default:
			continue
		}
	}

	return chains
}
