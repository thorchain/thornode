package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgReserveContributorSuite struct{}

var _ = Suite(&MsgReserveContributorSuite{})

func (s *MsgReserveContributorSuite) TestMsgReserveContributor(c *C) {
	addr := GetRandomBNBAddress()
	amt := sdk.NewUint(378 * common.One)
	res := NewReserveContributor(addr, amt)
	signer := GetRandomBech32Addr()

	msg := NewMsgReserveContributor(res, signer)
	c.Check(msg.Contributor.IsEmpty(), Equals, false)
	c.Check(msg.Signer.Equals(signer), Equals, true)
}
