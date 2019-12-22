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

type HandlerRefundSuite struct{}

type TestRefundValidKeepr struct {
	KVStoreDummy
	pool        Pool
	na          NodeAccount
	event       Event
	voter       ObservedTxVoter
	asgardVault Vault
	txOut       TxOut
}

// IsActiveObserver see whether it is an active observer
func (k *TestRefundValidKeepr) IsActiveObserver(_ sdk.Context, addr sdk.AccAddress) bool {
	return k.na.NodeAddress.Equals(addr)
}
func (k *TestRefundValidKeepr) GetNodeAccount(_ sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	if k.na.NodeAddress.Equals(addr) {
		return k.na, nil
	}
	return NodeAccount{}, errors.New("not exist")
}

func (k *TestRefundValidKeepr) UpsertEvent(_ sdk.Context, event Event) error {
	k.event = event
	return nil
}
func (k *TestRefundValidKeepr) GetObservedTxVoter(_ sdk.Context, _ common.TxID) (ObservedTxVoter, error) {
	return k.voter, nil
}

func (k *TestRefundValidKeepr) SetObservedTxVoter(_ sdk.Context, voter ObservedTxVoter) {
	k.voter = voter
}

func (k *TestRefundValidKeepr) GetPendingEventID(_ sdk.Context, _ common.TxID) ([]int64, error) {
	return []int64{k.event.ID}, nil
}
func (k *TestRefundValidKeepr) GetEvent(_ sdk.Context, eventID int64) (Event, error) {
	if eventID == k.event.ID {
		return k.event, nil
	}
	return Event{}, kaboom
}
func (k *TestRefundValidKeepr) VaultExists(_ sdk.Context, _ common.PubKey) bool {
	return !k.asgardVault.IsEmpty()
}

func (k *TestRefundValidKeepr) GetVault(_ sdk.Context, _ common.PubKey) (Vault, error) {
	return k.asgardVault, nil
}

func (k *TestRefundValidKeepr) SetVault(_ sdk.Context, vault Vault) error {
	k.asgardVault = vault
	return nil
}
func (k *TestRefundValidKeepr) GetVaultData(_ sdk.Context) (VaultData, error) {
	return NewVaultData(), nil
}

func (k *TestRefundValidKeepr) SetVaultData(_ sdk.Context, _ VaultData) error {
	return nil
}

func (k *TestRefundValidKeepr) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.pool, nil
}

func (k *TestRefundValidKeepr) SetPool(_ sdk.Context, pool Pool) error {
	k.pool = pool
	return nil
}

func (k *TestRefundValidKeepr) GetTxOut(_ sdk.Context, _ uint64) (*TxOut, error) {
	return &k.txOut, nil
}
func (k *TestRefundValidKeepr) SetTxOut(_ sdk.Context, _ *TxOut) error {
	return nil
}

var _ = Suite(&HandlerRefundSuite{})

func (HandlerRefundSuite) TestRefundValidation(c *C) {
	ctx, _ := setupKeeperForTest(c)

	// message signed by not active account
	k := &TestRefundValidKeepr{}
	refundHandler := NewRefundHandler(k)
	nodeAccount := GetRandomNodeAccount(NodeActive)
	msgRefund := NewMsgRefundTx(GetRandomObservedTx(), GetRandomTx().ID, nodeAccount.NodeAddress)
	ver := semver.MustParse("0.1.0")
	constAccessor := constants.GetConstantValues(ver)
	err := refundHandler.validate(ctx, msgRefund, ver, constAccessor)
	c.Assert(err, NotNil)

	// invalid version
	invalidVer := semver.MustParse("0.0.0")
	err = refundHandler.validate(ctx, msgRefund, invalidVer, constAccessor)
	c.Assert(err, NotNil)
}

func (HandlerRefundSuite) TestRefundHandler_HappyPath(c *C) {
	ctx, _ := setupKeeperForTest(c)
	txIn := GetRandomTx()
	pool := NewPool()
	pool.Status = PoolEnabled
	pool.BalanceRune = sdk.NewUint(100 * common.One)
	pool.BalanceAsset = sdk.NewUint(100 * common.One)
	swapEvt := NewEventSwap(
		common.BNBAsset,

		sdk.ZeroUint(),
		sdk.ZeroUint(),
		sdk.ZeroUint(),
	)
	buf, err := json.Marshal(swapEvt)
	c.Assert(err, IsNil)
	event := NewEvent(swapEvt.Type(), ctx.BlockHeight(), txIn, buf, EventFail)
	na := GetRandomNodeAccount(NodeActive)
	keeper := &TestRefundValidKeepr{
		pool:  pool,
		na:    na,
		voter: NewObservedTxVoter(txIn.ID, make(ObservedTxs, 0)),
		txOut: TxOut{
			Height: uint64(ctx.BlockHeight()),
			TxArray: []*TxOutItem{
				&TxOutItem{
					Chain:       common.BNBChain,
					ToAddress:   txIn.FromAddress,
					VaultPubKey: "",
					SeqNo:       0,
					Coin:        common.NewCoin(common.BNBAsset, sdk.NewUint(1)),
					Memo:        NewRefundMemo(txIn.ID).String(),
					InHash:      txIn.ID,
				},
			},
		},
	}
	c.Assert(keeper.UpsertEvent(ctx, event), IsNil)
	refundHandler := NewRefundHandler(keeper)
	observeTx := GetRandomObservedTx()
	msgRefund := NewMsgRefundTx(observeTx, txIn.ID, na.NodeAddress)
	ver := semver.MustParse("0.1.0")
	result := refundHandler.Run(ctx, msgRefund, ver, constants.GetConstantValues(ver))
	c.Assert(result.Code, Equals, sdk.CodeOK)

}
