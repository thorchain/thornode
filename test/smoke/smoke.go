package smoke

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	sdk "github.com/binance-chain/go-sdk/client"
	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/msg"

	"gitlab.com/thorchain/bepswap/thornode/test/smoke/types"
)

// Config : test config
type Config struct {
	delay         time.Duration
	debug         bool
	network       int
	resultsFile   string
	thorchainFile string
}

// Smoke : test instructions.
type Smoke struct {
	Config           Config
	ApiAddr          string
	Network          ctypes.ChainNetwork
	FaucetKey        string
	PoolKey          string
	Binance          Binance
	Thorchain        Thorchain
	Tests            types.Tests
	TestResults      []types.TestResults
	ThorchainResults []types.ThorchainResults
}

// NewSmoke : create a new Smoke instance.
func NewSmoke(apiAddr, faucetKey, poolKey, env string, config string, network int, resultsFile, thorchainFile string, debug bool) Smoke {
	cfg, err := ioutil.ReadFile(config)
	if err != nil {
		log.Fatal(err)
	}

	var tests types.Tests
	if err := json.Unmarshal(cfg, &tests); nil != err {
		log.Fatal(err)
	}

	var testResults []types.TestResults
	var thorchainResults []types.ThorchainResults
	n := NewNetwork(network)
	return Smoke{
		Config: Config{
			delay:         5 * time.Second,
			debug:         debug,
			network:       network,
			resultsFile:   resultsFile,
			thorchainFile: thorchainFile,
		},
		ApiAddr:          apiAddr,
		Network:          n.Type,
		FaucetKey:        faucetKey,
		PoolKey:          poolKey,
		Binance:          NewBinance(apiAddr, n.ChainID, debug),
		Thorchain:        NewThorchain(env),
		Tests:            tests,
		TestResults:      testResults,
		ThorchainResults: thorchainResults,
	}
}

// Setup : Generate/setup our accounts.
func (s *Smoke) Setup() {
	s.Tests.ActorKeys = make(map[string]types.Keys)

	// Faucet
	key, _ := keys.NewPrivateKeyManager(s.FaucetKey)
	client, _ := sdk.NewDexClient(s.ApiAddr, s.Network, key)
	s.Tests.ActorKeys["faucet"] = types.Keys{Key: key, Client: client}

	for _, actor := range s.Tests.ActorList {
		client, key = s.ClientKey()
		s.Tests.ActorKeys[actor] = types.Keys{Key: key, Client: client}
	}

	// Pool
	key, _ = keys.NewPrivateKeyManager(s.PoolKey)
	client, _ = sdk.NewDexClient(s.ApiAddr, s.Network, key)
	s.Tests.ActorKeys["pool"] = types.Keys{Key: key, Client: client}

	s.Summary()
}

// ClientKey : instantiate Client and Keys Binance SDK objects.
func (s *Smoke) ClientKey() (sdk.DexClient, keys.KeyManager) {
	keyManager, _ := keys.NewKeyManager()
	client, _ := sdk.NewDexClient(s.ApiAddr, s.Network, keyManager)

	return client, keyManager
}

// Summary : Private Keys
func (s *Smoke) Summary() {
	for name, actor := range s.Tests.ActorKeys {
		privKey, _ := actor.Key.ExportAsPrivateKey()
		log.Printf("%v: %v - %v\n", name, actor.Key.GetAddr(), privKey)
	}
}

// Run : Where there's smoke, there's fire!
func (s *Smoke) Run() {
	s.Setup()

	for tx, rule := range s.Tests.Rules {
		var payload []msg.Transfer

		for _, to := range rule.To {
			var coins []ctypes.Coin

			for _, coin := range to.Coins {
				coins = append(coins, ctypes.Coin{Denom: coin.Symbol, Amount: coin.Amount})
			}

			toAddr := s.Tests.ActorKeys[to.Actor].Key.GetAddr()
			payload = append(payload, msg.Transfer{toAddr, coins})
		}

		memo := rule.Memo
		if rule.SendTo != "" {
			sendTo := s.Tests.ActorKeys[rule.SendTo].Key.GetAddr()
			memo = memo + ":" + sendTo.String()
		}

		if rule.SlipLimit != 0 {
			memo = fmt.Sprintf("%s:%v", memo, rule.SlipLimit)
		}

		from := s.Tests.ActorKeys[rule.From]
		s.SendTxn(from.Client, from.Key, payload, memo)

		// Validate.
		delay := time.Second * rule.CheckDelay
		s.LogTestResults(tx, delay)
	}

	if s.Tests.SweepOnExit {
		s.Sweep()
	}

	// Save the test results.
	s.SaveResults()
}

// SaveLog : Save the log file.
func (s *Smoke) SaveResults() {
	testOutput, _ := json.Marshal(s.TestResults)
	_ = ioutil.WriteFile(s.Config.resultsFile, testOutput, 0644)

	thorchainOutput, _ := json.Marshal(s.ThorchainResults)
	_ = ioutil.WriteFile(s.Config.thorchainFile, thorchainOutput, 0644)
}

// LogResults : Log our results.
func (s *Smoke) LogTestResults(tx int, delay time.Duration) {
	time.Sleep(delay)

	s.BinanceState(tx)
	s.ThorchainState(tx)
}

// BinanceState : Compare expected vs actual Binance wallet values.
func (s *Smoke) BinanceState(tx int) {
	client := s.Tests.ActorKeys["faucet"].Client
	var output types.TestResults
	output.Tx = tx + 1

	s.TestResults = append(s.TestResults, output)

	for _, actor := range s.Tests.ActorList {
		balances := s.GetBinance(client, s.Tests.ActorKeys[actor].Key.GetAddr())
		for _, balance := range balances {
			amount := balance.Free.ToInt64()

			switch balance.Symbol {
			case "RUNE-A1F":
				s.ActorAmount(amount, &s.TestResults[tx].Rune, actor)
			case "BNB":
				s.ActorAmount(amount, &s.TestResults[tx].Bnb, actor)
			case "LOK-3C0":
				s.ActorAmount(amount, &s.TestResults[tx].Lok, actor)
			}
		}
	}
}

// GetBinance : Get Binance account balance.
func (s *Smoke) GetBinance(client sdk.DexClient, address ctypes.AccAddress) []ctypes.TokenBalance {
	acct, err := client.GetAccount(address.String())
	if err != nil {
		log.Fatal(err)
	}

	return acct.Balances
}

// ActorAmount : Amount for a given actor
func (s *Smoke) ActorAmount(amount int64, output *types.Balance, actor string) {
	switch actor {
	case "master":
		output.Master = amount
	case "admin":
		output.Admin = amount
	case "user":
		output.User = amount
	case "staker_1":
		output.Staker1 = amount
	case "staker_2":
		output.Staker2 = amount
	}
}

// ThorchainState : Current Thorchain state.
func (s *Smoke) ThorchainState(tx int) {
	thorchain := s.GetThorchain()

	var amount int64
	for _, pools := range thorchain {
		amount += pools.BalanceRune

		switch pools.Asset.Symbol {
		case "LOK-3C0":
			s.TestResults[tx].Lok.Pool = pools.BalanceAsset
		case "BNB":
			s.TestResults[tx].Bnb.Pool = pools.BalanceAsset
		}
	}

	// Record to our test summary.
	s.TestResults[tx].Rune.Pool = amount

	// Save for auditing purposes.
	idx := tx + 1
	s.ThorchainResults = append(s.ThorchainResults,
		types.ThorchainResults{idx, thorchain},
	)
}

// GetThorchain : Get the Thorchain pools.
func (s *Smoke) GetThorchain() types.ThorchainPools {
	// TODO : Fix this - this is a hack to get around the 1 query per second REST API limit.
	time.Sleep(1 * time.Second)

	var pools types.ThorchainPools

	resp, err := http.Get(s.Thorchain.PoolURL())
	if err != nil {
		log.Printf("%v\n", err)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%v\n", err)
	}

	if err := json.Unmarshal(data, &pools); nil != err {
		log.Fatal(err)
	}

	return pools
}

// Sweep : Transfer all assets back to the faucet.
func (s *Smoke) Sweep() {
	keys := make([]string, len(s.Tests.ActorList)+1)
	key, _ := s.Tests.ActorKeys["pool"].Key.ExportAsPrivateKey()
	keys = append(keys, key)

	for _, actor := range s.Tests.ActorList {
		key, _ = s.Tests.ActorKeys[actor].Key.ExportAsPrivateKey()
		if key != s.FaucetKey {
			keys = append(keys, key)
		}
	}

	// Empty the wallets.
	sweep := NewSweep(s.ApiAddr, s.FaucetKey, keys, s.Config.network, s.Config.debug)
	sweep.EmptyWallets()
}

// SendTxn : Send the transaction to Binance.
func (s *Smoke) SendTxn(client sdk.DexClient, key keys.KeyManager, payload []msg.Transfer, memo string) {
	s.Binance.SendTxn(client, key, payload, memo)
}
