package thorchain

import (
	"sort"
	"testing"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type ThorchainSuite struct{}

var _ = Suite(&ThorchainSuite{})

func (s *ThorchainSuite) TestChurn(c *C) {
	ctx, keeper := setupKeeperForTest(c)
	ver := semver.MustParse("0.1.0")
	consts := constants.GetConstantValues(ver)

	txOutStore := NewTxStoreDummy()
	vaultMgr := NewVaultMgr(keeper, txOutStore)
	validatorMgr := NewValidatorMgr(keeper, txOutStore, vaultMgr)

	// create starting point, vault and four node active node accounts
	vault := GetRandomVault()
	vault.AddFunds(common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
		common.NewCoin(common.BNBAsset, sdk.NewUint(79*common.One)),
	})
	addresses := make([]sdk.AccAddress, 4)
	for i := 0; i <= 3; i++ {
		na := GetRandomNodeAccount(NodeActive)
		addresses[i] = na.NodeAddress
		na.SignerMembership = common.PubKeys{vault.PubKey}
		if i == 0 { // give the first node account slash points
			na.RequestedToLeave = true
		}
		vault.Membership = append(vault.Membership, na.PubKeySet.Secp256k1)
		c.Assert(keeper.SetNodeAccount(ctx, na), IsNil)
	}
	c.Assert(keeper.SetVault(ctx, vault), IsNil)

	// create new node account to rotate in
	na := GetRandomNodeAccount(NodeReady)
	c.Assert(keeper.SetNodeAccount(ctx, na), IsNil)

	// trigger marking bad actors as well as a keygen
	rotateHeight := consts.GetInt64Value(constants.RotatePerBlockHeight)
	ctx = ctx.WithBlockHeight(rotateHeight)
	c.Assert(validatorMgr.BeginBlock(ctx, consts), IsNil)

	// check we've created a keygen, with the correct members
	keygens, err := keeper.GetKeygens(ctx, uint64(ctx.BlockHeight()))
	c.Assert(err, IsNil)
	c.Assert(keygens.Keygens, HasLen, 1)
	expected := append(vault.Membership[1:], na.PubKeySet.Secp256k1)
	// sort our slices so they are in the same order
	sort.Slice(expected, func(i, j int) bool { return expected[i].String() < expected[j].String() })
	sort.Slice(keygens.Keygens[0], func(i, j int) bool { return keygens.Keygens[0][i].String() < keygens.Keygens[0][j].String() })
	c.Assert(expected, HasLen, len(keygens.Keygens[0]))
	for i := range expected {
		c.Assert(expected[i].Equals(keygens.Keygens[0][i]), Equals, true, Commentf("%d: %s <==> %s", i, expected[i], keygens.Keygens[0][i]))
	}

	// generate a tss keygen handler event
	newVaultPk := GetRandomPubKey()
	msg := NewMsgTssPool(keygens.Keygens[0], newVaultPk, addresses[0])
	tssHandler := NewTssHandler(keeper, vaultMgr)

	voter := NewTssVoter(msg.ID, msg.PubKeys, msg.PoolPubKey)
	voter.Signers = addresses // ensure we have consensus, so handler is properly executed
	keeper.SetTssVoter(ctx, voter)

	result := tssHandler.Run(ctx, msg, ver, consts)
	c.Assert(result.IsOK(), Equals, true)

	// check that we've rotated our vaults
	vault1, err := keeper.GetVault(ctx, vault.PubKey)
	c.Assert(err, IsNil)
	c.Assert(vault1.Status == RetiringVault, Equals, true) // first vault should now be retiring
	vault2, err := keeper.GetVault(ctx, newVaultPk)
	c.Assert(err, IsNil)
	c.Assert(vault2.Status == ActiveVault, Equals, true) // new vault should now be active
	c.Assert(vault2.Membership, HasLen, 4)

	// check our validators get rotated appropriately
	validators := validatorMgr.EndBlock(ctx, consts)
	nas, err := keeper.ListActiveNodeAccounts(ctx)
	c.Assert(err, IsNil)
	c.Assert(nas, HasLen, 4)
	c.Assert(validators, HasLen, 4)
	// ensure that the first one is rotated out and the new one is rotated in
	standby, err := keeper.GetNodeAccount(ctx, addresses[0])
	c.Assert(err, IsNil)
	c.Check(standby.Status == NodeStandby, Equals, true)
	na, err = keeper.GetNodeAccount(ctx, na.NodeAddress)
	c.Assert(err, IsNil)
	c.Check(na.Status == NodeActive, Equals, true)

	// check that the funds can be migrated from the retiring vault to the new
	// vault
	ctx = ctx.WithBlockHeight(vault1.StatusSince)
	vaultMgr.EndBlock(ctx, consts) // should attempt to send 20% of the coin values
	vault, err = keeper.GetVault(ctx, vault1.PubKey)
	c.Assert(err, IsNil)
	c.Assert(txOutStore.GetOutboundItems(), HasLen, 2, Commentf("%d", len(txOutStore.GetOutboundItems())))
	item := txOutStore.GetOutboundItems()[0]
	c.Check(item.Coin.Amount.Uint64(), Equals, uint64(2500000000), Commentf("%d", item.Coin.Amount.Uint64()))
	item = txOutStore.GetOutboundItems()[1]
	c.Check(item.Coin.Amount.Uint64(), Equals, uint64(1975000000), Commentf("%d", item.Coin.Amount.Uint64()))
	// check we empty the rest at the last migration event
	migrateInterval := consts.GetInt64Value(constants.FundMigrationInterval)
	ctx = ctx.WithBlockHeight(vault.StatusSince + (migrateInterval * 7))
	vaultMgr.EndBlock(ctx, consts) // should attempt to send 100% of the coin values
	vault, err = keeper.GetVault(ctx, vault.PubKey)
	c.Assert(err, IsNil)
	c.Assert(txOutStore.GetOutboundItems(), HasLen, 4, Commentf("%d", len(txOutStore.GetOutboundItems())))
	item = txOutStore.GetOutboundItems()[2]
	c.Check(item.Coin.Amount.Uint64(), Equals, uint64(10000000000), Commentf("%d", item.Coin.Amount.Uint64()))
	item = txOutStore.GetOutboundItems()[3]
	c.Check(item.Coin.Amount.Uint64(), Equals, uint64(7900000000), Commentf("%d", item.Coin.Amount.Uint64()))
}

func (s *ThorchainSuite) TestRagnarok(c *C) {
	var err error
	ctx, keeper := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(10)
	ver := semver.MustParse("0.1.0")
	consts := constants.GetConstantValues(ver)

	txOutStore := NewTxStoreDummy()
	vaultMgr := NewVaultMgr(keeper, txOutStore)
	validatorMgr := NewValidatorMgr(keeper, txOutStore, vaultMgr)

	// create active asgard vault
	asgard := GetRandomVault()
	c.Assert(keeper.SetVault(ctx, asgard), IsNil)

	// create chains
	keeper.SetChains(ctx, common.Chains{common.BNBChain})

	// create pools
	pool := NewPool()
	pool.Asset = common.BNBAsset
	c.Assert(keeper.SetPool(ctx, pool), IsNil)
	boltAsset, err := common.NewAsset("BNB.BOLT-123")
	c.Assert(err, IsNil)
	pool.Asset = boltAsset
	c.Assert(keeper.SetPool(ctx, pool), IsNil)

	// add stakers
	staker1 := GetRandomBNBAddress() // Staker1
	_, err = stake(ctx, keeper, common.BNBAsset, sdk.NewUint(100*common.One), sdk.NewUint(10*common.One), staker1, staker1, GetRandomTxHash())
	c.Assert(err, IsNil)
	_, err = stake(ctx, keeper, boltAsset, sdk.NewUint(50*common.One), sdk.NewUint(11*common.One), staker1, staker1, GetRandomTxHash())
	c.Assert(err, IsNil)
	staker2 := GetRandomBNBAddress() // staker2
	_, err = stake(ctx, keeper, common.BNBAsset, sdk.NewUint(155*common.One), sdk.NewUint(15*common.One), staker2, staker2, GetRandomTxHash())
	c.Assert(err, IsNil)
	_, err = stake(ctx, keeper, boltAsset, sdk.NewUint(20*common.One), sdk.NewUint(4*common.One), staker2, staker2, GetRandomTxHash())
	c.Assert(err, IsNil)
	staker3 := GetRandomBNBAddress() // staker3
	_, err = stake(ctx, keeper, common.BNBAsset, sdk.NewUint(155*common.One), sdk.NewUint(15*common.One), staker3, staker3, GetRandomTxHash())
	c.Assert(err, IsNil)

	// get new pool data
	bnbPool, err := keeper.GetPool(ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	boltPool, err := keeper.GetPool(ctx, boltAsset)
	c.Assert(err, IsNil)

	// Add bonders/validators
	bonderCount := 3
	bonders := make(NodeAccounts, bonderCount)
	for i := 1; i <= bonderCount; i++ {
		na := GetRandomNodeAccount(NodeActive)
		na.Bond = sdk.NewUint(1_000_000 * uint64(i) * common.One)
		c.Assert(keeper.SetNodeAccount(ctx, na), IsNil)
		bonders[i-1] = na

		// Add bond to asgard
		asgard.AddFunds(common.Coins{
			common.NewCoin(common.RuneAsset(), na.Bond),
		})
		c.Assert(keeper.SetVault(ctx, asgard), IsNil)

		// create yggdrasil vault, with 1/3 of the staked funds
		ygg := GetRandomVault()
		ygg.PubKey = na.PubKeySet.Secp256k1
		ygg.Type = YggdrasilVault
		ygg.AddFunds(common.Coins{
			common.NewCoin(common.RuneAsset(), bnbPool.BalanceRune.QuoUint64(uint64(bonderCount))),
			common.NewCoin(common.BNBAsset, bnbPool.BalanceAsset.QuoUint64(uint64(bonderCount))),
			common.NewCoin(common.RuneAsset(), boltPool.BalanceRune.QuoUint64(uint64(bonderCount))),
			common.NewCoin(boltAsset, boltPool.BalanceAsset.QuoUint64(uint64(bonderCount))),
		})
		c.Assert(keeper.SetVault(ctx, ygg), IsNil)
	}

	// Add reserve contributors
	contrib1 := GetRandomBNBAddress()
	contrib2 := GetRandomBNBAddress()
	reserves := ReserveContributors{
		NewReserveContributor(contrib1, sdk.NewUint(400_000_000*common.One)),
		NewReserveContributor(contrib2, sdk.NewUint(100_000*common.One)),
	}
	c.Assert(keeper.SetReserveContributors(ctx, reserves), IsNil)
	// add reserve to asgard vault
	asgard.AddFunds(common.Coins{
		common.NewCoin(common.RuneAsset(), reserves[0].Amount),
		common.NewCoin(common.RuneAsset(), reserves[1].Amount),
	})
	c.Assert(keeper.SetVault(ctx, asgard), IsNil)

	//////////////////////////////////////////////////////////
	//////////////// Start Ragnarok Protocol /////////////////
	//////////////////////////////////////////////////////////

	active, err := keeper.ListActiveNodeAccounts(ctx)
	c.Assert(err, IsNil)
	// this should trigger stage 1 of the ragnarok protocol. We should see a tx
	// out per node account
	c.Assert(validatorMgr.processRagnarok(ctx, active, consts), IsNil)
	c.Assert(txOutStore.GetOutboundItems(), HasLen, bonderCount)

	// we'll assume the signer does it's job and sends the yggdrasil funds back
	// to asgard, and do it ourselves here manually
	for _, na := range active {
		ygg, err := keeper.GetVault(ctx, na.PubKeySet.Secp256k1)
		c.Assert(err, IsNil)
		asgard.AddFunds(ygg.Coins)
		c.Assert(keeper.SetVault(ctx, asgard), IsNil)
		ygg.SubFunds(ygg.Coins)
		c.Assert(keeper.SetVault(ctx, ygg), IsNil)
	}
	txOutStore.ClearOutboundItems() // clear out txs

	// run stage 2 of ragnarok protocol, nth = 1
	ragnarokHeight, err := keeper.GetRagnarokBlockHeight(ctx)
	c.Assert(err, IsNil)
	migrateInterval := consts.GetInt64Value(constants.FundMigrationInterval)
	for i := 1; i <= 10; i++ { // simulate each round of ragnarok (max of ten)
		ctx = ctx.WithBlockHeight(ragnarokHeight + (int64(i) * migrateInterval))
		c.Assert(validatorMgr.processRagnarok(ctx, active, consts), IsNil)
		items := txOutStore.GetOutboundItems()
		c.Assert(items, HasLen, 15, Commentf("%d", len(items)))

		// validate bonders have correct coin amounts being sent to them on each round of ragnarok
		for _, bonder := range bonders {
			item, ok := txOutStore.GetOutboundItemByToAddress(bonder.BondAddress)
			c.Assert(ok, Equals, true)
			outCoin := common.NewCoin(common.RuneAsset(), calcExpectedValue(bonder.Bond, i))
			c.Assert(item.Coin.Equals(outCoin), Equals, true, Commentf("%+v", item.Coin))
		}

		txOutStore.ClearOutboundItems() // clear out txs
	}

}

// calculate the expected coin amount taken from a original amount at nth round
// of ragnarok
func calcExpectedValue(total sdk.Uint, nth int) sdk.Uint {
	var amt sdk.Uint
	for i := uint64(1); i <= uint64(nth); i++ {
		amt = total.MulUint64(i).QuoUint64(10)
		total = total.Sub(amt)
	}
	return amt
}
