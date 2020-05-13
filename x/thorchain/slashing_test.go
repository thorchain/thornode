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
	slashPts                  map[string]int64
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

func (k *TestSlashObservingKeeper) IncNodeAccountSlashPoints(_ sdk.Context, addr sdk.AccAddress, pts int64) error {
	if _, ok := k.slashPts[addr.String()]; !ok {
		k.slashPts[addr.String()] = 0
	}
	k.slashPts[addr.String()] += pts
	return nil
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
		nas:      nas,
		addrs:    []sdk.AccAddress{nas[0].NodeAddress},
		slashPts: make(map[string]int64, 0),
	}
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)

	slasher, err := NewSlasher(keeper, ver, NewVersionedEventMgr())
	c.Assert(err, IsNil)
	// should slash na2 only
	lackOfObservationPenalty := constAccessor.GetInt64Value(constants.LackOfObservationPenalty)
	err = slasher.LackObserving(ctx, constAccessor)
	c.Assert(err, IsNil)
	c.Assert(keeper.slashPts[nas[0].NodeAddress.String()], Equals, int64(0))
	c.Assert(keeper.slashPts[nas[1].NodeAddress.String()], Equals, lackOfObservationPenalty)

	// manually clear the observing address, as clear observing address had been moved to moduleManager begin block
	keeper.ClearObservingAddresses(ctx)
	// since THORNode have cleared all node addresses in slashForObservingAddresses,
	// running it a second time should result in slashing nobody.
	err = slasher.LackObserving(ctx, constAccessor)
	c.Assert(err, IsNil)
	c.Assert(keeper.slashPts[nas[0].NodeAddress.String()], Equals, int64(0))
	c.Assert(keeper.slashPts[nas[1].NodeAddress.String()], Equals, lackOfObservationPenalty)
}

func (s *SlashingSuite) TestLackObservingErrors(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)

	nas := NodeAccounts{
		GetRandomNodeAccount(NodeActive),
		GetRandomNodeAccount(NodeActive),
	}
	keeper := &TestSlashObservingKeeper{
		nas:      nas,
		addrs:    []sdk.AccAddress{nas[0].NodeAddress},
		slashPts: make(map[string]int64, 0),
	}
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)
	slasher, err := NewSlasher(keeper, ver, NewVersionedEventMgr())
	c.Assert(err, IsNil)
	keeper.failGetObservingAddress = true
	c.Assert(slasher.LackObserving(ctx, constAccessor), NotNil)
	keeper.failGetObservingAddress = false
	keeper.failListActiveNodeAccount = true
	c.Assert(slasher.LackObserving(ctx, constAccessor), NotNil)
	keeper.failListActiveNodeAccount = false
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
	slashPts                   map[string]int64
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

func (k *TestSlashingLackKeeper) IncNodeAccountSlashPoints(_ sdk.Context, addr sdk.AccAddress, pts int64) error {
	if _, ok := k.slashPts[addr.String()]; !ok {
		k.slashPts[addr.String()] = 0
	}
	k.slashPts[addr.String()] += pts
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
			name: "fail to get tx out should return an error",
			condition: func(keeper *TestSlashingLackKeeper) {
				keeper.failGetTxOut = true
			},
			shouldError: true,
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
		inTx := common.NewTx(
			GetRandomTxHash(),
			GetRandomBNBAddress(),
			GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(common.BNBAsset, sdk.NewUint(320000000)),
				common.NewCoin(common.RuneAsset(), sdk.NewUint(420000000)),
			},
			nil,
			"SWAP:BNB.BNB",
		)
		swapEvt := NewEventSwap(
			common.BNBAsset,
			sdk.NewUint(5),
			sdk.NewUint(5),
			sdk.NewUint(5),
			sdk.NewUint(5),
			inTx,
		)

		swapBytes, _ := json.Marshal(swapEvt)
		evt := NewEvent(swapEvt.Type(),
			3,
			inTx,
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
			slashPts: make(map[string]int64, 0),
		}
		signingTransactionPeriod := constAccessor.GetInt64Value(constants.SigningTransactionPeriod)
		ctx = ctx.WithBlockHeight(evt.Height + signingTransactionPeriod)
		version := constants.SWVersion
		slasher, err := NewSlasher(keeper, version, NewVersionedEventMgr())
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
	inTx := common.NewTx(
		GetRandomTxHash(),
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(320000000)),
			common.NewCoin(common.RuneAsset(), sdk.NewUint(420000000)),
		},
		nil,
		"SWAP:BNB.BNB",
	)
	swapEvt := NewEventSwap(
		common.BNBAsset,
		sdk.NewUint(5),
		sdk.NewUint(5),
		sdk.NewUint(5),
		sdk.NewUint(5),
		inTx,
	)

	swapBytes, _ := json.Marshal(swapEvt)
	evt := NewEvent(swapEvt.Type(),
		3,
		inTx,
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
		slashPts: make(map[string]int64, 0),
	}
	signingTransactionPeriod := constAccessor.GetInt64Value(constants.SigningTransactionPeriod)
	ctx = ctx.WithBlockHeight(evt.Height + signingTransactionPeriod)
	version := constants.SWVersion
	slasher, err := NewSlasher(keeper, version, NewVersionedEventMgr())
	c.Assert(err, IsNil)
	c.Assert(slasher.LackSigning(ctx, constAccessor, txOutStore), IsNil)

	c.Check(keeper.slashPts[na.NodeAddress.String()], Equals, int64(600), Commentf("%+v\n", na))

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
		nas:      nas,
		addrs:    []sdk.AccAddress{nas[0].NodeAddress},
		slashPts: make(map[string]int64, 0),
	}
	ver := semver.MustParse("0.0.1")
	slasher, err := NewSlasher(keeper, ver, NewVersionedEventMgr())
	c.Assert(err, Equals, errBadVersion)
	c.Assert(slasher, IsNil)
}

type TestDoubleSlashKeeper struct {
	KVStoreDummy
	na          NodeAccount
	vaultData   VaultData
	slashPoints map[string]int64
}

func (k *TestDoubleSlashKeeper) ListActiveNodeAccounts(ctx sdk.Context) (NodeAccounts, error) {
	return NodeAccounts{k.na}, nil
}

func (k *TestDoubleSlashKeeper) SetNodeAccount(ctx sdk.Context, na NodeAccount) error {
	k.na = na
	return nil
}

func (k *TestDoubleSlashKeeper) GetVaultData(ctx sdk.Context) (VaultData, error) {
	return k.vaultData, nil
}

func (k *TestDoubleSlashKeeper) SetVaultData(ctx sdk.Context, data VaultData) error {
	k.vaultData = data
	return nil
}

func (k *TestDoubleSlashKeeper) IncNodeAccountSlashPoints(ctx sdk.Context, addr sdk.AccAddress, pts int64) error {
	k.slashPoints[addr.String()] += pts
	return nil
}

func (k *TestDoubleSlashKeeper) DecNodeAccountSlashPoints(ctx sdk.Context, addr sdk.AccAddress, pts int64) error {
	k.slashPoints[addr.String()] -= pts
	return nil
}

func (s *SlashingSuite) TestDoubleSign(c *C) {
	ctx, _ := setupKeeperForTest(c)
	constAccessor := constants.GetConstantValues(constants.SWVersion)

	na := GetRandomNodeAccount(NodeActive)
	na.Bond = sdk.NewUint(100 * common.One)

	keeper := &TestDoubleSlashKeeper{
		na:        na,
		vaultData: NewVaultData(),
	}
	slasher, err := NewSlasher(keeper, constants.SWVersion, NewVersionedEventMgr())
	c.Assert(err, IsNil)

	pk, err := sdk.GetConsPubKeyBech32(na.ValidatorConsPubKey)
	c.Assert(err, IsNil)
	err = slasher.HandleDoubleSign(ctx, pk.Address(), 0, constAccessor)
	c.Assert(err, IsNil)

	c.Check(keeper.na.Bond.Equal(sdk.NewUint(9995000000)), Equals, true, Commentf("%d", keeper.na.Bond.Uint64()))
	c.Check(keeper.vaultData.TotalReserve.Equal(sdk.NewUint(5000000)), Equals, true)
}

func (s *SlashingSuite) TestIncreaseDecreaseSlashPoints(c *C) {
	ctx, _ := setupKeeperForTest(c)

	na := GetRandomNodeAccount(NodeActive)
	na.Bond = sdk.NewUint(100 * common.One)

	keeper := &TestDoubleSlashKeeper{
		na:          na,
		vaultData:   NewVaultData(),
		slashPoints: make(map[string]int64),
	}
	slasher, err := NewSlasher(keeper, constants.SWVersion, NewVersionedEventMgr())
	c.Assert(err, IsNil)
	addr := GetRandomBech32Addr()
	slasher.IncSlashPoints(ctx, 1, addr)
	slasher.DecSlashPoints(ctx, 1, addr)
	c.Assert(keeper.slashPoints[addr.String()], Equals, int64(0))
}
