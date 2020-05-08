package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type MemoSuite struct{}

var _ = Suite(&MemoSuite{})

func (s *MemoSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (s *MemoSuite) TestTxType(c *C) {
	for _, trans := range []TxType{TxStake, TxUnstake, TxSwap, TxOutbound, TxAdd, TxBond, TxLeave, TxSwitch} {
		tx, err := StringToTxType(trans.String())
		c.Assert(err, IsNil)
		c.Check(tx, Equals, trans)
	}
}

func (s *MemoSuite) TestParseWithAbbreviated(c *C) {
	// happy paths
	memo, err := ParseMemo("%:BNB.RUNE-1BA")
	c.Assert(err, IsNil)
	c.Check(memo.GetAsset().String(), Equals, "BNB.RUNE-1BA")
	c.Check(memo.IsType(TxAdd), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.IsInbound(), Equals, true)

	memo, err = ParseMemo("+:BNB.RUNE-1BA")
	c.Assert(err, IsNil)
	c.Check(memo.GetAsset().String(), Equals, "BNB.RUNE-1BA")
	c.Check(memo.IsType(TxStake), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.IsInbound(), Equals, true)

	memo, err = ParseMemo("-:BNB.RUNE-1BA:25")
	c.Assert(err, IsNil)
	c.Check(memo.GetAsset().String(), Equals, "BNB.RUNE-1BA")
	c.Check(memo.IsType(TxUnstake), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetAmount(), Equals, "25")
	c.Check(memo.IsInbound(), Equals, true)

	memo, err = ParseMemo("=:BNB.RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:870000000")
	c.Assert(err, IsNil)
	c.Check(memo.GetAsset().String(), Equals, "BNB.RUNE-1BA")
	c.Check(memo.IsType(TxSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Log(memo.GetSlipLimit().Uint64())
	c.Check(memo.GetSlipLimit().Equal(sdk.NewUint(870000000)), Equals, true)
	c.Check(memo.IsInbound(), Equals, true)

	memo, err = ParseMemo("=:BNB.RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Assert(err, IsNil)
	c.Check(memo.GetAsset().String(), Equals, "BNB.RUNE-1BA")
	c.Check(memo.IsType(TxSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit().Uint64(), Equals, uint64(0))
	c.Check(memo.IsInbound(), Equals, true)

	memo, err = ParseMemo("=:BNB.RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:")
	c.Assert(err, IsNil)
	c.Check(memo.GetAsset().String(), Equals, "BNB.RUNE-1BA")
	c.Check(memo.IsType(TxSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit().Equal(sdk.ZeroUint()), Equals, true)

	memo, err = ParseMemo("OUTBOUND:MUKVQILIHIAUSEOVAXBFEZAJKYHFJYHRUUYGQJZGFYBYVXCXYNEMUOAIQKFQLLCX")
	c.Assert(err, IsNil)
	c.Check(memo.IsType(TxOutbound), Equals, true, Commentf("%s", memo.GetType()))
	c.Check(memo.IsOutbound(), Equals, true)

	memo, err = ParseMemo("REFUND:MUKVQILIHIAUSEOVAXBFEZAJKYHFJYHRUUYGQJZGFYBYVXCXYNEMUOAIQKFQLLCX")
	c.Assert(err, IsNil)
	c.Check(memo.IsType(TxRefund), Equals, true)
	c.Check(memo.IsOutbound(), Equals, true)

	memo, err = ParseMemo("leave:whatever")
	c.Assert(err, IsNil)
	c.Check(memo.IsType(TxLeave), Equals, true)

	memo, err = ParseMemo("yggdrasil+:30")
	c.Assert(err, IsNil)
	c.Check(memo.IsType(TxYggdrasilFund), Equals, true)
	c.Check(memo.IsInternal(), Equals, true)
	memo, err = ParseMemo("yggdrasil-:30")
	c.Assert(err, IsNil)
	c.Check(memo.IsType(TxYggdrasilReturn), Equals, true)
	c.Check(memo.IsInternal(), Equals, true)
	memo, err = ParseMemo("migrate:100")
	c.Assert(err, IsNil)
	c.Check(memo.IsType(TxMigrate), Equals, true)
	c.Check(memo.IsInternal(), Equals, true)

	memo, err = ParseMemo("ragnarok:100")
	c.Assert(err, IsNil)
	c.Check(memo.IsType(TxRagnarok), Equals, true)
	c.Check(memo.IsInternal(), Equals, true)

	mem := fmt.Sprintf("switch:%s", GetRandomBech32Addr())
	memo, err = ParseMemo(mem)
	c.Assert(err, IsNil)
	c.Check(memo.IsType(TxSwitch), Equals, true)
	c.Check(memo.IsInbound(), Equals, true)

	// unhappy paths
	_, err = ParseMemo("")
	c.Assert(err, NotNil)
	_, err = ParseMemo("bogus")
	c.Assert(err, NotNil)
	_, err = ParseMemo("CREATE") // missing symbol
	c.Assert(err, NotNil)
	_, err = ParseMemo("c:") // bad symbol
	c.Assert(err, NotNil)
	_, err = ParseMemo("-:bnb") // withdraw basis points is optional
	c.Assert(err, IsNil)
	_, err = ParseMemo("-:bnb:twenty-two") // bad amount
	c.Assert(err, NotNil)
	_, err = ParseMemo("=:bnb:bad_DES:5.6") // bad destination
	c.Assert(err, NotNil)
	_, err = ParseMemo(">:bnb:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:five") // bad slip limit
	c.Assert(err, NotNil)
	_, err = ParseMemo("!:key:val") // not enough arguments
	c.Assert(err, NotNil)
	_, err = ParseMemo("!:bogus:key:value") // bogus admin command type
	c.Assert(err, NotNil)
	_, err = ParseMemo("nextpool:whatever")
	c.Assert(err, NotNil)
	_, err = ParseMemo("migrate")
	c.Assert(err, NotNil)
	_, err = ParseMemo("switch")
	c.Assert(err, NotNil)
	_, err = ParseMemo("switch:")
	c.Assert(err, NotNil)
}

func (s *MemoSuite) TestParse(c *C) {
	// happy paths
	memo, err := ParseMemo("add:BNB.RUNE-1BA")
	c.Assert(err, IsNil)
	c.Check(memo.GetAsset().String(), Equals, "BNB.RUNE-1BA")
	c.Check(memo.IsType(TxAdd), Equals, true, Commentf("MEMO: %+v", memo))

	memo, err = ParseMemo("STAKE:BNB.RUNE-1BA")
	c.Assert(err, IsNil)
	c.Check(memo.GetAsset().String(), Equals, "BNB.RUNE-1BA")
	c.Check(memo.IsType(TxStake), Equals, true, Commentf("MEMO: %+v", memo))

	memo, err = ParseMemo("STAKE:BTC.BTC")
	c.Assert(err, NotNil)
	memo, err = ParseMemo("STAKE:BTC.BTC:bc1qwqdg6squsna38e46795at95yu9atm8azzmyvckulcc7kytlcckxswvvzej")
	c.Assert(err, IsNil)
	c.Check(memo.GetDestination().String(), Equals, "bc1qwqdg6squsna38e46795at95yu9atm8azzmyvckulcc7kytlcckxswvvzej")
	c.Check(memo.IsType(TxStake), Equals, true, Commentf("MEMO: %+v", memo))

	memo, err = ParseMemo("WITHDRAW:BNB.RUNE-1BA:25")
	c.Assert(err, IsNil)
	c.Check(memo.GetAsset().String(), Equals, "BNB.RUNE-1BA")
	c.Check(memo.IsType(TxUnstake), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetAmount(), Equals, "25")

	memo, err = ParseMemo("SWAP:BNB.RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:870000000")
	c.Assert(err, IsNil)
	c.Check(memo.GetAsset().String(), Equals, "BNB.RUNE-1BA")
	c.Check(memo.IsType(TxSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Log(memo.GetSlipLimit().String())
	c.Check(memo.GetSlipLimit().Equal(sdk.NewUint(870000000)), Equals, true)

	memo, err = ParseMemo("SWAP:BNB.RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Assert(err, IsNil)
	c.Check(memo.GetAsset().String(), Equals, "BNB.RUNE-1BA")
	c.Check(memo.IsType(TxSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit().Uint64(), Equals, uint64(0))

	memo, err = ParseMemo("SWAP:BNB.RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:")
	c.Assert(err, IsNil)
	c.Check(memo.GetAsset().String(), Equals, "BNB.RUNE-1BA")
	c.Check(memo.IsType(TxSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit().Uint64(), Equals, uint64(0))

	whiteListAddr := GetRandomBech32Addr()
	memo, err = ParseMemo("bond:" + whiteListAddr.String())
	c.Assert(err, IsNil)
	c.Assert(memo.IsType(TxBond), Equals, true)
	c.Assert(memo.GetAccAddress().String(), Equals, whiteListAddr.String())

	memo, err = ParseMemo("leave")
	c.Assert(err, IsNil)
	c.Assert(memo.IsType(TxLeave), Equals, true)

	memo, err = ParseMemo("migrate:100")
	c.Assert(err, IsNil)
	c.Check(memo.IsType(TxMigrate), Equals, true)

	// unhappy paths
	_, err = ParseMemo("")
	c.Assert(err, NotNil)
	_, err = ParseMemo("bogus")
	c.Assert(err, NotNil)
	_, err = ParseMemo("CREATE") // missing symbol
	c.Assert(err, NotNil)
	_, err = ParseMemo("CREATE:") // bad symbol
	c.Assert(err, NotNil)
	_, err = ParseMemo("withdraw:bnb") // withdraw basis points is optional
	c.Assert(err, IsNil)
	_, err = ParseMemo("withdraw:bnb:twenty-two") // bad amount
	c.Assert(err, NotNil)
	_, err = ParseMemo("swap:bnb:STAKER-1:5.6") // bad destination
	c.Assert(err, NotNil)
	_, err = ParseMemo("swap:bnb:bad_DES:5.6") // bad destination
	c.Assert(err, NotNil)
	_, err = ParseMemo("swap:bnb:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:five") // bad slip limit
	c.Assert(err, NotNil)
	_, err = ParseMemo("admin:key:val") // not enough arguments
	c.Assert(err, NotNil)
	_, err = ParseMemo("admin:bogus:key:value") // bogus admin command type
	c.Assert(err, NotNil)
	_, err = ParseMemo("migrate:abc")
	c.Assert(err, NotNil)
}
