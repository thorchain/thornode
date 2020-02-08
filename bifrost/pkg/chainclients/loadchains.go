package chainclients

import (
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients/binance"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/common"
)

// LoadChains returns chain clients from chain configuration
func LoadChains(thorKeys *thorclient.Keys, cfg []config.ChainConfiguration, tss config.TSSConfiguration, thorchainBridge *thorclient.ThorchainBridge) map[common.Chain]ChainClient {
	chains := make(map[common.Chain]ChainClient, 0)

	for _, chain := range cfg {	
		switch chain.ChainID {
		case common.BNBChain:
			bnb, err := binance.NewBinance(thorKeys, chain.RPCHost, tss, thorchainBridge)
			if err == nil {
				chains[common.BNBChain] = bnb
			}
		default:
			continue
		}
	}

	return chains
}
