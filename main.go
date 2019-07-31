package main

import (
	"os"
	"encoding/json"

	log "github.com/rs/zerolog/log"

	jungle "gitlab.com/thorchain/bepswap/observe/x/jungle"
	silverback "gitlab.com/thorchain/bepswap/observe/x/silverback"
)

func main() {
	db := jungle.RedisClient()
	bal, _ := db.Get("balances").Result()
	log.Info().Msgf("Current (Saved) Balances: %v", bal)

	binance := silverback.NewBinance()
	json, _ := json.Marshal(binance.GetAccount().Balances)

	err := db.Set("balances", json, 0).Err()
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
		os.Exit(1)
	}

	log.Info().Msg("Balances updated.")

	silverback.NewServer(*binance).Start()
	silverback.NewClient(*binance).Start()
}
