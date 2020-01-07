package vaultmanager

import (
	"testing"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrostv2/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain"
)

func Test(t *testing.T) {
	TestingT(t)
}

type VaultsMgrSuite struct {
	mapping chainAddressPubKeyVaultMap
}

var _ = Suite(&VaultsMgrSuite{})

func (s *VaultsMgrSuite) SetUpSuite(c *C) {
	thorchain.SetupConfigForTest()
	s.mapping = vaultMgr.processRawAsgardVaults()
	c.Assert(len(s.mapping[common.BNBChain]), Equals, 3)
}

// Setup a fixture
var vaultMgr = VaultManager{
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

func (s *VaultsMgrSuite) TestProcessRawAsgardVaults(c *C) {
	asgard := vaultMgr.processRawAsgardVaults()
	c.Assert(len(asgard[common.BNBChain]), Equals, 3)
	c.Assert(asgard[common.BNBChain]["bnb1l8tt4f2xycdz4e5u6veqmj5qwhp4vsktdkl447"].String(), Equals, "thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck")
	c.Assert(asgard[common.BNBChain]["bnb1x9e5l0ml47sxrhl699fedj4kqfm30kfr8df2tg"].String(), Equals, "thorpub1addwnpepq2flfr96skc5lkwdv0n5xjsnhmuju20x3zndgu42zd8dtkrud9m2v0zl2qu")
	c.Assert(asgard[common.BNBChain]["bnb1l7af43v2vq32jwq85vdagukf0z0qqdzr5lnnwq"].String(), Equals, "thorpub1addwnpepqwhnus6xs4208d4ynm05lv493amz3fexfjfx4vptntedd7k0ajlcup0pzgk")
	c.Assert(asgard[common.BTCChain]["bc1l8tt4f2xycdz4e5u6veqmj5qwhp4vskt83lp92"].String(), Equals, "thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck")
	c.Assert(asgard[common.BTCChain]["bc1x9e5l0ml47sxrhl699fedj4kqfm30kfrd2f7mu"].String(), Equals, "thorpub1addwnpepq2flfr96skc5lkwdv0n5xjsnhmuju20x3zndgu42zd8dtkrud9m2v0zl2qu")
	c.Assert(asgard[common.BTCChain]["bc1l7af43v2vq32jwq85vdagukf0z0qqdzr7cn875"].String(), Equals, "thorpub1addwnpepqwhnus6xs4208d4ynm05lv493amz3fexfjfx4vptntedd7k0ajlcup0pzgk")
}

func (s *VaultsMgrSuite) TestProcessRawYggdrasilVaults(c *C) {
	yggdrasil := vaultMgr.processRawYggdrasilVaults()
	c.Assert(len(yggdrasil[common.BNBChain]), Equals, 2)
}

func (s *VaultsMgrSuite) TestGet(c *C) {
	pk := vaultMgr.get(common.BNBChain, "BAD", vaultMgr.asgard)
	c.Assert(pk.String(), Equals, "")

	pk = vaultMgr.get(common.BNBChain, "", vaultMgr.asgard)
	c.Assert(pk.String(), Equals, "")

	pk = vaultMgr.get(common.BNBChain, "bnb1l8tt4f2xycdz4e5u6veqmj5qwhp4vsktdkl447", s.mapping)
	c.Assert(pk.String(), Equals, "thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck")

	pk = vaultMgr.get(common.BTCChain, "bc1l7af43v2vq32jwq85vdagukf0z0qqdzr7cn875", s.mapping)
	c.Assert(pk.String(), Equals, "thorpub1addwnpepqwhnus6xs4208d4ynm05lv493amz3fexfjfx4vptntedd7k0ajlcup0pzgk")
}
