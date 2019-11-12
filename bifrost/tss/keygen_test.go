package tss

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/keys"
	cKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/bifrost/config"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/bepswap/thornode/x/thorchain"
)

func TestTSSKeyGen(t *testing.T) { TestingT(t) }

type KeyGenTestSuite struct{}

var _ = Suite(&KeyGenTestSuite{})

func (*KeyGenTestSuite) SetUpSuite(c *C) {
	thorchain.SetupConfigForTest()
}

const (
	signerNameForTest     = `jack`
	signerPasswordForTest = `password`
)

func (*KeyGenTestSuite) setupKeysForTest(c *C) string {
	thorcliDir := filepath.Join(os.TempDir(), ".thorcli")
	kb, err := keys.NewKeyBaseFromDir(thorcliDir)
	c.Assert(err, IsNil)
	_, _, err = kb.CreateMnemonic(signerNameForTest, cKeys.English, signerPasswordForTest, cKeys.Secp256k1)
	c.Assert(err, IsNil)
	kb.CloseDB()
	return thorcliDir
}
func (kts *KeyGenTestSuite) TestNewTssKenGen(c *C) {
	folder := kts.setupKeysForTest(c)
	defer func() {
		err := os.RemoveAll(folder)
		c.Assert(err, IsNil)
	}()

	keyGenCfg := config.TSSConfiguration{
		Scheme: "http",
		Host:   "localhost",
		Port:   0,
		NodeId: "whaterver",
	}
	scCfg := config.StateChainConfiguration{
		ChainID:         "thorchain",
		ChainHost:       "localhost",
		ChainHomeFolder: folder,
		SignerName:      signerNameForTest,
		SignerPasswd:    signerPasswordForTest,
	}
	k, err := thorclient.NewKeys(folder, scCfg.SignerName, scCfg.SignerPasswd)
	c.Assert(err, IsNil)
	c.Assert(k, NotNil)
	kg, err := NewTssKeyGen(keyGenCfg, k)
	c.Assert(err, IsNil)
	c.Assert(kg, NotNil)
}
