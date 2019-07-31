package silverback

import (
	"encoding/json"

	log "github.com/rs/zerolog/log"

	types "gitlab.com/thorchain/bepswap/observe/x/silverback/types"
	jungle "gitlab.com/thorchain/bepswap/observe/x/jungle"
)

type pool struct {
	PoolAddress string
	SymbolX string
	SymbolY string
}

func NewPool(poolAddress string) *pool {
	log.Info().Msgf("Initialising pool %s...", poolAddress)

	return &pool{
		PoolAddress: poolAddress,
		SymbolX: "RUNE-A1F",
		SymbolY: "BNB",
	}
}

func (p *pool) GetBalances() types.Balances {
	db := jungle.RedisClient()
	data, _ := db.Get("balances").Result()

	var balances types.Balances
	var tokens types.Tokens

	log.Info().Msgf("Data: %v", data)

	err := json.Unmarshal([]byte(data), &tokens)
	if err != nil {
		log.Error().Msgf("Error: %v", err)
		return balances
	}

	for _, coin := range tokens {
		if coin.Symbol == p.SymbolX {
			balances.X = coin.Free
		} else {
			balances.Y = coin.Free
		}
	}

	return balances
}
