package thorchain

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	"gitlab.com/thorchain/bepswap/thornode/constants"
	. "gopkg.in/check.v1"
)

type SlashingSuite struct{}

var _ = Suite(&SlashingSuite{})

func (s *SlashingSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (s *SlashingSuite) TestObservingSlashing(c *C) {
	var err error
	ctx, k := setupKeeperForTest(c)

	// add one
	na1 := GetRandomNodeAccount(NodeActive)
	k.SetNodeAccount(ctx, na1)

	// add two
	na2 := GetRandomNodeAccount(NodeActive)
	k.SetNodeAccount(ctx, na2)

	k.AddObservingAddresses(ctx, []sdk.AccAddress{na1.NodeAddress})

	// should slash na2 only
	slashForObservingAddresses(ctx, k)

	na1, err = k.GetNodeAccount(ctx, na1.NodeAddress)
	c.Assert(err, IsNil)
	na2, err = k.GetNodeAccount(ctx, na2.NodeAddress)
	c.Assert(err, IsNil)

	c.Assert(na1.SlashPoints, Equals, int64(0))
	c.Assert(na2.SlashPoints, Equals, int64(constants.LackOfObservationPenalty))

	// since we have cleared all node addresses in slashForObservingAddresses,
	// running it a second time should result in slashing nobody.
	slashForObservingAddresses(ctx, k)
	c.Assert(na1.SlashPoints, Equals, int64(0))
	c.Assert(na2.SlashPoints, Equals, int64(constants.LackOfObservationPenalty))
}

func (s *SlashingSuite) TestNotSigningSlash(c *C) {
	var err error
	ctx, k := setupKeeperForTest(c)
	ctx.WithBlockHeight(201) // set blockheight
	poolAddrMgr := NewPoolAddressManager(k)
	poolAddrMgr.BeginBlock(ctx)
	poolPubKey := GetRandomPubKey()
	pk1, err := common.NewPoolPubKey(common.BNBChain, 0, poolPubKey)
	c.Assert(err, IsNil)
	poolAddrMgr.currentPoolAddresses.Current = common.PoolPubKeys{pk1}
	txOutStore := NewTxOutStore(k, poolAddrMgr)
	txOutStore.NewBlock(uint64(201))

	na := GetRandomNodeAccount(NodeActive)
	k.SetNodeAccount(ctx, na)

	swapEvt := NewEventSwap(
		common.BNBAsset,
		sdk.NewUint(5),
		sdk.NewUint(5),
		sdk.NewDec(5),
	)

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

	swapBytes, _ := json.Marshal(swapEvt)
	evt := NewEvent(
		swapEvt.Type(),
		3,
		inTx,
		swapBytes,
		EventSuccess,
	)

	k.AddIncompleteEvents(ctx, evt)

	txOutItem := &TxOutItem{
		Chain:       common.BNBChain,
		InHash:      inTx.ID,
		PoolAddress: na.NodePubKey.Secp256k1,
		ToAddress:   GetRandomBNBAddress(),
		Coin: common.NewCoin(
			common.BNBAsset, sdk.NewUint(3980500*common.One),
		),
	}
	txs := NewTxOut(uint64(evt.Height))
	txs.TxArray = append(txs.TxArray, txOutItem)
	k.SetTxOut(ctx, txs)

	outItems := txOutStore.GetOutboundItems()
	c.Assert(outItems, HasLen, 0)

	slashForNotSigning(ctx, k, txOutStore)

	na, err = k.GetNodeAccount(ctx, na.NodeAddress)
	c.Assert(err, IsNil)
	c.Check(na.SlashPoints, Equals, int64(200), Commentf("%+v\n", na))

	outItems = txOutStore.GetOutboundItems()
	c.Assert(outItems, HasLen, 1)
	c.Assert(outItems[0].PoolAddress.Equals(poolPubKey), Equals, true)
}
