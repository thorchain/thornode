// +build healthcheck

package healthcheck

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

func Test(t *testing.T) {
	TestingT(t)
}

type HealthCheckSuite struct {
	Pools []Pool
}

var _ = Suite(&HealthCheckSuite{})

type Pool struct {
	Asset     common.Asset
	Midgard   MidgardPool
	Thorchain types.Pool
}

type MidgardPool struct {
	Asset      common.Asset     `json:"asset"`
	Status     types.PoolStatus `json:"status"`
	RuneDepth  uint64           `json:"runeDepth"`
	AssetDepth uint64           `json:"assetDepth"`
	PoolUnits  uint64           `json:"poolUnits"`
}

func (s *HealthCheckSuite) SetUpSuite(c *C) {
	thorchain, ok := os.LookupEnv("CHAIN_API")
	if !ok {
		c.Fatal("Missing thorchain api url")
	}
	midgard, ok := os.LookupEnv("MIDGARD_API")
	if !ok {
		c.Fatal("Missing midgard api url")
	}
	thorchainBaseURL := fmt.Sprintf("http://%s/thorchain", thorchain)
	midgardBaseURL := fmt.Sprintf("http://%s/v1", midgard)
	log.Info().Msgf("Testing Pools on thorchain: %s, midgard: %s", thorchainBaseURL, midgardBaseURL)

	// Getting pools to run tests
	s.Pools = []Pool{}
	thorchainPools, err := getThorchainPools(thorchainBaseURL)
	c.Assert(err, IsNil)
	for _, thorchainPool := range thorchainPools {
		midgardPool, err := getMidgardPool(midgardBaseURL, thorchainPool.Asset.String())
		c.Assert(err, IsNil)
		pool := Pool{
			Asset:     midgardPool.Asset,
			Midgard:   midgardPool,
			Thorchain: thorchainPool,
		}
		s.Pools = append(s.Pools, pool)
	}
}

func get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to GET from url %s", url))
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			log.Error().Err(err).Msg("Failed to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Status code: " + strconv.Itoa(resp.StatusCode) + " returned")
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read response body")
	}
	return buf, nil

}

func getThorchainPools(baseURL string) (types.Pools, error) {
	url := fmt.Sprintf("%s/pools", baseURL)
	body, err := get(url)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to GET thorchain pools: %s", url))
	}
	pools := types.Pools{}
	err = json.Unmarshal(body, &pools)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal thorchain pools")
	}
	return pools, nil
}

func getMidgardPool(baseURL string, poolName string) (MidgardPool, error) {
	pool := MidgardPool{}
	url := fmt.Sprintf("%s/pools/%s", baseURL, poolName)
	body, err := get(url)
	if err != nil {
		return pool, errors.Wrap(err, fmt.Sprintf("Failed to GET midgard pool: %s", url))
	}
	err = json.Unmarshal(body, &pool)
	if err != nil {
		return pool, errors.Wrap(err, "Failed to unmarshal midgard pool")
	}
	return pool, nil

}

func (s *HealthCheckSuite) TestPoolBalanceAsset(c *C) {
	for _, pool := range s.Pools {
		c.Check(pool.Midgard.AssetDepth, Equals, pool.Thorchain.BalanceAsset.Uint64(), Commentf("\nPool %s \nMidgard balance asset:   %d \nThorchain balance asset: %d", pool.Asset.String(), pool.Midgard.AssetDepth, pool.Thorchain.BalanceAsset.Uint64()))
	}
}

func (s *HealthCheckSuite) TestPoolBalanceRune(c *C) {
	for _, pool := range s.Pools {
		c.Check(pool.Midgard.RuneDepth, Equals, pool.Thorchain.BalanceRune.Uint64(), Commentf("\nPool %s \nMidgard balance rune:   %d \nThorchain balance rune: %d", pool.Asset.String(), pool.Midgard.RuneDepth, pool.Thorchain.BalanceRune.Uint64()))
	}
}

func (s *HealthCheckSuite) TestPoolUnits(c *C) {
	for _, pool := range s.Pools {
		c.Check(pool.Midgard.PoolUnits, Equals, pool.Thorchain.PoolUnits.Uint64(), Commentf("\nPool %s \nMidgard pool units:   %d \nThorchain pool units: %d", pool.Asset.String(), pool.Midgard.PoolUnits, pool.Thorchain.PoolUnits.Uint64()))
	}
}
