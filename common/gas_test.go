package common

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type GasSuite struct{}

var _ = Suite(&GasSuite{})

func (s GasSuite) TestMultiGasCalc(c *C) {
	fmt.Println("FOO1")
	gas := GetBNBGasFeeMulti(1)
	amt := gas[0].Amount
	c.Check(amt.Equal(sdk.NewUint(30000)), Equals, true, Commentf("%d", amt.Uint64()))

	fmt.Println("BAR1")
	gas = GetBNBGasFeeMulti(3)
	amt = gas[0].Amount
	c.Check(amt.Equal(sdk.NewUint(90000)), Equals, true, Commentf("%d", amt.Uint64()))
}
