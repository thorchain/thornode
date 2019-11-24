package smoke

import (
	"log"

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
	KeyList    []string
}

// NewHoover : Create a new instance of Sweep.
func NewSweep(apiAddr, masterPrivKey string, keyList []string, network btypes.ChainNetwork, debug bool) Sweep {
	keyManager, err := keys.NewPrivateKeyManager(masterPrivKey)
	if err != nil {
		log.Fatalf("Error creating key manager: %s", err)
	}

	return Sweep{
		ApiAddr:    apiAddr,
		Network:    network,
		Binance:    NewBinance(apiAddr, network, debug),
		KeyManager: keyManager,
		KeyList:    keyList,
	}
}

// EmptyWallets : Empty and transfer all assets out of the wallet.
func (s Sweep) EmptyWallets() {
	for _, key := range s.KeyList {
		keyManager, _ := keys.NewPrivateKeyManager(key)

		var coins []btypes.Coin
		balances := s.Balances(keyManager.GetAddr())
		for _, asset := range balances {
			amount := asset.Amount

			// Binance fees.
			if len(balances) > 1 && asset.Denom == "BNB" {
				amount = amount - (int64(len(balances)) * 30000)
			} else if asset.Denom == "BNB" {
				amount = amount - 37500
			}

			if amount > 0 {
				coins = append(coins, btypes.Coin{Denom: asset.Denom, Amount: amount})
			}
		}

		if len(coins) > 0 {
			payload := []msg.Transfer{msg.Transfer{s.KeyManager.GetAddr(), coins}}
			_ = s.SendTxn(keyManager, payload, "SWEEP:RETURN")
		}
	}
}

// Balances : Get the account balances of a given wallet.
func (s Sweep) Balances(address btypes.AccAddress) btypes.Coins {
	acct, err := s.Binance.GetAccount(address)
	if err != nil {
		log.Fatal(err)
	}
	return acct.Coins
}

// SendTxn : Send the transaction to Binance.
func (s *Sweep) SendTxn(key keys.KeyManager, payload []msg.Transfer, memo string) error {
	sendMsg, err := s.Binance.ParseTx(key, payload)
	if err != nil {
		return err
	}

	hex, params, err := s.Binance.SignTx(key, sendMsg, memo)
	if err != nil {
		return err
	}

	err = s.Binance.BroadcastTx(hex, params)
	return err
}
