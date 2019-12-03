package smoke

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/msg"
	. "github.com/logrusorgru/aurora"

	"gitlab.com/thorchain/thornode/test/smoke/types"
)

// Smoke : test instructions.
type Smoke struct {
	Balances     types.BalancesConfigs
	Transactions []types.TransactionConfig
	ApiAddr      string
	PoolAddress  ctypes.AccAddress
	VaultKey     string
	FaucetKey    string
	Binance      Binance
	Thorchain    Thorchain
	Keys         map[string]keys.KeyManager
	SweepOnExit  bool
	GenBalance   bool
	FastFail     bool
	Debug        bool
	Results      types.Results
}

// NewSmoke : create a new Smoke instance.
func NewSmoke(apiAddr, faucetKey string, vaultKey, env string, bal, txns string, genBal, fastFail, debug bool) Smoke {
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

	thor := NewThorchain(env)
	// wait for thorchain to become available
	thor.WaitForAvailability()

	// Detect if THORNode should sweep for funds at the end
	sweep := false
	if len(faucetKey) > 0 {
		sweep = true
	}
	addr, err := thor.PoolAddress()
	if err != nil {
		log.Fatal(err)
	}

	return Smoke{
		Balances:     balConfig,
		Transactions: txnConfig,
		ApiAddr:      apiAddr,
		Binance:      NewBinance(apiAddr, debug),
		Thorchain:    thor,
		PoolAddress:  addr,
		FaucetKey:    faucetKey,
		VaultKey:     vaultKey,
		Keys:         keyMgr,
		GenBalance:   genBal,
		FastFail:     fastFail,
		SweepOnExit:  sweep,
		Debug:        debug,
	}
}

// Gets the key manager for a given name. If one does not exist already, create
// it.
func (s *Smoke) GetKey(name string) keys.KeyManager {
	k := s.Keys[name]
	if k != nil {
		return k
	}

	// Faucet
	if name == "faucet" && len(s.FaucetKey) > 0 {
		var err error
		s.Keys["faucet"], err = keys.NewPrivateKeyManager(s.FaucetKey)
		if err != nil {
			log.Fatalf("Failed to create faucet key manager: %s", err)
		}
		return s.Keys["faucet"]
	}

	// Pool
	if name == "vault" && len(s.VaultKey) > 0 {
		var err error
		s.Keys["vault"], err = keys.NewPrivateKeyManager(s.VaultKey)
		if err != nil {
			log.Fatalf("Failed to create pool key manager: %s", err)
		}
		return s.Keys["vault"]
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

func (s *Smoke) Summarize() {
	failed := 0
	success := 0
	for _, result := range s.Results {
		if result.Success {
			success += 1
		} else {
			failed += 1
		}
	}

	prefix := Green("Pass")
	if failed > 0 {
		prefix = Red("Fail")
	}

	log.Printf("%s %d/%d correct", prefix, success, success+failed)
}

func (s *Smoke) Seed() error {
	from := s.GetKey("faucet")
	to := s.GetKey("MASTER")
	var coins []ctypes.Coin
	for denom, amount := range s.Balances[0].Master {
		coins = append(coins, ctypes.Coin{Denom: denom, Amount: amount})
	}
	payload := []msg.Transfer{
		msg.Transfer{to.GetAddr(), coins},
	}
	return s.SendTxn(from, payload, "SEED")
}

func (s *Smoke) Transfer(txn types.TransactionConfig) error {
	from := s.GetKey(txn.From)

	var to ctypes.AccAddress
	// check if THORNode are given a pool address
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

	return s.SendTxn(from, payload, txn.Memo)
}

func (s *Smoke) GetCurrentBalances() types.BalancesConfig {
	var bal types.BalancesConfig
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

	pools := s.Thorchain.GetPools()
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

	return bal
}

// Wait for transactions to occur
func (s *Smoke) WaitForTransactions(count int64) error {
	time.Sleep(100 * time.Millisecond)
	if count == 0 {
		return nil
	}
	startHeight, err := s.Binance.GetBlockHeight()
	if err != nil {
		return err
	}
	for {
		height, err := s.Binance.GetBlockHeight()
		if err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
		if startHeight+count <= height {
			return nil
		}
	}
}

// Wait for a block on thorchain
func (s *Smoke) WaitBlocks(count int) {
	if count == 0 {
		return
	}
	// Wait for the thorchain to process a block
	thorchainHeight := s.Thorchain.GetHeight()
	for {
		newHeight := s.Thorchain.GetHeight()
		if thorchainHeight+count <= newHeight {
			return
		}
	}
}

// Run : Where there's smoke, there's fire!
func (s *Smoke) Run() bool {

	// Check that THORNode are starting with a blank set of thorchain data
	pools := s.Thorchain.GetPools()
	if len(pools) > 0 {
		log.Fatal("Thorchain isn't blank. Smoke tests assume THORNode are starting from a clean state")
	}

	if err := s.Seed(); err != nil {
		log.Fatalf("Send seed Tx failure: %s", err)
	}

	stopID := int64(0)
	if id := os.Getenv("STOP_ID"); id != "" {
		var err error
		stopID, err = strconv.ParseInt(os.Getenv("STOP_ID"), 10, 64)
		if err != nil {
			stopID = 0
		}
	}

	obtainedBalances := make(types.BalancesConfigs, 0)
	obtainedBalances = append(obtainedBalances, s.Balances.GetByTx(0))

	for _, txn := range s.Transactions {

		// check if THORNode are stopping at this tx
		if stopID > 0 && txn.Tx > stopID {
			s.Summarize()
			// exit it successfully
			return true
		}

		if err := s.Transfer(txn); err != nil {
			log.Fatalf("Send Tx failure: %s", err)
		}

		expectedBal := s.Balances.GetByTx(txn.Tx)

		// if we have no outbound tx, wait a block
		if expectedBal.Out == 0 && txn.Memo != "SEED" {
			s.WaitBlocks(1)
		} else {
			// Wait for the thorchain to process blocks and send txs
			err := s.WaitForTransactions(expectedBal.Out)
			if err != nil {
				log.Fatalf("Failed to wait for txs: %w", err)
			}
		}

		obtainedBal := s.GetCurrentBalances()
		obtainedBal.Tx = txn.Tx
		obtainedBal.Out = expectedBal.Out
		obtainedBalances = append(obtainedBalances, obtainedBal)

		// Compare expected vs obtained balances
		ok, offender, ob, ex := obtainedBal.Equals(expectedBal)
		result := types.NewResult(ok, txn, obtainedBal)
		s.Results = append(s.Results, result)

		if !result.Success {
			fmt.Printf("%s ... (Tx %d)\n", Red("Fail"), result.Transaction.Tx)
			fmt.Printf("\tTransaction: %+v\n", result.Transaction)
			fmt.Printf("\tObtained: %s %+v\n", offender, ob)
			fmt.Printf("\tExpected: %s %+v\n", offender, ex)
			if s.FastFail && !s.GenBalance {
				return false
			}
		} else {
			fmt.Printf("%s ... (Tx %d)\n", Green("Pass"), result.Transaction.Tx)
		}
	}

	if s.GenBalance {
		// Save obtained balances
		file, _ := json.MarshalIndent(obtainedBalances, "", "  ")
		_ = ioutil.WriteFile("obtained_balances.json", file, 0644)

		// Save exported obtained balances (this is for google spreadsheet importing)
		generatedBalances := make([]types.BalanceExport, len(obtainedBalances))
		for i, bal := range obtainedBalances {
			generatedBalances[i] = bal.Export()
		}
		file, _ = json.MarshalIndent(generatedBalances, "", "  ")
		_ = ioutil.WriteFile("exported_balances.json", file, 0644)
	}

	if s.SweepOnExit {
		s.Sweep()
	}

	s.Summarize()

	return s.Results.Success()
}

// Sweep : Transfer all assets back to the faucet.
func (s *Smoke) Sweep() {
	// TODO: send thorchain txs to cause ragnarok
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

	err = s.Binance.BroadcastTx(hex, params)
	if err != nil {
		return errors.Wrap(err, "failed to broadcast tx:")
	}

	return nil
}
