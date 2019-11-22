package smoke

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	sdk "github.com/binance-chain/go-sdk/client"
	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	ttypes "github.com/binance-chain/go-sdk/types"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/bepswap/thornode/test/smoke/types"
)

// Config : test config
type Config struct {
	delay   time.Duration
	debug   bool
	network ctypes.ChainNetwork
	logFile string
}

// Smoke : test instructions.
type Smoke struct {
	Config      Config
	ApiAddr     string
	Network     ctypes.ChainNetwork
	FaucetKey   string
	PoolAddress ctypes.AccAddress
	PoolKey     string
	Binance     Binance
	Statechain  Statechain
	Tests       types.Tests
	Results     []types.Output
	SweepOnExit bool
}

// NewSmoke : create a new Smoke instance.
func NewSmoke(apiAddr, faucetKey string, poolKey, env string, config string, network ctypes.ChainNetwork, logFile string, sweep, debug bool) Smoke {
	cfg, err := ioutil.ReadFile(config)
	if err != nil {
		log.Fatal(err)
	}

	var tests types.Tests
	if err := json.Unmarshal(cfg, &tests); nil != err {
		log.Fatal(err)
	}

	var results []types.Output
	smoke := Smoke{
		Config: Config{
			delay:   5 * time.Second,
			debug:   debug,
			network: network,
			logFile: logFile,
		},
		ApiAddr:     apiAddr,
		Network:     network,
		FaucetKey:   faucetKey,
		Binance:     NewBinance(apiAddr, network, debug),
		Statechain:  NewStatechain(env),
		Tests:       tests,
		Results:     results,
		SweepOnExit: sweep,
	}

	// detect pool address
	smoke.PoolAddress = smoke.StatechainPoolAddress()

	return smoke
}

// Setup : Generate/setup our accounts.
func (s *Smoke) Setup() {
	rand.Seed(time.Now().UnixNano())

	s.Tests.ActorKeys = make(map[string]keys.KeyManager)

	// Faucet
	key, err := keys.NewPrivateKeyManager(s.FaucetKey)
	if err != nil {
		log.Fatalf("Failed to create key manager: %s", err)
	}
	s.Tests.ActorKeys["faucet"] = key

	for _, actor := range s.Tests.ActorList {
		_, key := s.ClientKey()
		s.Tests.ActorKeys[actor] = key
	}

	// Pool
	if len(s.PoolKey) > 0 {
		key, err = keys.NewPrivateKeyManager(s.PoolKey)
		if err != nil {
			log.Fatalf("Failed to create key manager for pool: %s", err)
		}
		s.Tests.ActorKeys["pool"] = key
	}

	s.Summary()
}

// Get Client, retry if we fail to get it (ie API Rate limited)
func (s *Smoke) GetClient(k keys.KeyManager) sdk.DexClient {
	return GetClient(s.ApiAddr, s.Network, k)
}

// ClientKey : instantiate Client and Keys Binance SDK objects.
func (s *Smoke) ClientKey() (sdk.DexClient, keys.KeyManager) {
	keyManager, err := keys.NewKeyManager()
	if err != nil {
		log.Fatalf("Error creating key manager: %s", err)
	}
	return s.GetClient(keyManager), keyManager
}

// Summary : Private Keys
func (s *Smoke) Summary() {
	for name, actor := range s.Tests.ActorKeys {
		privKey, _ := actor.ExportAsPrivateKey()
		log.Printf("%v: %v - %v\n", name, actor.GetAddr(), privKey)
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

			// since we don't have a key for pool, inject it
			var toAddr ctypes.AccAddress
			if to.Actor == "pool" && len(s.PoolAddress) > 0 {
				toAddr = s.PoolAddress
			} else {
				toAddr = s.Tests.ActorKeys[to.Actor].GetAddr()
			}
			payload = append(payload, msg.Transfer{toAddr, coins})
		}

		memo := rule.Memo
		if rule.SendTo != "" {
			sendTo := s.Tests.ActorKeys[rule.SendTo].GetAddr()
			memo = memo + ":" + sendTo.String()
		}

		if rule.SlipLimit != 0 {
			memo = fmt.Sprintf("%s:%v", memo, rule.SlipLimit)
		}

		from := s.Tests.ActorKeys[rule.From]
		err := s.SendTxn(from, payload, memo)
		if err != nil {
			log.Fatalf("Send Tx failure: %s", err)
		}

		// Validate.
		delay := time.Second * rule.CheckDelay
		err = s.LogResults(tx, delay)
		if err != nil {
			log.Fatalf("Log Results failed: %s", err)
		}
	}

	if s.SweepOnExit {
		s.Sweep()
	}

	// Save the log.
	s.SaveLog()
}

// SaveLog : Save the log file.
func (s *Smoke) SaveLog() {
	output, _ := json.Marshal(s.Results)
	_ = ioutil.WriteFile(s.Config.logFile, output, 0644)
}

// LogResults : Log our results.
func (s *Smoke) LogResults(tx int, delay time.Duration) error {
	time.Sleep(delay)

	err := s.BinanceState(tx)
	if err != nil {
		return errors.Wrap(err, "failed to get binance state:")
	}
	s.StatechainState(tx)

	return nil
}

// BinanceState : Compare expected vs actual Binance wallet values.
func (s *Smoke) BinanceState(tx int) error {
	var output types.Output
	output.Tx = tx + 1

	s.Results = append(s.Results, output)

	for _, actor := range s.Tests.ActorList {
		balances, err := s.GetBalances(s.Tests.ActorKeys[actor].GetAddr())
		if err != nil {
			return errors.Wrap(err, "failed to get balances")
		}
		for _, coin := range balances {
			switch coin.Denom {
			case "RUNE-A1F":
				s.ActorAmount(coin.Amount, &s.Results[tx].Rune, actor)
			case "BNB":
				s.ActorAmount(coin.Amount, &s.Results[tx].Bnb, actor)
			case "LOK-3C0":
				s.ActorAmount(coin.Amount, &s.Results[tx].Lok, actor)
			}
		}
	}

	return nil
}

// GetBinance : Get Binance account balance.
func (s *Smoke) GetBalances(address ctypes.AccAddress) (ctypes.Coins, error) {
	key := append([]byte("account:"), address.Bytes()...)
	args := fmt.Sprintf("path=\"/store/acc/key\"&data=0x%x", key)
	uri := url.URL{
		Scheme:   "http", // TODO: don't hard code this
		Host:     s.ApiAddr,
		Path:     "abci_query",
		RawQuery: args,
	}
	resp, err := http.Get(uri.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type queryResult struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      string `json:"id"`
		Result  struct {
			Response struct {
				Key         string `json:"key"`
				Value       string `json:"value"`
				BlockHeight string `json:"height"`
			} `json:"response"`
		} `json:"result"`
	}

	var result queryResult
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(result.Result.Response.Value)
	if err != nil {
		return nil, err
	}

	cdc := ttypes.NewCodec()
	var acc ctypes.AppAccount
	err = cdc.UnmarshalBinaryBare(data, &acc)

	return acc.BaseAccount.Coins, err
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

// StatechainPoolAddresses : Get Current pool address
func (s *Smoke) StatechainPoolAddress() ctypes.AccAddress {
	// TODO : Fix this - this is a hack to get around the 1 query per second REST API limit.
	time.Sleep(1 * time.Second)

	var addrs types.StatechainPoolAddress

	resp, err := http.Get(s.Statechain.PoolAddressesURL())
	if err != nil {
		log.Fatalf("Failed getting statechain: %v\n", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed reading body: %v\n", err)
	}

	if err := json.Unmarshal(data, &addrs); nil != err {
		log.Fatalf("Failed to unmarshal pool addresses: %s", err)
	}

	if len(addrs.Current) == 0 {
		log.Fatal("No pool addresses are currently available")
	}
	poolAddr := addrs.Current[0]

	addr, err := ctypes.AccAddressFromBech32(poolAddr.Address.String())
	if err != nil {
		log.Fatalf("Failed to parse address: %s", err)
	}

	return addr
}

// StatechainState : Current Statechain state.
func (s *Smoke) StatechainState(tx int) {
	statechain := s.GetStatechain()

	var amount int64
	for _, pools := range statechain {
		amount += pools.BalanceRune

		switch pools.Asset.Symbol {
		case "LOK-3C0":
			s.Results[tx].Lok.Pool = pools.BalanceAsset
		case "BNB":
			s.Results[tx].Bnb.Pool = pools.BalanceAsset
		}
	}

	s.Results[tx].Rune.Pool = amount
}

// GetStatechain : Get the Statehcain pools.
func (s *Smoke) GetStatechain() types.StatechainPools {
	// TODO : Fix this - this is a hack to get around the 1 query per second REST API limit.
	time.Sleep(1 * time.Second)

	var pools types.StatechainPools

	resp, err := http.Get(s.Statechain.PoolURL())
	if err != nil {
		log.Fatalf("Failed getting statechain: %v\n", err)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed reading body: %v\n", err)
	}

	if err := json.Unmarshal(data, &pools); nil != err {
		log.Fatalf("Failed to unmarshal pools: %s", err)
	}

	return pools
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

// Get Client, retry if we fail to get it (ie API Rate limited)
func GetClient(addr string, network ctypes.ChainNetwork, k keys.KeyManager) sdk.DexClient {
	// we can get rate limited, so have a retry system.
	attempts := 25 // number of attempts
	sleep := 5 * time.Second
	var err error
	var client sdk.DexClient
	if attempts--; attempts > 0 {
		client, err = sdk.NewDexClient(addr, network, k)
		if err != nil {
			// Add some randomness to prevent creating a Thundering Herd
			jitter := time.Duration(rand.Int63n(int64(sleep)))
			sleep = sleep + jitter/2

			time.Sleep(sleep)
		}
		return client
	}
	if err != nil {
		log.Fatalf("Failed to create client: %s", err)
	}
	return client
}
