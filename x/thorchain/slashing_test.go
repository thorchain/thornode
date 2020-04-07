package thorchain

import (
	"encoding/json"
	"errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type SlashingSuite struct{}

var _ = Suite(&SlashingSuite{})

func (s *SlashingSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

type TestSlashObservingKeeper struct {
	KVStoreDummy
	addrs                     []sdk.AccAddress
	nas                       NodeAccounts
	failGetObservingAddress   bool
	failListActiveNodeAccount bool
	failSetNodeAccount        bool
}

func (k *TestSlashObservingKeeper) GetObservingAddresses(_ sdk.Context) ([]sdk.AccAddress, error) {
	if k.failGetObservingAddress {
		return nil, kaboom
	}
	return k.addrs, nil
}

func (k *TestSlashObservingKeeper) ClearObservingAddresses(_ sdk.Context) {
	k.addrs = nil
}

func (k *TestSlashObservingKeeper) ListActiveNodeAccounts(_ sdk.Context) (NodeAccounts, error) {
	if k.failListActiveNodeAccount {
		return nil, kaboom
	}
	return k.nas, nil
}

func (k *TestSlashObservingKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	if k.failSetNodeAccount {
		return kaboom
	}
	for i := range k.nas {
		if k.nas[i].NodeAddress.Equals(na.NodeAddress) {
			k.nas[i] = na
			return nil
		}
	}
	return errors.New("node account not found")
}

func (s *SlashingSuite) TestObservingSlashing(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)

	nas := NodeAccounts{
		GetRandomNodeAccount(NodeActive),
		GetRandomNodeAccount(NodeActive),
	}
	keeper := &TestSlashObservingKeeper{
		nas:   nas,
		addrs: []sdk.AccAddress{nas[0].NodeAddress},
	}
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)

	slasher, err := NewSlasher(keeper, ver)
	c.Assert(err, IsNil)
	// should slash na2 only
	lackOfObservationPenalty := constAccessor.GetInt64Value(constants.LackOfObservationPenalty)
	err = slasher.LackObserving(ctx, constAccessor)
	c.Assert(err, IsNil)
	c.Assert(keeper.nas[0].SlashPoints, Equals, int64(0))
	c.Assert(keeper.nas[1].SlashPoints, Equals, lackOfObservationPenalty)

	// manually clear the observing address, as clear observing address had been moved to moduleManager begin block
	keeper.ClearObservingAddresses(ctx)
	// since THORNode have cleared all node addresses in slashForObservingAddresses,
	// running it a second time should result in slashing nobody.
	err = slasher.LackObserving(ctx, constAccessor)
	c.Assert(err, IsNil)
	c.Assert(keeper.nas[0].SlashPoints, Equals, int64(0))
	c.Assert(keeper.nas[1].SlashPoints, Equals, lackOfObservationPenalty)
}

func (s *SlashingSuite) TestLackObservingErrors(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)

	nas := NodeAccounts{
		GetRandomNodeAccount(NodeActive),
		GetRandomNodeAccount(NodeActive),
	}
	keeper := &TestSlashObservingKeeper{
		nas:   nas,
		addrs: []sdk.AccAddress{nas[0].NodeAddress},
	}
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)
	slasher, err := NewSlasher(keeper, ver)
	c.Assert(err, IsNil)
	keeper.failGetObservingAddress = true
	c.Assert(slasher.LackObserving(ctx, constAccessor), NotNil)
	keeper.failGetObservingAddress = false
	keeper.failListActiveNodeAccount = true
	c.Assert(slasher.LackObserving(ctx, constAccessor), NotNil)
	keeper.failListActiveNodeAccount = false
	keeper.failSetNodeAccount = true
	c.Assert(slasher.LackObserving(ctx, constAccessor), NotNil)
	keeper.failSetNodeAccount = false
}

type TestSlashingLackKeeper struct {
	KVStoreDummy
	evts                       Events
	txOut                      *TxOut
	na                         NodeAccount
	vaults                     Vaults
	voter                      ObservedTxVoter
	failToGetAllPendingEvents  bool
	failGetTxOut               bool
	failGetVault               bool
	failGetNodeAccountByPubKey bool
	failSetNodeAccount         bool
	failGetAsgardByStatus      bool
	failGetObservedTxVoter     bool
	failSetTxOut               bool
}

func (k *TestSlashingLackKeeper) GetObservedTxVoter(_ sdk.Context, _ common.TxID) (ObservedTxVoter, error) {
	if k.failGetObservedTxVoter {
		return ObservedTxVoter{}, kaboom
	}
	return k.voter, nil
}

func (k *TestSlashingLackKeeper) SetObservedTxVoter(_ sdk.Context, voter ObservedTxVoter) {
	k.voter = voter
}

func (k *TestSlashingLackKeeper) GetVault(_ sdk.Context, pk common.PubKey) (Vault, error) {
	if k.failGetVault {
		return Vault{}, kaboom
	}
	return k.vaults[0], nil
}

func (k *TestSlashingLackKeeper) GetAsgardVaultsByStatus(_ sdk.Context, _ VaultStatus) (Vaults, error) {
	if k.failGetAsgardByStatus {
		return nil, kaboom
	}
	return k.vaults, nil
}

func (k *TestSlashingLackKeeper) GetAllPendingEvents(_ sdk.Context) (Events, error) {
	if k.failToGetAllPendingEvents {
		return nil, kaboom
	}
	return k.evts, nil
}

func (k *TestSlashingLackKeeper) GetTxOut(_ sdk.Context, _ int64) (*TxOut, error) {
	if k.failGetTxOut {
		return nil, kaboom
	}
	return k.txOut, nil
}

func (k *TestSlashingLackKeeper) SetTxOut(_ sdk.Context, tx *TxOut) error {
	if k.failSetTxOut {
		return kaboom
	}
	k.txOut = tx
	return nil
}

func (k *TestSlashingLackKeeper) GetNodeAccountByPubKey(_ sdk.Context, _ common.PubKey) (NodeAccount, error) {
	if k.failGetNodeAccountByPubKey {
		return NodeAccount{}, kaboom
	}
	return k.na, nil
}

func (k *TestSlashingLackKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	if k.failSetNodeAccount {
		return kaboom
	}
	k.na = na
	return nil
}

func (s *SlashingSuite) TestNodeSignSlashErrors(c *C) {
	testCases := []struct {
		name        string
		condition   func(keeper *TestSlashingLackKeeper)
		shouldError bool
	}{
		{
			name: "fail to get all pending events should return an error",
			condition: func(keeper *TestSlashingLackKeeper) {
				keeper.failToGetAllPendingEvents = true
			},
			shouldError: true,
		},
		{
			name: "fail to get tx out should return an error",
			condition: func(keeper *TestSlashingLackKeeper) {
				keeper.failGetTxOut = true
			},
			shouldError: false,
		},
		{
			name: "fail to get vault should return an error",
			condition: func(keeper *TestSlashingLackKeeper) {
				keeper.failGetVault = true
			},
			shouldError: false,
		},
		{
			name: "fail to get node account by pub key should return an error",
			condition: func(keeper *TestSlashingLackKeeper) {
				keeper.failGetNodeAccountByPubKey = true
			},
			shouldError: false,
		},
		{
			name: "fail to get asgard vault by status should return an error",
			condition: func(keeper *TestSlashingLackKeeper) {
				keeper.failGetAsgardByStatus = true
			},
			shouldError: true,
		},
		{
			name: "fail to get observed tx voter should return an error",
			condition: func(keeper *TestSlashingLackKeeper) {
				keeper.failGetObservedTxVoter = true
			},
			shouldError: true,
		},
		{
			name: "fail to set tx out should return an error",
			condition: func(keeper *TestSlashingLackKeeper) {
				keeper.failSetTxOut = true
			},
			shouldError: true,
		},
	}
	for _, item := range testCases {
		c.Logf("name:%s", item.name)
		ctx, _ := setupKeeperForTest(c)
		ctx = ctx.WithBlockHeight(201) // set blockheight
		txOutStore := NewTxStoreDummy()
		ver := constants.SWVersion
		constAccessor := constants.GetConstantValues(ver)
		na := GetRandomNodeAccount(NodeActive)

		swapEvt := NewEventSwap(
			common.BNBAsset,
			sdk.NewUint(5),
			sdk.NewUint(5),
			sdk.NewUint(5),
			sdk.NewUint(5),
		)

		swapBytes, _ := json.Marshal(swapEvt)
		evt := NewEvent(swapEvt.Type(),
			3,
			common.NewTx(
				GetRandomTxHash(),
				GetRandomBNBAddress(),
				GetRandomBNBAddress(),
				common.Coins{
					common.NewCoin(common.BNBAsset, sdk.NewUint(320000000)),
					common.NewCoin(common.RuneAsset(), sdk.NewUint(420000000)),
				},
				nil,
				"SWAP:BNB.BNB",
			),
			swapBytes,
			EventSuccess,
		)

		txOutItem := &TxOutItem{
			Chain:       common.BNBChain,
			InHash:      evt.InTx.ID,
			VaultPubKey: na.PubKeySet.Secp256k1,
			ToAddress:   GetRandomBNBAddress(),
			Coin: common.NewCoin(
				common.BNBAsset, sdk.NewUint(3980500*common.One),
			),
		}
		txOut := NewTxOut(evt.Height)
		txOut.TxArray = append(txOut.TxArray, txOutItem)

		ygg := GetRandomVault()
		ygg.Type = YggdrasilVault
		keeper := &TestSlashingLackKeeper{
			txOut:  txOut,
			evts:   Events{evt},
			na:     na,
			vaults: Vaults{ygg},
			voter: ObservedTxVoter{
				Actions: []TxOutItem{*txOutItem},
			},
		}
		signingTransactionPeriod := constAccessor.GetInt64Value(constants.SigningTransactionPeriod)
		ctx = ctx.WithBlockHeight(evt.Height + signingTransactionPeriod)
		version := constants.SWVersion
		slasher, err := NewSlasher(keeper, version)
		c.Assert(err, IsNil)
		item.condition(keeper)
		if item.shouldError {
			c.Assert(slasher.LackSigning(ctx, constAccessor, txOutStore), NotNil)
		} else {
			c.Assert(slasher.LackSigning(ctx, constAccessor, txOutStore), IsNil)
		}
	}
}

func (s *SlashingSuite) TestNotSigningSlash(c *C) {
	ctx, _ := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(201) // set blockheight
	txOutStore := NewTxStoreDummy()
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)
	na := GetRandomNodeAccount(NodeActive)

	swapEvt := NewEventSwap(
		common.BNBAsset,
		sdk.NewUint(5),
		sdk.NewUint(5),
		sdk.NewUint(5),
		sdk.NewUint(5),
	)

	swapBytes, _ := json.Marshal(swapEvt)
	evt := NewEvent(swapEvt.Type(),
		3,
		common.NewTx(
			GetRandomTxHash(),
			GetRandomBNBAddress(),
			GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(common.BNBAsset, sdk.NewUint(320000000)),
				common.NewCoin(common.RuneAsset(), sdk.NewUint(420000000)),
			},
			nil,
			"SWAP:BNB.BNB",
		),
		swapBytes,
		EventSuccess,
	)

	txOutItem := &TxOutItem{
		Chain:       common.BNBChain,
		InHash:      evt.InTx.ID,
		VaultPubKey: na.PubKeySet.Secp256k1,
		ToAddress:   GetRandomBNBAddress(),
		Coin: common.NewCoin(
			common.BNBAsset, sdk.NewUint(3980500*common.One),
		),
	}
	txOut := NewTxOut(evt.Height)
	txOut.TxArray = append(txOut.TxArray, txOutItem)

	ygg := GetRandomVault()
	ygg.Type = YggdrasilVault
	keeper := &TestSlashingLackKeeper{
		txOut:  txOut,
		evts:   Events{evt},
		na:     na,
		vaults: Vaults{ygg},
		voter: ObservedTxVoter{
			Actions: []TxOutItem{*txOutItem},
		},
	}
	signingTransactionPeriod := constAccessor.GetInt64Value(constants.SigningTransactionPeriod)
	ctx = ctx.WithBlockHeight(evt.Height + signingTransactionPeriod)
	version := constants.SWVersion
	slasher, err := NewSlasher(keeper, version)
	c.Assert(err, IsNil)
	c.Assert(slasher.LackSigning(ctx, constAccessor, txOutStore), IsNil)

	c.Check(keeper.na.SlashPoints, Equals, int64(200), Commentf("%+v\n", na))

	outItems, err := txOutStore.GetOutboundItems(ctx)
	c.Assert(err, IsNil)
	c.Assert(outItems, HasLen, 1)
	c.Assert(outItems[0].VaultPubKey.Equals(keeper.vaults[0].PubKey), Equals, true)
	c.Assert(keeper.voter.Actions, HasLen, 1)
	// ensure we've updated our action item
	c.Assert(keeper.voter.Actions[0].VaultPubKey.Equals(outItems[0].VaultPubKey), Equals, true)
}

func (s *SlashingSuite) TestNewSlasher(c *C) {
	nas := NodeAccounts{
		GetRandomNodeAccount(NodeActive),
		GetRandomNodeAccount(NodeActive),
	}
	keeper := &TestSlashObservingKeeper{
		nas:   nas,
		addrs: []sdk.AccAddress{nas[0].NodeAddress},
	}
	ver := semver.MustParse("0.0.1")
	slasher, err := NewSlasher(keeper, ver)
	c.Assert(err, Equals, errBadVersion)
	c.Assert(slasher, IsNil)
}
