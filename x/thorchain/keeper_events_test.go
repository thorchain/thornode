package thorchain

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperEventsSuite struct{}

var _ = Suite(&KeeperEventsSuite{})

func (s *KeeperEventsSuite) TestEvents(c *C) {
	ctx, k := setupKeeperForTest(c)

	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	swap := NewEventSwap(
		common.BNBAsset,
		sdk.NewUint(5),
		sdk.NewUint(5),
		sdk.NewUint(5),
	)

	swapBytes, _ := json.Marshal(swap)
	evt := NewEvent(
		swap.Type(),
		12,
		common.NewTx(
			txID,
			GetRandomBNBAddress(),
			GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(common.BNBAsset, sdk.NewUint(320000000)),
				common.NewCoin(common.RuneAsset(), sdk.NewUint(420000000)),
			},
			common.BNBGasFeeSingleton,
			"SWAP:BNB.BNB",
		),
		swapBytes,
		EventSuccess,
	)

	k.AddIncompleteEvents(ctx, evt)
	evts, err := k.GetIncompleteEvents(ctx)
	c.Assert(err, IsNil)
	c.Assert(evts, HasLen, 1)
	c.Check(evts[0].Height, Equals, int64(12))

	last, err := k.GetLastEventID(ctx)
	c.Assert(err, IsNil)
	c.Check(last, Equals, int64(0))

	tx := common.Tx{ID: common.BlankTxID}
	err = completeEvents(ctx, k, txID, common.Txs{tx})
	c.Assert(err, IsNil)

	last, err = k.GetLastEventID(ctx)
	c.Assert(err, IsNil)
	c.Check(last, Equals, int64(1))

	evt, err = k.GetCompletedEvent(ctx, 1)
	c.Assert(err, IsNil)
	c.Check(evts[0].Height, Equals, int64(12))
}
