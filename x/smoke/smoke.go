package smoke

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	sdk "github.com/binance-chain/go-sdk/client"
	stypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/msg"

	"gitlab.com/thorchain/bepswap/statechain/x/smoke/types"
)

// Smoke : wallets.
type Smoke struct {
	delay     time.Duration
	MasterKey string
	PoolKey   string
	Binance   Binance
	Tests     types.Tests
}

// NewSmoke : create a new Smoke instance
func NewSmoke(masterKey, poolKey, config string) Smoke {
	cfg, err := ioutil.ReadFile(config)
	if err != nil {
		log.Fatal(err)
	}

	var tests types.Tests
	json.Unmarshal(cfg, &tests)

	return Smoke{
		delay:     2 * time.Second,
		MasterKey: masterKey,
		PoolKey:   poolKey,
		Binance:   NewBinance(true),
		Tests:     tests,
	}
}

// Setup : Generate/setup our accounts.
func (s *Smoke) Setup() {
	// Master
	mKey, _ := keys.NewPrivateKeyManager(s.MasterKey)
	mClient, _ := sdk.NewDexClient(types.TestNet, stypes.TestNetwork, mKey)

	s.Tests.Actors.Master.Key = mKey
	s.Tests.Actors.Master.Client = mClient

	// Admin
	aClient, aKey := s.ClientKey()
	s.Tests.Actors.Admin.Key = aKey
	s.Tests.Actors.Admin.Client = aClient

	// Pool
	pKey, _ := keys.NewPrivateKeyManager(s.PoolKey)
	pClient, _ := sdk.NewDexClient(types.TestNet, stypes.TestNetwork, pKey)

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
	client, _ := sdk.NewDexClient("testnet-dex.binance.org", stypes.TestNetwork, keyManager)

	return client, keyManager
}

// Summary : Private Keys
func (s *Smoke) Summary() {
	privKey, _ := s.Tests.Actors.Admin.Key.ExportAsPrivateKey()
	fmt.Printf("Admin: %v - %v\n", s.Tests.Actors.Admin.Key.GetAddr(), privKey)

	privKey, _ = s.Tests.Actors.User.Key.ExportAsPrivateKey()
	fmt.Printf("User: %v - %v\n", s.Tests.Actors.User.Key.GetAddr(), privKey)

	for idx, staker := range s.Tests.Actors.Stakers {
		privKey, _ = staker.Key.ExportAsPrivateKey()
		fmt.Printf("Staker %v: %v - %v\n", idx, staker.Key.GetAddr(), privKey)
	}
}

// Run : Where there's smoke, there's fire!
func (s *Smoke) Run() {
	s.Setup()

	for _, rule := range s.Tests.Rules {
		var payload []msg.Transfer
		var coins []stypes.Coin

		for _, coin := range rule.Coins {
			coins = append(coins, stypes.Coin{Denom: coin.Denom, Amount: int64(coin.Amount * 100000000)})
		}

		for _, to := range rule.To {
			toAddr := s.ToAddr(to)
			payload = append(payload, msg.Transfer{toAddr, coins})
		}

		client, key := s.FromClientKey(rule.From)
		s.SendTxn(client, key, payload, rule.Memo)

		if rule.Check.Target == "to" {
			for _, to := range rule.To {
				toAddr := s.ToAddr(to)
				s.CheckWallet(toAddr, rule.Check, rule.Description)
			}
		} else {
			s.CheckWallet(key.GetAddr(), rule.Check, rule.Description)
		}

		s.CheckPool(rule.Check.Pool, rule.Description)
	}
}

// FromClientKey : Client and key based on the rule "from".
func (s *Smoke) FromClientKey(from string) (sdk.DexClient, keys.KeyManager) {
	switch from {
	case "master":
		return s.Tests.Actors.Master.Client, s.Tests.Actors.Master.Key
	case "admin":
		return s.Tests.Actors.Admin.Client, s.Tests.Actors.Admin.Key
	case "user":
		return s.Tests.Actors.User.Client, s.Tests.Actors.User.Key
	default:
		stakerIdx := strings.Split(from, "_")[1]
		i, _ := strconv.Atoi(stakerIdx)
		staker := s.Tests.Actors.Stakers[i-1]
		return staker.Client, staker.Key
	}
}

// ToAddr : To address
func (s *Smoke) ToAddr(to string) stypes.AccAddress {
	switch to {
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

// Balances : Get the account balances of a given wallet.
func (s *Smoke) Balances(address stypes.AccAddress) []stypes.TokenBalance {
	acct, err := s.Tests.Actors.Master.Client.GetAccount(address.String())
	if err != nil {
		log.Fatal(err)
	}

	return acct.Balances
}

// SendTxn : Send the transaction to Binance.
func (s *Smoke) SendTxn(client sdk.DexClient, key keys.KeyManager, payload []msg.Transfer, memo string) {
	s.Binance.SendTxn(client, key, payload, memo)
}

// CheckWallet : Check the balances
func (s *Smoke) CheckWallet(address stypes.AccAddress, check types.Check, memo string) {
	time.Sleep(s.delay)
	balances := s.Balances(address)

	for _, coins := range check.Wallet {
		for _, balance := range balances {
			if coins.Denom == balance.Symbol {
				amount := coins.Amount * 100000000
				free := float64(balance.Free)

				if amount != free {
					log.Printf("%v: %v - FAIL: Amounts do not match - %f versus %f",
						memo,
						address.String(),
						amount,
						free,
					)
				} else {
					log.Printf("%v: %v - PASS", memo, address.String())
				}
			}
		}
	}
}

// GetPools : Get our pools.
func (s *Smoke) GetPools() types.StatechainPools {
	var pools types.StatechainPools

	resp, err := http.Get(types.StatechainURL)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("%v\n", err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%v\n", err)
	}

	json.Unmarshal(data, &pools)

	return pools
}

// CheckPoolUnits : Check pool units
func (s *Smoke) CheckPool(pool types.Pool, memo string) {
	pools := s.GetPools()
	for _, p := range pools {
		if p.Symbol == pool.Symbol {
			if pool.Units != 0 {
				poolUnits, _ := strconv.ParseFloat(p.PoolUnits, 64)
				if poolUnits != pool.Units {
					log.Printf("%v: FAIL: Pool Units do not match - %f versus %f",
						memo,
						pool.Units,
						poolUnits,
					)
				} else {
					log.Printf("%v: PASS", memo)
				}
			}

			if pool.Rune != 0 {

			}

			if pool.Token != 0 {

			}
		}
	}
}
