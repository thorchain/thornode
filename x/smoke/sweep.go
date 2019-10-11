package smoke

import (
	"log"

	sdk "github.com/binance-chain/go-sdk/client"
	btypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/msg"
)

// Sweep : our main sweep type.
type Sweep struct {
	ApiAddr    string
	Network    btypes.ChainNetwork
	Binance    Binance
	KeyManager keys.KeyManager
	Client     sdk.DexClient
	KeyList    []string
}

// NewHoover : Create a new instance of Sweep.
func NewSweep(apiAddr, masterPrivKey string, keyList []string, network int, debug bool) Sweep {
	n := NewNetwork(network)

	keyManager, _ := keys.NewPrivateKeyManager(masterPrivKey)
	client, _ := sdk.NewDexClient(apiAddr, n.Type, keyManager)

	return Sweep{
		ApiAddr:    apiAddr,
		Network:    n.Type,
		Binance:    NewBinance(apiAddr, n.ChainID,debug),
		KeyManager: keyManager,
		Client:     client,
		KeyList:    keyList,
	}
}

// EmptyWallets : Empty and transfer all assets out of the wallet.
func (s Sweep) EmptyWallets() {
	for _, key := range s.KeyList {
		keyManager, _ := keys.NewPrivateKeyManager(key)
		client, _ := sdk.NewDexClient(s.ApiAddr, s.Network, keyManager)

		var coins []btypes.Coin
		balances := s.Balances(keyManager.GetAddr())
		for _, token := range balances {
			free := float64(token.Free)
			amt := int64(free)
			if token.Symbol == "BNB" {
				amt = amt - 210000
			}

			if amt > 0 {
				coins = append(coins, btypes.Coin{Denom: token.Symbol, Amount: amt})
			}
		}

		if len(coins) > 0 {
			payload := []msg.Transfer{msg.Transfer{s.KeyManager.GetAddr(), coins}}
			s.SendTxn(client, keyManager, payload, "SWEEP:RETURN")
		}
	}
}

// Balances : Get the account balances of a given wallet.
func (s Sweep) Balances(address btypes.AccAddress) []btypes.TokenBalance {
	acct, err := s.Client.GetAccount(address.String())
	if err != nil {
		log.Fatal(err)
	}

	return acct.Balances
}

// SendTxn : Send our transaction to Binance
func (s Sweep) SendTxn(client sdk.DexClient, key keys.KeyManager, payload []msg.Transfer, memo string) {
	s.Binance.SendTxn(client, key, payload, memo)
}
