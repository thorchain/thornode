package thorclient

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/keys"
	cKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/x/thorchain"
)

func TestKeys(t *testing.T) { TestingT(t) }

type KeysSuite struct{}

var _ = Suite(&KeysSuite{})

func (*KeysSuite) SetUpSuite(c *C) {
	thorchain.SetupConfigForTest()
}

const (
	signerNameForTest     = `jack`
	signerPasswordForTest = `password`
)

func (*KeysSuite) setupKeysForTest(c *C) string {
	thorcliDir := filepath.Join(os.TempDir(), ".thorcli")
	kb, err := keys.NewKeyBaseFromDir(thorcliDir)
	c.Assert(err, IsNil)
	_, _, err = kb.CreateMnemonic(signerNameForTest, cKeys.English, signerPasswordForTest, cKeys.Secp256k1)
	c.Assert(err, IsNil)
	kb.CloseDB()
	return thorcliDir
}

func (ks *KeysSuite) TestNewKeys(c *C) {
	folder := ks.setupKeysForTest(c)
	defer func() {
		err := os.RemoveAll(folder)
		c.Assert(err, IsNil)
	}()

	k, err := NewKeys(folder, "", signerPasswordForTest)
	c.Assert(err, NotNil)
	c.Assert(k, IsNil)
	k, err = NewKeys(folder, signerNameForTest, "")
	c.Assert(err, NotNil)
	c.Assert(k, IsNil)

	k, err = NewKeys(folder, signerNameForTest, signerPasswordForTest)
	c.Assert(k, NotNil)
	c.Assert(err, IsNil)
	kInfo := k.GetSignerInfo()
	c.Assert(kInfo, NotNil)
	c.Assert(kInfo.GetName(), Equals, signerNameForTest)
	priKey, err := k.GetPrivateKey()
	c.Assert(err, IsNil)
	c.Assert(priKey, NotNil)
	c.Assert(priKey.Bytes(), HasLen, 37)
	kb := k.GetKeybase()
	c.Assert(kb, NotNil)
	kb.CloseDB()
}
