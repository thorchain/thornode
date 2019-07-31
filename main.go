package main

import (
	"fmt"
	"os"
	"time"
	"encoding/json"

	log "github.com/rs/zerolog/log"

	"github.com/binance-chain/go-sdk/keys"
	sdk "github.com/binance-chain/go-sdk/client"
	types "github.com/binance-chain/go-sdk/common/types"

	silverback "gitlab.com/thorchain/bepswap/observe/x/silverback"
	jungle "gitlab.com/thorchain/bepswap/observe/x/jungle"
)

func main() {
	db := jungle.RedisClient()
	bal, _ := db.Get("balances").Result()
	log.Info().Msgf("Current (Saved) Balances: %v", bal)

	port := os.Getenv("PORT")
	server := silverback.NewServer(port)
	server.Start()

	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal().Msg("No private key set!")
		Exit()
	}

	keyManager, err := keys.NewPrivateKeyManager(privateKey)
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
		Exit()
	}

	bClient, err := sdk.NewDexClient(os.Getenv("DEX_HOST"), types.TestNetwork, keyManager)
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
		Exit()
	}

	log.Info().Msgf("Using address: %s", keyManager.GetAddr())

	fmtAddr := fmt.Sprintf("%s", keyManager.GetAddr())
	account, err := bClient.GetAccount(fmtAddr)
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
		Exit()
	}

	json, _ := json.Marshal(account.Balances)
	err = db.Set("balances", json, 0).Err()
	if err != nil {
		panic(err)
	}

	log.Info().Msg("Balances updated.")

	dexHost := os.Getenv("DEX_HOST")
	client := silverback.NewClient(30 * time.Second, fmtAddr, dexHost)
	client.Start()
}

func Exit() {
	os.Exit(1)
}
