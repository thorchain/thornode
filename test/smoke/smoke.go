package smoke

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/bepswap/thornode/test/smoke/types"
)

// Smoke : test instructions.
type Smoke struct {
	Balances     []types.BalancesConfig
	Transactions []types.TransactionConfig
	ApiAddr      string
	Network      ctypes.ChainNetwork
	PoolAddress  ctypes.AccAddress
	PoolKey      string
	Binance      Binance
	Statechain   Statechain
	Keys         map[string]keys.KeyManager
	SweepOnExit  bool
	Results      types.Results
}

// NewSmoke : create a new Smoke instance.
func NewSmoke(apiAddr, faucetKey string, poolKey, env string, bal, txns string, network ctypes.ChainNetwork, logFile string, sweep, debug bool) Smoke {
	balRaw, err := ioutil.ReadFile(bal)
	if err != nil {
		log.Fatal(err)
	}

	var balConfig []types.BalancesConfig
	if err := json.Unmarshal(balRaw, &balConfig); nil != err {
		log.Fatal(err)
	}

	txnRaw, err := ioutil.ReadFile(txns)
	if err != nil {
		log.Fatal(err)
	}

	var txnConfig []types.TransactionConfig
	if err := json.Unmarshal(txnRaw, &txnConfig); nil != err {
		log.Fatal(err)
	}

	keyMgr := make(map[string]keys.KeyManager, 0)

	// Faucet
	if len(faucetKey) > 0 {
		var err error
		keyMgr["faucet"], err = keys.NewPrivateKeyManager(faucetKey)
		if err != nil {
			log.Fatalf("Failed to create faucet key manager: %s", err)
		}
	}

	// Pool
	if len(poolKey) > 0 {
		var err error
		keyMgr["faucet"], err = keys.NewPrivateKeyManager(poolKey)
		if err != nil {
			log.Fatalf("Failed to create pool key manager: %s", err)
		}
	}

	// TODO: pull network from binance node
	smoke := Smoke{
		Balances:     balConfig,
		Transactions: txnConfig,
		ApiAddr:      apiAddr,
		Network:      network,
		Binance:      NewBinance(apiAddr, network, debug),
		Statechain:   NewStatechain(env),
		Keys:         keyMgr,
		SweepOnExit:  sweep,
	}

	// detect pool address
	smoke.PoolAddress = smoke.Statechain.PoolAddress()

	return smoke
}

func (s *Smoke) GetKey(name string) keys.KeyManager {
	k := s.Keys[name]
	if k != nil {
		return k
	}

	// build key, and save
	var err error
	k, err = keys.NewKeyManager()
	if err != nil {
		log.Fatalf("Error creating key manager: %s", err)
	}
	s.Keys[name] = k

	return k
}

func (s *Smoke) Summary() {
	failed := 0
	success := 0
	for _, result := range s.Results {
		if result.Success {
			success += 1
		} else {
			failed += 1
		}
	}
	log.Printf("%d/%d correct", success, success+failed)
	/*
		for name, actor := range s.Tests.ActorKeys {
			privKey, _ := actor.ExportAsPrivateKey()
			log.Printf("%v: %v - %v\n", name, actor.GetAddr(), privKey)
		}
	*/
}

// Run : Where there's smoke, there's fire!
func (s *Smoke) Run() {

	for i, txn := range s.Transactions {

		from := s.GetKey(txn.From)

		var to ctypes.AccAddress
		// check if we are given a pool address
		if txn.To == "pool" && len(s.PoolAddress) > 0 {
			to = s.PoolAddress
		} else {
			to = s.GetKey(txn.To).GetAddr()
		}

		var coins []ctypes.Coin

		for denom, amount := range txn.Coins {
			if amount > 0 {
				coins = append(coins, ctypes.Coin{Denom: denom, Amount: amount})
			}
		}

		payload := []msg.Transfer{
			msg.Transfer{to, coins},
		}

		err := s.SendTxn(from, payload, txn.Memo)
		if err != nil {
			log.Fatalf("Send Tx failure: %s", err)
		}

		if txn.Memo == "SEED" {
			// this is a seed transaction, no validation needed
			continue
		}

		// TODO: Validate.
		var bal types.BalancesConfig
		for name, key := range s.Keys {
			acc, err := s.Binance.GetAccount(key.GetAddr())
			if err != nil {
				log.Fatalf("Error checking balance: %s", err)
			}
			var balances map[string]int64
			for _, coin := range acc.Coins {
				balances[coin.Denom] = coin.Amount
			}

			fmt.Printf("Name: %s\n", name)
			fmt.Printf("Balances: %+v\n", balances)
			fmt.Printf("Coins: %+v\n", acc.Coins)
			switch strings.ToLower(name) {
			case "master":
				bal.Master = balances
			case "user-1":
				bal.User1 = balances
			case "staker-1":
				bal.Staker1 = balances
			case "staker-2":
				bal.Staker2 = balances
			case "vault":
				bal.Vault = balances
			}
		}

		pools := s.Statechain.GetPools()
		for _, pool := range pools {
			var balances map[string]int64
			balances["RUNE-A1F"] = pool.BalanceRune
			balances[pool.Asset.Symbol] = pool.BalanceRune
			fmt.Printf("Pool Name: %s\n", pool.Asset.Ticker)
			switch strings.ToLower(pool.Asset.Ticker) {
			case "bnb":
				bal.PoolBNB = balances
			case "loki":
				bal.PoolLoki = balances
			}
		}

		result := types.NewResult(bal.Equals(s.Balances[i]), txn, bal)
		s.Results = append(s.Results, result)

		if !result.Success {
			fmt.Printf("Test failed (%d): %+v\n", result.Transaction.Tx, result.Transaction)
			fmt.Printf("Obtained: %+v\n", result.Obtained)
			fmt.Printf("Expected: %+v\n", s.Balances[i])
		} else {
			fmt.Printf("Test Success! (%d)", result.Transaction.Tx)
		}
	}

	if s.SweepOnExit {
		s.Sweep()
	}

	s.Summary()
}

// Sweep : Transfer all assets back to the faucet.
func (s *Smoke) Sweep() {
	// TODO: send statechain txs to cause ragnarok
	/*
		keys := make([]string, len(s.Tests.ActorList)+1)
		key, _ := s.Tests.ActorKeys["pool"].ExportAsPrivateKey()
		keys = append(keys, key)

		for _, actor := range s.Tests.ActorList {
			key, _ = s.Tests.ActorKeys[actor].ExportAsPrivateKey()
			if key != s.FaucetKey {
				keys = append(keys, key)
			}
		}

		// Empty the wallets.
		sweep := NewSweep(s.ApiAddr, s.FaucetKey, keys, s.Config.network, s.Config.debug)
		sweep.EmptyWallets()
	*/
}

// SendTxn : Send the transaction to Binance.
func (s *Smoke) SendTxn(key keys.KeyManager, payload []msg.Transfer, memo string) error {
	sendMsg, err := s.Binance.ParseTx(key, payload)
	if err != nil {
		return errors.Wrap(err, "failed to parse tx:")
	}

	hex, params, err := s.Binance.SignTx(key, sendMsg, memo)
	if err != nil {
		return errors.Wrap(err, "Failed to sign tx:")
	}

	_, err = s.Binance.BroadcastTx(hex, params)
	if err != nil {
		return errors.Wrap(err, "failed to broadcast tx:")
	}

	return nil
}
