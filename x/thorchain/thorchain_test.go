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
