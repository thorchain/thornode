package types

import (
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

func TestPackage(t *testing.T) { TestingT(t) }

var (
	bnbSingleTxFee = sdk.NewUint(37500)
	bnbMultiTxFee  = sdk.NewUint(30000)
)

// Gas Fees
var BNBGasFeeSingleton = common.Gas{
	{Asset: common.BNBAsset, Amount: bnbSingleTxFee},
}

var BNBGasFeeMulti = common.Gas{
	{Asset: common.BNBAsset, Amount: bnbMultiTxFee},
}

type TypesSuite struct{}

var _ = Suite(&TypesSuite{})

func (s TypesSuite) TestHasSuperMajority(c *C) {
	// happy path
	c.Check(HasSuperMajority(3, 4), Equals, true)
	c.Check(HasSuperMajority(2, 3), Equals, true)
	c.Check(HasSuperMajority(4, 4), Equals, true)
	c.Check(HasSuperMajority(1, 1), Equals, true)
	c.Check(HasSuperMajority(67, 100), Equals, true)

	// unhappy path
	c.Check(HasSuperMajority(2, 4), Equals, false)
	c.Check(HasSuperMajority(9, 4), Equals, false)
	c.Check(HasSuperMajority(-9, 4), Equals, false)
	c.Check(HasSuperMajority(9, -4), Equals, false)
	c.Check(HasSuperMajority(0, 0), Equals, false)
	c.Check(HasSuperMajority(3, 0), Equals, false)
}

func (TypesSuite) TestHasSimpleMajority(c *C) {
	c.Check(HasSimpleMajority(3, 4), Equals, true)
	c.Check(HasSimpleMajority(2, 3), Equals, true)
	c.Check(HasSimpleMajority(1, 2), Equals, true)
	c.Check(HasSimpleMajority(1, 3), Equals, false)
	c.Check(HasSimpleMajority(2, 4), Equals, true)
	c.Check(HasSimpleMajority(100000, 3000000), Equals, false)
}

func (TypesSuite) TestGetThreshold(c *C) {
	_, err := GetThreshold(-2)
	c.Assert(err, NotNil)
	output, err := GetThreshold(4)
	c.Assert(err, IsNil)
	c.Assert(output, Equals, 3)
	output, err = GetThreshold(9)
	c.Assert(err, IsNil)
	c.Assert(output, Equals, 6)
	output, err = GetThreshold(10)
	c.Assert(err, IsNil)
	c.Assert(output, Equals, 7)
	output, err = GetThreshold(99)
	c.Assert(err, IsNil)
	c.Assert(output, Equals, 66)
}

func (TypesSuite) TestChooseSignerParty(c *C) {
	// when total is negative number, which is not going to happen when it does, it should return an err
	keys, err := ChooseSignerParty(common.PubKeys{}, 1024, -1)
	c.Assert(err, NotNil)
	c.Assert(keys, HasLen, 0)

	// total 9 signer, 8 available, choose 6
	pubKeys := common.PubKeys{}
	for i := 0; i < 8; i++ {
		pubKeys = append(pubKeys, GetRandomPubKey())
	}
	keys, err = ChooseSignerParty(pubKeys, 1024, 9)
	c.Assert(err, IsNil)
	c.Assert(keys, HasLen, 6)
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})
	// total 9 signer,8 available, choose 6, different seed should return different result
	keys1, err := ChooseSignerParty(pubKeys, 2048, 9)
	c.Assert(err, IsNil)
	c.Assert(keys1, HasLen, 6)
	sort.SliceStable(keys1, func(i, j int) bool {
		return keys1[i].String() < keys1[j].String()
	})
	c.Assert(keys.String() == keys1.String(), Equals, false)

	// same seed should choose the same nodes
	keys2, err := ChooseSignerParty(pubKeys, 1024, 9)
	c.Assert(err, IsNil)
	c.Assert(keys2, HasLen, 6)
	sort.SliceStable(keys2, func(i, j int) bool {
		return keys2[i].String() < keys2[j].String()
	})
	c.Assert(keys.String() == keys2.String(), Equals, true)

	// when there are less nodes than threshold
	keys3, err := ChooseSignerParty(pubKeys[:5], 3096, 9)
	c.Assert(err, NotNil)
	c.Assert(keys3, HasLen, 0)

	// choose all
	keys4, err := ChooseSignerParty(pubKeys[:6], 3096, 9)
	c.Assert(err, IsNil)
	c.Assert(keys4, HasLen, 6)
}
