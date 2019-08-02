package silverback

import (
	"encoding/json"
	"math"

	log "github.com/rs/zerolog/log"

	jungle "gitlab.com/thorchain/bepswap/observe/x/jungle"
	types "gitlab.com/thorchain/bepswap/observe/x/silverback/types"
)

type Pool struct {
	PoolAddress string
	X string
	Y string
}

func NewPool(poolAddress string, X string, Y string) *Pool {
	return &Pool{
		PoolAddress: poolAddress,
		X: X,
		Y: Y,
	}
}

func (p *Pool) GetBal() types.Balances {
	db := jungle.RedisClient()
	data, _ := db.Get("balances").Result()

	var balances types.Balances
	var tokens types.Tokens

	err := json.Unmarshal([]byte(data), &tokens)
	if err != nil {
		log.Error().Msgf("Error: %v", err)
		return balances
	}

	for _, coin := range tokens {
		if coin.Symbol == p.X {
			balances.X = coin.Free
		} else {
			balances.Y = coin.Free
		}
	}

	return balances
}

// ( x * Y ) / ( x + X )
func (p *Pool) CalcOutput(x, X, Y float64) float64 {
	return ((x*Y)/(x+X))
}

// ( x ) / ( x + X )
func (p *Pool) CalcOutputSlip(x, X float64) float64 {
	return (x/(x+X))
}

// ( x^2 *  Y ) / ( x + X )^2
func (p *Pool) CalcLiquidityFee(x, X, Y float64) float64 {
	return ((math.Pow(x, 2)*Y)/math.Pow((x+X), 2))
}

// ( x * X * Y ) / ( x + X )^2
func (p *Pool) CalcTokensEmitted(x, X, Y float64) float64 {
	return ((x*X*Y)/math.Pow((x+X), 2))
}

// x * ( 2X + x) / ( x + X )^2
func (p *Pool) CalcTradeSlip(x, X, Y float64) float64 {
	return (x*((2*X)+x)/(math.Pow((x+X), 2)))
}

func (p *Pool) CalcBalance() {}

// x * ( 2X + x) / ( X * X )
func (p *Pool) CalcPoolSlip(x, X, Y float64) float64 {
	return (x *((2*X)+x)/(X*X))
}

func SyncBal(binance Binance) {
	db := jungle.RedisClient()
	log.Info().Msgf("Balances: %v", binance.GetAccount().Balances)
	balances, _ := json.Marshal(binance.GetAccount().Balances)

	err := db.Set("balances", balances, 0).Err()
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
	}
}
