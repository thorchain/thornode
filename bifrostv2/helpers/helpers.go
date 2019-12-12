package helpers

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client/keys"
	cKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/metrics"
	"gitlab.com/thorchain/thornode/x/thorchain"
)

func SetupStateChainForTest(c *C) (config.ThorChainConfiguration, cKeys.Info, func()) {
	thorchain.SetupConfigForTest()
	thorcliDir := SetupThorCliDirForTest()
	cfg := config.ThorChainConfiguration{
		ChainID:         "statechain",
		ChainHost:       "localhost",
		SignerName:      "bob",
		SignerPasswd:    "password",
		ChainHomeFolder: thorcliDir,
	}
	kb, err := keys.NewKeyBaseFromDir(thorcliDir)
	c.Assert(err, IsNil)
	info, _, err := kb.CreateMnemonic(cfg.SignerName, cKeys.English, cfg.SignerPasswd, cKeys.Secp256k1)
	c.Assert(err, IsNil)
	return cfg, info, func() {
		if err := os.RemoveAll(thorcliDir); nil != err {
			c.Error(err)
		}
	}
}

func SetupThorCliDirForTest() string {
	// Added a rand path so that this method can be called from many test suites so they don't clash.
	rand.Seed(time.Now().UnixNano())
	r := rand.Int63()
	dir := filepath.Join(os.TempDir(), strconv.Itoa(int(r)), ".thorcli")
	return dir
}

func GetMetricForTest(c *C) *metrics.Metrics {
	m, err := metrics.NewMetrics(config.MetricConfiguration{
		Enabled:      false,
		ListenPort:   9000,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	})
	c.Assert(m, NotNil)
	c.Assert(err, IsNil)
	return m
}
