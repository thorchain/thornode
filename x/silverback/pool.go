package silverback

import (
	"encoding/json"

	log "github.com/rs/zerolog/log"

	types "gitlab.com/thorchain/bepswap/observe/x/silverback/types"
	jungle "gitlab.com/thorchain/bepswap/observe/x/jungle"
)

type Pool struct {
	PoolAddress string
	SymbolX string
	SymbolY string
}

func NewPool(poolAddress string) *Pool {
	log.Info().Msgf("Initialising pool %s...", poolAddress)

	return &Pool{
		PoolAddress: poolAddress,
		SymbolX: "RUNE-A1F",
		SymbolY: "BNB",
	}
}

func (p *Pool) GetBalances() types.Balances {
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

// ( x * Y ) / ( x + X )
func (p *Pool) CalcOutput() {}

// ( x ) / ( x + X )
func (p *Pool) CalcOutputSlip() {}

// ( x^2 *  Y ) / ( x + X )^2
func (p *Pool) CalcLiquidityFee() {}

// ( x * X * Y ) / ( x + X )^2
func (p *Pool) CalcTokensEmitted() {}

// x * ( 2X + x) / ( x + X )^2
func (p *Pool) CalcTradeSlip() {}

func (p *Pool) CalcBalance() {}

// x * ( 2X + x) / ( X * X )
func (p *Pool) CalcPoolSlip() {}
