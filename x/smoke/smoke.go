package smoke

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	sdk "github.com/binance-chain/go-sdk/client"
	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/msg"

	"gitlab.com/thorchain/bepswap/statechain/x/smoke/types"
)

// Config : internal config.
type Config struct {
	delay   time.Duration
	debug   bool
	network int
}

// Smoke : Rules for our tests.
type Smoke struct {
	Config     Config
	ApiAddr    string
	Network    ctypes.ChainNetwork
	BankKey    string
	PoolKey    string
	Binance    Binance
	Statechain Statechain
	Tests      types.Tests
}

// NewSmoke : create a new Smoke instance
func NewSmoke(apiAddr, bankKey, poolKey, env string, config string, network int, debug bool) Smoke {
	cfg, err := ioutil.ReadFile(config)
	if err != nil {
		log.Fatal(err)
	}

	var tests types.Tests

	if err := json.Unmarshal(cfg, &tests); nil != err {
		log.Fatal(err)
	}

	n := NewNetwork(network)
	return Smoke{
		Config: Config{
			delay:   5 * time.Second,
			debug:   debug,
			network: network,
		},
		ApiAddr:    apiAddr,
		Network:    n.Type,
		BankKey:    bankKey,
		PoolKey:    poolKey,
		Binance:    NewBinance(apiAddr, n.ChainID, debug),
		Statechain: NewStatechain(env),
		Tests:      tests,
	}
}

// Setup : Generate/setup our accounts.
func (s *Smoke) Setup() {
	// Bank
	bKey, _ := keys.NewPrivateKeyManager(s.BankKey)
	bClient, _ := sdk.NewDexClient(s.ApiAddr, s.Network, bKey)

	s.Tests.Actors.Bank.Key = bKey
	s.Tests.Actors.Bank.Client = bClient

	// Master
	mClient, mKey := s.ClientKey()
	s.Tests.Actors.Master.Key = mKey
	s.Tests.Actors.Master.Client = mClient

	// Admin
	aClient, aKey := s.ClientKey()
	s.Tests.Actors.Admin.Key = aKey
	s.Tests.Actors.Admin.Client = aClient

	// Pool
	pKey, _ := keys.NewPrivateKeyManager(s.PoolKey)
	pClient, _ := sdk.NewDexClient(s.ApiAddr, s.Network, pKey)

	s.Tests.Actors.Pool.Key = pKey
	s.Tests.Actors.Pool.Client = pClient

	// Stakers
	for i := 1; i <= s.Tests.StakerCount; i++ {
		sClient, sKey := s.ClientKey()
		s.Tests.Actors.Stakers = append(s.Tests.Actors.Stakers, types.Keys{Key: sKey, Client: sClient})
	}

	// User
	uClient, uKey := s.ClientKey()
	s.Tests.Actors.User.Key = uKey
	s.Tests.Actors.User.Client = uClient

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
	privKey, _ := s.Tests.Actors.Master.Key.ExportAsPrivateKey()
	log.Printf("Master: %v - %v\n", s.Tests.Actors.Master.Key.GetAddr(), privKey)

	privKey, _ = s.Tests.Actors.Admin.Key.ExportAsPrivateKey()
	log.Printf("Admin: %v - %v\n", s.Tests.Actors.Admin.Key.GetAddr(), privKey)

	privKey, _ = s.Tests.Actors.User.Key.ExportAsPrivateKey()
	log.Printf("User: %v - %v\n", s.Tests.Actors.User.Key.GetAddr(), privKey)

	for idx, staker := range s.Tests.Actors.Stakers {
		privKey, _ = staker.Key.ExportAsPrivateKey()
		log.Printf("Staker %v: %v - %v\n", idx, staker.Key.GetAddr(), privKey)
	}
}

// Run : Where there's smoke, there's fire!
func (s *Smoke) Run() {
	s.Setup()

	for _, rule := range s.Tests.Rules {
		var payload []msg.Transfer
		var coins []ctypes.Coin

		for _, coin := range rule.Coins {
			coins = append(coins, ctypes.Coin{Denom: coin.Symbol, Amount: int64(coin.Amount * types.Multiplier)})
		}

		if len(coins) > 0 {
			for _, to := range rule.To {
				toAddr := s.ToAddr(to)
				payload = append(payload, msg.Transfer{toAddr, coins})
			}

			client, key := s.FromClientKey(rule.From)
			s.SendTxn(client, key, payload, rule.Memo)
		}

		// Validate.
		s.ValidateTest(rule)
	}

	s.Sweep()
}

// FromClientKey : Client and key based on the rule "from".
func (s *Smoke) FromClientKey(from string) (sdk.DexClient, keys.KeyManager) {
	switch from {
	case "bank":
		return s.Tests.Actors.Bank.Client, s.Tests.Actors.Bank.Key
	case "master":
		return s.Tests.Actors.Master.Client, s.Tests.Actors.Master.Key
	case "admin":
		return s.Tests.Actors.Admin.Client, s.Tests.Actors.Admin.Key
	case "user":
		return s.Tests.Actors.User.Client, s.Tests.Actors.User.Key
	case "pool":
		return s.Tests.Actors.Pool.Client, s.Tests.Actors.Pool.Key
	default:
		stakerIdx := strings.Split(from, "_")[1]
		i, _ := strconv.Atoi(stakerIdx)
		staker := s.Tests.Actors.Stakers[i-1]
		return staker.Client, staker.Key
	}
}

// ToAddr : To address
func (s *Smoke) ToAddr(to string) ctypes.AccAddress {
	switch to {
	case "master":
		return s.Tests.Actors.Master.Key.GetAddr()
	case "admin":
		return s.Tests.Actors.Admin.Key.GetAddr()
	case "user":
		return s.Tests.Actors.User.Key.GetAddr()
	case "pool":
		return s.Tests.Actors.Pool.Key.GetAddr()
	default:
		stakerIdx := strings.Split(to, "_")[1]
		i, _ := strconv.Atoi(stakerIdx)
		return s.Tests.Actors.Stakers[i-1].Key.GetAddr()
	}
}

// ValidateTest : Determine if the test passed or failed.
func (s *Smoke) ValidateTest(rule types.Rule) {
	if rule.Check.Target == "to" {
		for _, to := range rule.To {
			toAddr := s.ToAddr(to)
			s.CheckBinance(toAddr, rule.Check, rule.Description)
		}
	} else {
		_, key := s.FromClientKey(rule.From)
		s.CheckBinance(key.GetAddr(), rule.Check, rule.Description)
	}

	_, fromKey := s.FromClientKey(rule.From)
	s.CheckPool(fromKey.GetAddr(), rule)
}

// Balances : Get the account balances of a given wallet.
func (s *Smoke) Balances(address ctypes.AccAddress) []ctypes.TokenBalance {
	acct, err := s.Tests.Actors.Bank.Client.GetAccount(address.String())
	if err != nil {
		log.Fatal(err)
	}

	return acct.Balances
}

// SendTxn : Send the transaction to Binance.
func (s *Smoke) SendTxn(client sdk.DexClient, key keys.KeyManager, payload []msg.Transfer, memo string) {
	s.Binance.SendTxn(client, key, payload, memo)
}

// GetPools : Get our pools.
func (s *Smoke) GetPools() types.Pools {
	var pools types.Pools

	resp, err := http.Get(s.Statechain.PoolURL())
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

// CheckBinance : Check the balances
func (s *Smoke) CheckBinance(address ctypes.AccAddress, check types.Check, memo string) {
	time.Sleep(s.Config.delay)
	balances := s.Balances(address)

	for _, coins := range check.Binance {
		for _, balance := range balances {
			if coins.Symbol == balance.Symbol {
				amount := coins.Amount * types.Multiplier
				free := float64(balance.Free)

				if amount != free {
					log.Printf("%v: FAIL - Binance Balance - %v - Amounts do not match! %f versus %f - %v",
						memo,
						address.String(),
						amount,
						free,
						coins.Symbol,
					)
				} else {
					log.Printf("%v: PASS - Binance Balance - %v - %v",
						memo,
						address.String(),
						coins.Symbol,
					)
				}
			}
		}
	}
}

// CheckPool : Check Statechain pool
func (s *Smoke) CheckPool(address ctypes.AccAddress, rule types.Rule) {
	time.Sleep(s.Config.delay)

	pool := rule.Check.Statechain
	pools := s.GetPools()

	for _, p := range pools {
		if p.Symbol == pool.Symbol {
			// Check pool units
			poolUnits, _ := strconv.ParseFloat(p.PoolUnits, 64)
			if poolUnits != pool.Units {
				log.Printf("%v: FAIL - Pool Units - Units do not match! %f versus %f",
					rule.Description,
					pool.Units,
					poolUnits,
				)
			} else {
				log.Printf("%v: PASS - Pool Units - %v (%v)",
					rule.Description,
					address,
					rule.Memo,
				)
			}

			// Check Rune
			balanceRune, _ := strconv.ParseFloat(p.BalanceRune, 64)
			if balanceRune != pool.Rune {
				log.Printf("%v: FAIL - Pool Rune - Balance does not match! %f versus %f",
					rule.Description,
					pool.Rune,
					balanceRune,
				)
			} else {
				log.Printf("%v: PASS - Pool Rune - %v (%v)",
					rule.Description,
					address,
					rule.Memo,
				)
			}

			// Check token
			balanceToken, _ := strconv.ParseFloat(p.BalanceToken, 64)
			if balanceToken != pool.Token {
				log.Printf("%v: FAIL - Pool Token - Balance does not match! %f versus %f",
					rule.Description,
					pool.Token,
					balanceToken,
				)
			} else {
				log.Printf("%v: PASS - Pool Token - %v (%v)",
					rule.Description,
					address,
					rule.Memo,
				)
			}

			// Check status (used only for enabling a pool)
			if pool.Status != "" {
				if pool.Status != p.Status {
					log.Printf("%v: FAIL - Pool Status - Status does not match! %v versus %v",
						rule.Description,
						pool.Status,
						p.Status,
					)
				} else {
					log.Printf("%v: PASS - Pool Status - %v (%v)",
						rule.Description,
						address,
						rule.Memo,
					)
				}
			}
		}
	}
}

// Sweep : Transfer all assets back to master
func (s *Smoke) Sweep() {
	keys := make([]string, 5)
	keys = append(keys, s.PoolKey)

	// Master
	mKey, _ := s.Tests.Actors.Master.Key.ExportAsPrivateKey()
	keys = append(keys, mKey)

	// Admin
	aKey, _ := s.Tests.Actors.Admin.Key.ExportAsPrivateKey()
	keys = append(keys, aKey)

	// Stakers
	for _, staker := range s.Tests.Actors.Stakers {
		sKey, _ := staker.Key.ExportAsPrivateKey()
		keys = append(keys, sKey)
	}

	// User
	uKey, _ := s.Tests.Actors.User.Key.ExportAsPrivateKey()
	keys = append(keys, uKey)

	// Empty the wallets.
	sweep := NewSweep(s.ApiAddr, s.BankKey, keys, s.Config.network, s.Config.debug)
	sweep.EmptyWallets()
}
