package silverback

import (
	"fmt"
	"os"
	log "github.com/rs/zerolog/log"

	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/msg"
	sdk "github.com/binance-chain/go-sdk/client"
	types "github.com/binance-chain/go-sdk/common/types"
	transaction "github.com/binance-chain/go-sdk/client/transaction"
)

type Binance struct {
	PoolAddress string
	PrivateKey string
	DexHost string
	Client sdk.DexClient
}

func NewBinance() *Binance {
	key := os.Getenv("PRIVATE_KEY")
	if key == "" {
		log.Fatal().Msg("No private key set!")
		os.Exit(1)
	}

	keyManager, err := keys.NewPrivateKeyManager(key)
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
		os.Exit(1)
	}

	dexHost := os.Getenv("DEX_HOST")
	bClient, err := sdk.NewDexClient(dexHost, types.TestNetwork, keyManager)
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
		os.Exit(1)
	}

	poolAddress := fmt.Sprintf("%s", keyManager.GetAddr())

	return &Binance{
		PoolAddress: poolAddress,
		PrivateKey: key,
		DexHost: dexHost,
		Client: bClient,
	}
}

func (b *Binance) GetAccount() *types.BalanceAccount {
	account, err := b.Client.GetAccount(b.PoolAddress)
	if err != nil {
		log.Error().Msgf("Error: %v", err)
	}

	return account
}

func (b *Binance) SendToken(to string, symbol string, amount int64) *transaction.SendTokenResult {
	log.Info().Msgf("to: %v, symbol: %v, amount: %v", to, symbol, amount)

	coin := types.Coin{Denom: symbol, Amount: amount}
	var coins []types.Coin
	c := append(coins, coin)

	address := types.AccAddress(to)

	transfer := msg.Transfer{ToAddr: address, Coins: c}
	var transfers []msg.Transfer
	t := append(transfers, transfer)
	log.Info().Msgf("Transaction to send: %v", t)

	send, err := b.Client.SendToken(t, true)
	if err != nil {
		log.Error().Msgf("Error: %v", err)
	}

	return send
}
