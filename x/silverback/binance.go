package silverback

import (
	"fmt"
	"os"
	log "github.com/rs/zerolog/log"

	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/common/types"
	//ctypes "github.com/binance-chain/go-sdk/common/types"
	sdk "github.com/binance-chain/go-sdk/client"
	transaction "github.com/binance-chain/go-sdk/client/transaction"
)

type Binance struct {
	PoolAddress string
	PrivateKey string
	DexHost string
	Client sdk.DexClient
	KeyManager keys.KeyManager
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
		KeyManager: keyManager,
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
	toAddr, _ := types.AccAddressFromBech32(string(types.AccAddress(to)))
	send, err := b.Client.SendToken([]msg.Transfer{{toAddr, types.Coins{types.Coin{Denom: symbol, Amount: amount}}}}, true)

	if err != nil {
		log.Error().Msgf("Error: %v", err)
	}

	log.Info().Msgf("Send: %v", send)

	return send
}
