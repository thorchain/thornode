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
	Balances     types.BalancesConfigs
	Transactions []types.TransactionConfig
	ApiAddr      string
	Network      ctypes.ChainNetwork
	PoolAddress  ctypes.AccAddress
	PoolKey      string
	Binance      Binance
	Statechain   Statechain
	Keys         map[string]keys.KeyManager
	SweepOnExit  bool
	FastFail     bool
	Debug        bool
	Results      types.Results
}

// NewSmoke : create a new Smoke instance.
func NewSmoke(apiAddr, faucetKey string, poolKey, env string, bal, txns string, network ctypes.ChainNetwork, logFile string, sweep, fastFail, debug bool) Smoke {
	balRaw, err := ioutil.ReadFile(bal)
	if err != nil {
		log.Fatal(err)
	}

	var balConfig types.BalancesConfigs
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
		keyMgr["vault"], err = keys.NewPrivateKeyManager(poolKey)
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
		FastFail:     fastFail,
		SweepOnExit:  sweep,
		Debug:        debug,
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
	fmt.Printf("Name: %s %s\n", name, k.GetAddr())

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
func (s *Smoke) Run() bool {

	// Check that we are starting with a blank set of statechain data
	pools := s.Statechain.GetPools()
	if len(pools) > 0 {
		log.Fatal("Statechain isn't blank. Smoke tests assume we are starting from a clean state")
	}

	////////// Run the faucet ////////
	from := s.GetKey("faucet")
	to := s.GetKey("MASTER")
	var coins []ctypes.Coin
	for denom, amount := range s.Balances[0].Master {
		coins = append(coins, ctypes.Coin{Denom: denom, Amount: amount})
	}
	payload := []msg.Transfer{
		msg.Transfer{to.GetAddr(), coins},
	}

	err := s.SendTxn(from, payload, "SEED")
	if err != nil {
		log.Fatalf("Send Tx failure: %s", err)
	}
	/////////////////////////////////

	for _, txn := range s.Transactions {

		from := s.GetKey(txn.From)

		var to ctypes.AccAddress
		// check if we are given a pool address
		if strings.EqualFold(txn.To, "vault") && len(s.PoolAddress) > 0 {
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

		if txn.Memo != "SEED" {
			// Wait for the statechain to process a block
			statechainHeight := s.Statechain.GetHeight()
			for {
				newHeight := s.Statechain.GetHeight()
				if statechainHeight < newHeight {
					break
				}
			}
		}

		targetBal := s.Balances.GetByTx(txn.Tx)
		var bal types.BalancesConfig
		bal.Tx = targetBal.Tx
		for name, key := range s.Keys {
			acc, err := s.Binance.GetAccount(key.GetAddr())
			if err != nil {
				log.Fatalf("Error checking balance: %s", err)
			}
			balances := make(map[string]int64, 0)
			for _, coin := range acc.Coins {
				balances[coin.Denom] = coin.Amount
			}

			switch strings.ToLower(name) {
			case "master":
				bal.Master = balances
			case "user-1":
				bal.User1 = balances
			case "staker-1":
				bal.Staker1 = balances
			case "staker-2":
				bal.Staker2 = balances
			}
		}

		// get vault balance
		acc, err := s.Binance.GetAccount(s.PoolAddress)
		if err != nil {
			log.Fatalf("Error checking balance: %s", err)
		}

		balances := make(map[string]int64, 0)
		for _, coin := range acc.Coins {
			balances[coin.Denom] = coin.Amount
		}
		bal.Vault = balances

		pools := s.Statechain.GetPools()
		for _, pool := range pools {
			balances := make(map[string]int64, 0)
			balances["RUNE-A1F"] = pool.BalanceRune
			balances[pool.Asset.Symbol] = pool.BalanceAsset
			switch pool.Asset.Symbol {
			case "BNB":
				bal.PoolBNB = balances
			case "LOK-3C0":
				bal.PoolLoki = balances
			}
		}

		ok, label, ob, ex := bal.Equals(targetBal)
		result := types.NewResult(ok, txn, bal)
		s.Results = append(s.Results, result)

		if !result.Success {
			fmt.Printf("Fail (Tx %d)\n", result.Transaction.Tx)
			fmt.Printf("Transaction: %+v\n", result.Transaction)
			fmt.Printf("Obtained: %s %+v\n", label, ob)
			fmt.Printf("Expected: %s %+v\n", label, ex)
			if s.FastFail {
				return false
			}
		} else {
			if s.Debug {
				fmt.Printf("Test Success! (%d)\n", result.Transaction.Tx)
			}
		}
	}

	if s.SweepOnExit {
		s.Sweep()
	}

	s.Summary()

	return s.Results.Success()
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
