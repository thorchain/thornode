package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type HandlerMigrateSuite struct{}

var _ = Suite(&HandlerMigrateSuite{})

type TestMigrateKeeper struct {
	KVStoreDummy
	activeNodeAccount NodeAccount
	vault             Vault
}

// GetNodeAccount
func (k *TestMigrateKeeper) GetNodeAccount(_ sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	if k.activeNodeAccount.NodeAddress.Equals(addr) {
		return k.activeNodeAccount, nil
	}
	return NodeAccount{}, nil
}

func (HandlerMigrateSuite) TestMigrate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestMigrateKeeper{
		activeNodeAccount: GetRandomNodeAccount(NodeActive),
		vault:             GetRandomVault(),
	}

	handler := NewMigrateHandler(keeper, NewVersionedEventMgr())

	addr, err := keeper.vault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)

	ver := constants.SWVersion

	tx := NewObservedTx(common.Tx{
		ID:          GetRandomTxHash(),
		Chain:       common.BNBChain,
		Coins:       common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(1*common.One))},
		Memo:        "",
		FromAddress: GetRandomBNBAddress(),
		ToAddress:   addr,
		Gas:         BNBGasFeeSingleton,
	}, 12, GetRandomPubKey())

	msgMigrate := NewMsgMigrate(tx, 1, keeper.activeNodeAccount.NodeAddress)
	err = handler.validate(ctx, msgMigrate, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msgMigrate, semver.Version{})
	c.Assert(err, Equals, errInvalidVersion)

	// invalid msg
	msgMigrate = MsgMigrate{}
	err = handler.validate(ctx, msgMigrate, ver)
	c.Assert(err, NotNil)

	// not signed observer
	msgMigrate = NewMsgMigrate(tx, 1, GetRandomBech32Addr())
	err = handler.validate(ctx, msgMigrate, ver)
	c.Assert(err, Equals, notAuthorized)
}

type TestMigrateKeeperHappyPath struct {
	KVStoreDummy
	activeNodeAccount NodeAccount
	newVault          Vault
	retireVault       Vault
	txout             *TxOut
	pool              Pool
}

func (k *TestMigrateKeeperHappyPath) GetTxOut(ctx sdk.Context, blockHeight int64) (*TxOut, error) {
	if k.txout != nil && k.txout.Height == blockHeight {
		return k.txout, nil
	}
	return nil, kaboom
}

func (k *TestMigrateKeeperHappyPath) SetTxOut(ctx sdk.Context, blockOut *TxOut) error {
	if k.txout.Height == blockOut.Height {
		k.txout = blockOut
		return nil
	}
	return kaboom
}

func (k *TestMigrateKeeperHappyPath) GetNodeAccountByPubKey(_ sdk.Context, _ common.PubKey) (NodeAccount, error) {
	return k.activeNodeAccount, nil
}

func (k *TestMigrateKeeperHappyPath) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	k.activeNodeAccount = na
	return nil
}

func (k *TestMigrateKeeperHappyPath) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.pool, nil
}

func (k *TestMigrateKeeperHappyPath) SetPool(_ sdk.Context, p Pool) error {
	k.pool = p
	return nil
}

func (k *TestMigrateKeeperHappyPath) UpsertEvent(_ sdk.Context, _ Event) error {
	return nil
}

func (HandlerMigrateSuite) TestMigrateHappyPath(c *C) {
	ctx, _ := setupKeeperForTest(c)
	retireVault := GetRandomVault()

	newVault := GetRandomVault()
	txout := NewTxOut(1)
	newVaultAddr, err := newVault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	txout.TxArray = append(txout.TxArray, &TxOutItem{
		Chain:       common.BNBChain,
		InHash:      common.BlankTxID,
		ToAddress:   newVaultAddr,
		VaultPubKey: retireVault.PubKey,
		Coin:        common.NewCoin(common.BNBAsset, sdk.NewUint(1024)),
		Memo:        NewMigrateMemo(1).String(),
	})
	keeper := &TestMigrateKeeperHappyPath{
		activeNodeAccount: GetRandomNodeAccount(NodeActive),
		newVault:          newVault,
		retireVault:       retireVault,
		txout:             txout,
	}
	addr, err := keeper.retireVault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	handler := NewMigrateHandler(keeper, NewVersionedEventMgr())
	tx := NewObservedTx(common.Tx{
		ID:    GetRandomTxHash(),
		Chain: common.BNBChain,
		Coins: common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(1024)),
		},
		Memo:        NewMigrateMemo(1).String(),
		FromAddress: addr,
		ToAddress:   newVaultAddr,
		Gas:         BNBGasFeeSingleton,
	}, 1, retireVault.PubKey)

	msgMigrate := NewMsgMigrate(tx, 1, keeper.activeNodeAccount.NodeAddress)
	result := handler.handleV1(ctx, constants.SWVersion, msgMigrate)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	c.Assert(keeper.txout.TxArray[0].OutHash.Equals(tx.Tx.ID), Equals, true)
}

func (HandlerMigrateSuite) TestSlash(c *C) {
	ctx, _ := setupKeeperForTest(c)
	retireVault := GetRandomVault()

	newVault := GetRandomVault()
	txout := NewTxOut(1)
	newVaultAddr, err := newVault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)

	pool := NewPool()
	pool.Asset = common.BNBAsset
	pool.BalanceAsset = sdk.NewUint(100 * common.One)
	pool.BalanceRune = sdk.NewUint(100 * common.One)
	na := GetRandomNodeAccount(NodeActive)
	na.Bond = sdk.NewUint(100 * common.One)
	keeper := &TestMigrateKeeperHappyPath{
		activeNodeAccount: na,
		newVault:          newVault,
		retireVault:       retireVault,
		txout:             txout,
		pool:              pool,
	}
	addr, err := keeper.retireVault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	handler := NewMigrateHandler(keeper, NewVersionedEventMgr())
	tx := NewObservedTx(common.Tx{
		ID:    GetRandomTxHash(),
		Chain: common.BNBChain,
		Coins: common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(1024)),
		},
		Memo:        NewMigrateMemo(1).String(),
		FromAddress: addr,
		ToAddress:   newVaultAddr,
		Gas:         BNBGasFeeSingleton,
	}, 1, retireVault.PubKey)

	msgMigrate := NewMsgMigrate(tx, 1, keeper.activeNodeAccount.NodeAddress)
	result := handler.handleV1(ctx, constants.SWVersion, msgMigrate)
	c.Assert(result.Code, Equals, sdk.CodeOK, Commentf("%s", result.Log))
	c.Assert(keeper.activeNodeAccount.Bond.Equal(sdk.NewUint(9999998464)), Equals, true, Commentf("%d", keeper.activeNodeAccount.Bond.Uint64()))
}
