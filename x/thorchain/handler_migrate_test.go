package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type HandlerMigrateSuite struct{}

var _ = Suite(&HandlerMigrateSuite{})

type TestMigrateKeeper struct {
	KVStoreDummy
	activeNodeAccount NodeAccount
	vault             Vault
}

// IsActiveObserver see whether it is an active observer
func (k *TestMigrateKeeper) IsActiveObserver(_ sdk.Context, addr sdk.AccAddress) bool {
	return k.activeNodeAccount.NodeAddress.Equals(addr)
}

func (HandlerMigrateSuite) TestMigrate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestMigrateKeeper{
		activeNodeAccount: GetRandomNodeAccount(NodeActive),
		vault:             GetRandomVault(),
	}

	handler := NewMigrateHandler(keeper)

	addr, err := keeper.vault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)

	ver := semver.MustParse("0.1.0")

	tx := NewObservedTx(common.Tx{
		ID:          GetRandomTxHash(),
		Chain:       common.BNBChain,
		Coins:       common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(1*common.One))},
		Memo:        "",
		FromAddress: GetRandomBNBAddress(),
		ToAddress:   addr,
		Gas:         common.BNBGasFeeSingleton,
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
	handler := NewMigrateHandler(keeper)
	tx := NewObservedTx(common.Tx{
		ID:    GetRandomTxHash(),
		Chain: common.BNBChain,
		Coins: common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(1024)),
		},
		Memo:        NewMigrateMemo(1).String(),
		FromAddress: addr,
		ToAddress:   newVaultAddr,
		Gas:         common.BNBGasFeeSingleton,
	}, 1, retireVault.PubKey)

	msgMigrate := NewMsgMigrate(tx, 1, keeper.activeNodeAccount.NodeAddress)
	result := handler.handleV1(ctx, msgMigrate)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	c.Assert(keeper.txout.TxArray[0].OutHash.Equals(tx.Tx.ID), Equals, true)
}
