package thorchain

import (
	"encoding/json"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

var _ = Suite(&HandlerErrataTxSuite{})

type HandlerErrataTxSuite struct{}

type TestErrataTxKeeper struct {
	KVStoreDummy
	event      Event
	observedTx ObservedTxVoter
	pool       Pool
	na         NodeAccount
	stakers    []Staker
	err        error
}

func (k *TestErrataTxKeeper) ListActiveNodeAccounts(_ sdk.Context) (NodeAccounts, error) {
	return NodeAccounts{k.na}, k.err
}

func (k *TestErrataTxKeeper) GetNodeAccount(_ sdk.Context, _ sdk.AccAddress) (NodeAccount, error) {
	return k.na, k.err
}

func (k *TestErrataTxKeeper) UpsertEvent(_ sdk.Context, evt Event) error {
	k.event = evt
	return nil
}

func (k *TestErrataTxKeeper) GetObservedTxVoter(_ sdk.Context, txID common.TxID) (ObservedTxVoter, error) {
	return k.observedTx, k.err
}

func (k *TestErrataTxKeeper) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.pool, k.err
}

func (k *TestErrataTxKeeper) SetPool(_ sdk.Context, pool Pool) error {
	k.pool = pool
	return k.err
}

func (k *TestErrataTxKeeper) GetStaker(_ sdk.Context, asset common.Asset, addr common.Address) (Staker, error) {
	for _, staker := range k.stakers {
		if staker.RuneAddress.Equals(addr) {
			return staker, k.err
		}
	}
	return Staker{}, k.err
}

func (k *TestErrataTxKeeper) SetStaker(_ sdk.Context, staker Staker) {
	for i, skr := range k.stakers {
		if skr.RuneAddress.Equals(staker.RuneAddress) {
			k.stakers[i] = staker
		}
	}
}

func (k *TestErrataTxKeeper) GetErrataTxVoter(_ sdk.Context, txID common.TxID, chain common.Chain) (ErrataTxVoter, error) {
	return NewErrataTxVoter(txID, chain), k.err
}

func (s *HandlerErrataTxSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestErrataTxKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewErrataTxHandler(keeper, NewDummyVersionedEventMgr())
	// happy path
	ver := constants.SWVersion
	msg := NewMsgErrataTx(GetRandomTxHash(), common.BNBChain, keeper.na.NodeAddress)
	err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, errBadVersion)

	// invalid msg
	msg = MsgErrataTx{}
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

func (s *HandlerErrataTxSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)
	ver := constants.SWVersion

	txID := GetRandomTxHash()
	na := GetRandomNodeAccount(NodeActive)
	addr := GetRandomBNBAddress()
	totalUnits := sdk.NewUint(1600)

	keeper := &TestErrataTxKeeper{
		na: na,
		observedTx: ObservedTxVoter{
			Tx: ObservedTx{
				Tx: common.Tx{
					ID:          txID,
					Chain:       common.BNBChain,
					FromAddress: addr,
					Coins: common.Coins{
						common.NewCoin(common.RuneAsset(), sdk.NewUint(30*common.One)),
					},
					Memo: "STAKE:BNB.BNB",
				},
			},
		},
		pool: Pool{
			Asset:        common.BNBAsset,
			PoolUnits:    totalUnits,
			BalanceRune:  sdk.NewUint(100 * common.One),
			BalanceAsset: sdk.NewUint(100 * common.One),
		},
		stakers: []Staker{
			Staker{
				RuneAddress:     addr,
				LastStakeHeight: 5,
				Units:           totalUnits.QuoUint64(2),
				PendingRune:     sdk.ZeroUint(),
			},
			Staker{
				RuneAddress:     GetRandomBNBAddress(),
				LastStakeHeight: 10,
				Units:           totalUnits.QuoUint64(2),
				PendingRune:     sdk.ZeroUint(),
			},
		},
	}
	versionedEventManager := NewVersionedEventMgr()
	handler := NewErrataTxHandler(keeper, versionedEventManager)
	msg := NewMsgErrataTx(txID, common.BNBChain, na.NodeAddress)
	result := handler.handle(ctx, msg, ver)
	c.Assert(result.IsOK(), Equals, true)
	c.Check(keeper.pool.BalanceRune.Equal(sdk.NewUint(70*common.One)), Equals, true)
	c.Check(keeper.pool.BalanceAsset.Equal(sdk.NewUint(100*common.One)), Equals, true)
	c.Check(keeper.stakers[0].Units.IsZero(), Equals, true, Commentf("%d", keeper.stakers[0].Units.Uint64()))
	c.Check(keeper.stakers[0].LastStakeHeight, Equals, int64(18))

	c.Assert(keeper.event.Type, Equals, "errata")
	var evt EventErrata
	c.Assert(json.Unmarshal(keeper.event.Event, &evt), IsNil)
	c.Check(evt.Pools, HasLen, 1)
	c.Check(evt.Pools[0].Asset.Equals(common.BNBAsset), Equals, true)
	c.Check(evt.Pools[0].RuneAmt.Equal(sdk.NewUint(30*common.One)), Equals, true)
	c.Check(evt.Pools[0].RuneAdd, Equals, false)
	c.Check(evt.Pools[0].AssetAmt.IsZero(), Equals, true)
	c.Check(evt.Pools[0].AssetAdd, Equals, false)
}
