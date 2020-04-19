package thorchain

import (
	"encoding/json"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
	. "gopkg.in/check.v1"
)

var _ = Suite(&HandlerErrataTxSuite{})

type HandlerErrataTxSuite struct{}

type TestErrataTxKeeper struct {
	KVStoreDummy
	event Event
	pool  Pool
	na    NodeAccount
	ps    PoolStaker
	err   error
}

func (k *TestErrataTxKeeper) ListActiveNodeAccounts(_ sdk.Context) (NodeAccounts, error) {
	return NodeAccounts{k.na}, k.err
}

func (k *TestErrataTxKeeper) GetNodeAccount(_ sdk.Context, _ sdk.AccAddress) (NodeAccount, error) {
	return k.na, k.err
}

func (k *TestErrataTxKeeper) GetEventsIDByTxHash(_ sdk.Context, _ common.TxID) ([]int64, error) {
	return []int64{1}, k.err
}

func (k *TestErrataTxKeeper) GetEvent(_ sdk.Context, _ int64) (Event, error) {
	return k.event, k.err
}

func (k *TestErrataTxKeeper) UpsertEvent(_ sdk.Context, evt Event) error {
	k.event = evt
	return nil
}

func (k *TestErrataTxKeeper) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.pool, k.err
}

func (k *TestErrataTxKeeper) SetPool(_ sdk.Context, pool Pool) error {
	k.pool = pool
	return k.err
}

func (k *TestErrataTxKeeper) GetPoolStaker(_ sdk.Context, _ common.Asset) (PoolStaker, error) {
	return k.ps, k.err
}

func (k *TestErrataTxKeeper) SetPoolStaker(_ sdk.Context, ps PoolStaker) {
	k.ps = ps
}

func (k *TestErrataTxKeeper) GetErrataTxVoter(_ sdk.Context, txID common.TxID, chain common.Chain) (ErrataTxVoter, error) {
	return NewErrataTxVoter(txID, chain), k.err
}

func (s *HandlerErrataTxSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestErrataTxKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewErrataTxHandler(keeper)
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
	ps := NewPoolStaker(common.BNBAsset, sdk.NewUint(1000))
	addr := GetRandomBNBAddress()
	ps.Stakers = []StakerUnit{
		StakerUnit{
			RuneAddress:  addr,
			AssetAddress: addr,
			Height:       23,
			Units:        ps.TotalUnits,
		},
	}

	keeper := &TestErrataTxKeeper{
		na: na,
		ps: ps,
		pool: Pool{
			Asset:        common.BNBAsset,
			PoolUnits:    ps.TotalUnits,
			BalanceRune:  sdk.NewUint(100 * common.One),
			BalanceAsset: sdk.NewUint(100 * common.One),
		},
		event: Event{
			InTx: common.Tx{
				ID:          txID,
				Chain:       common.BNBChain,
				FromAddress: addr,
				Coins: common.Coins{
					common.NewCoin(common.RuneAsset(), sdk.NewUint(30*common.One)),
				},
				Memo: "STAKE:BNB.BNB",
			},
		},
	}

	handler := NewErrataTxHandler(keeper)

	msg := NewMsgErrataTx(txID, common.BNBChain, na.NodeAddress)
	result := handler.handle(ctx, msg, ver)
	c.Assert(result.IsOK(), Equals, true)
	c.Check(keeper.pool.BalanceRune.Equal(sdk.NewUint(70*common.One)), Equals, true)
	c.Check(keeper.pool.BalanceAsset.Equal(sdk.NewUint(100*common.One)), Equals, true)
	c.Check(keeper.ps.TotalUnits.IsZero(), Equals, true)

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
