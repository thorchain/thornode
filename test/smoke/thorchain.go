package smoke

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"gitlab.com/thorchain/bepswap/thornode/test/smoke/types"
)

var endpoints = map[string]string{
	"local":      "localhost:1317",
	"staging":    "testnet-chain.bepswap.io",
	"develop":    "testnet-chain.bepswap.net",
	"production": "testnet-chain.bepswap.com",
}

type Thorchain struct {
	Env string
}

// NewThorchain : Create a new Thorchain instance.
func NewThorchain(env string) Thorchain {
	return Thorchain{
		Env: env,
	}
}

func (s Thorchain) PoolAddress() ctypes.AccAddress {
	// TODO : Fix this - this is a hack to get around the 1 query per second REST API limit.
	time.Sleep(1 * time.Second)

	var addrs types.ThorchainPoolAddress

	resp, err := http.Get(s.PoolAddressesURL())
	if err != nil {
		log.Fatalf("Failed getting thorchain: %v\n", err)
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

// GetThorchain : Get the Statehcain pools.
func (s Thorchain) GetPools() types.ThorchainPools {
	// TODO : Fix this - this is a hack to get around the 1 query per second REST API limit.
	time.Sleep(1 * time.Second)

	var pools types.ThorchainPools

	resp, err := http.Get(s.PoolURL())
	if err != nil {
		log.Fatalf("Failed getting thorchain: %v\n", err)
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

func (s Thorchain) GetHeight() int {
	// TODO : Fix this - this is a hack to get around the 1 query per second REST API limit.
	time.Sleep(1 * time.Second)

	var block types.LastBlock

	resp, err := http.Get(s.BlockURL())
	if err != nil {
		log.Fatalf("Failed getting thorchain: %v\n", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed reading body: %v\n", err)
	}

	if err := json.Unmarshal(data, &block); nil != err {
		log.Fatalf("Failed to unmarshal pools: %s", err)
	}

	height, _ := strconv.Atoi(block.Height)
	return height
}

// Scheme : SSL or not.
func (s Thorchain) scheme() string {
	scheme := "https"

	if s.Env == "local" {
		scheme = "http"
	}

	return scheme
}

func (s Thorchain) BlockURL() string {
	return fmt.Sprintf("%v://%v/thorchain/lastblock", s.scheme(), endpoints[s.Env])
}

// PoolURL : Return the Pool URL based on the selected environment.
func (s Thorchain) PoolURL() string {
	return fmt.Sprintf("%v://%v/thorchain/pools", s.scheme(), endpoints[s.Env])
}

// StakerURL  : Return the Staker URL based on the selected environment.
func (s Thorchain) StakerURL(staker string) string {
	return fmt.Sprintf("%v://%v/thorchain/staker/%v", s.scheme(), endpoints[s.Env], staker)
}

func (s Thorchain) PoolAddressesURL() string {
	return fmt.Sprintf("%v://%v/thorchain/pooladdresses", s.scheme(), endpoints[s.Env])
}
