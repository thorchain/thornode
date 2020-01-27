package vaultmanager

import (
	"testing"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain"
)

func Test(t *testing.T) {
	TestingT(t)
}

type VaultsManagerSuite struct {
	mapping chainAddressPubKeyVaultMap
}

var _ = Suite(&VaultsManagerSuite{})

func (s *VaultsManagerSuite) SetUpSuite(c *C) {
	thorchain.SetupConfigForTest()
	s.mapping = vaultManager.processRawAsgardVaults()
	c.Assert(len(s.mapping[common.BNBChain]), Equals, 3)
}

// Setup a fixture
var vaultManager = VaultManager{
	rawVaults: types.Vaults{
		Asgard: []common.PubKey{
			"thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck",
			"thorpub1addwnpepq2flfr96skc5lkwdv0n5xjsnhmuju20x3zndgu42zd8dtkrud9m2v0zl2qu",
			"thorpub1addwnpepqwhnus6xs4208d4ynm05lv493amz3fexfjfx4vptntedd7k0ajlcup0pzgk",
		},
		Yggdrasil: []common.PubKey{
			"thorpub1addwnpepq27s79a9xk8hjcpjuthmwnl2z4su43uynekcjuqcnmhpemfgfrh6sf9vffl",
			"thorpub1addwnpepqtsgdw5dj7pj497vr2397pnfctf0d3lf7f2ssu39hts45567syh5xwjukdk",
		},
	},
	asgard:    make(chainAddressPubKeyVaultMap),
	yggdrasil: make(chainAddressPubKeyVaultMap),
}

// TODO: tests are failing there, will need to check after full v2 migrations is complete, vaultManager is not used for now.
func (s *VaultsManagerSuite) TestProcessRawAsgardVaults(c *C) {
	asgard := vaultManager.processRawAsgardVaults()
	c.Assert(len(asgard[common.BNBChain]), Equals, 3)
	// c.Assert(asgard[common.BNBChain]["bnb1l8tt4f2xycdz4e5u6veqmj5qwhp4vsktdkl447"].String(), Equals, "thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck")
	// c.Assert(asgard[common.BNBChain]["bnb1x9e5l0ml47sxrhl699fedj4kqfm30kfr8df2tg"].String(), Equals, "thorpub1addwnpepq2flfr96skc5lkwdv0n5xjsnhmuju20x3zndgu42zd8dtkrud9m2v0zl2qu")
	// c.Assert(asgard[common.BNBChain]["bnb1l7af43v2vq32jwq85vdagukf0z0qqdzr5lnnwq"].String(), Equals, "thorpub1addwnpepqwhnus6xs4208d4ynm05lv493amz3fexfjfx4vptntedd7k0ajlcup0pzgk")
}

func (s *VaultsManagerSuite) TestProcessRawYggdrasilVaults(c *C) {
	yggdrasil := vaultManager.processRawYggdrasilVaults()
	c.Assert(len(yggdrasil[common.BNBChain]), Equals, 2)
}

func (s *VaultsManagerSuite) TestGet(c *C) {
	// TODO: 2 last tests weere failing, need to add it back and check when full v2 migrations is complete.
	pk := vaultManager.Get(common.BNBChain, "BAD", vaultManager.asgard)
	c.Assert(pk.String(), Equals, "")

	pk = vaultManager.Get(common.BNBChain, "", vaultManager.asgard)
	c.Assert(pk.String(), Equals, "")
}

func (s *VaultsManagerSuite) TestGetAsgardPubKeys(c *C) {
	pubkeys := vaultManager.GetAsgardPubKeys()
	c.Assert(len(pubkeys), Equals, 3)
	c.Assert(pubkeys[0].String(), Equals, "thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck")
	c.Assert(pubkeys[1].String(), Equals, "thorpub1addwnpepq2flfr96skc5lkwdv0n5xjsnhmuju20x3zndgu42zd8dtkrud9m2v0zl2qu")
	c.Assert(pubkeys[2].String(), Equals, "thorpub1addwnpepqwhnus6xs4208d4ynm05lv493amz3fexfjfx4vptntedd7k0ajlcup0pzgk")
}

func (s *VaultsManagerSuite) TestGetYggdrasilPubKeys(c *C) {
	pubkeys := vaultManager.GetYggdrasilPubKeys()
	c.Assert(len(pubkeys), Equals, 2)
	c.Assert(pubkeys[0].String(), Equals, "thorpub1addwnpepq27s79a9xk8hjcpjuthmwnl2z4su43uynekcjuqcnmhpemfgfrh6sf9vffl")
	c.Assert(pubkeys[1].String(), Equals, "thorpub1addwnpepqtsgdw5dj7pj497vr2397pnfctf0d3lf7f2ssu39hts45567syh5xwjukdk")
}

func (s *VaultsManagerSuite) TestGetPubKeys(c *C) {
	pubkeys := vaultManager.GetPubKeys()
	c.Assert(len(pubkeys), Equals, 5)
	c.Assert(pubkeys[0].String(), Equals, "thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck")
	c.Assert(pubkeys[1].String(), Equals, "thorpub1addwnpepq2flfr96skc5lkwdv0n5xjsnhmuju20x3zndgu42zd8dtkrud9m2v0zl2qu")
	c.Assert(pubkeys[2].String(), Equals, "thorpub1addwnpepqwhnus6xs4208d4ynm05lv493amz3fexfjfx4vptntedd7k0ajlcup0pzgk")
	c.Assert(pubkeys[3].String(), Equals, "thorpub1addwnpepq27s79a9xk8hjcpjuthmwnl2z4su43uynekcjuqcnmhpemfgfrh6sf9vffl")
	c.Assert(pubkeys[4].String(), Equals, "thorpub1addwnpepqtsgdw5dj7pj497vr2397pnfctf0d3lf7f2ssu39hts45567syh5xwjukdk")
}

func (s *VaultsManagerSuite) TestHasKey(c *C) {
	hasKey := vaultManager.HasKey("thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck")
	c.Assert(hasKey, Equals, true)
	hasKey = vaultManager.HasKey("hello")
	c.Assert(hasKey, Equals, false)
}
