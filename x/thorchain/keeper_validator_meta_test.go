package thorchain

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/constants"
)

type KeeperValidatorMetaSuite struct{}

var _ = Suite(&KeeperValidatorMetaSuite{})

func (KeeperValidatorMetaSuite) TestValidatorMeta(c *C) {
	ctx, k := setupKeeperForTest(c)
	meta := ValidatorMeta{
		RotateAtBlockHeight:           constants.RotatePerBlockHeight,
		RotateWindowOpenAtBlockHeight: constants.ValidatorsChangeWindow,
		LeaveOpenWindow:               constants.LeaveProcessPerBlockHeight,
		LeaveProcessAt:                1024,
		Ragnarok:                      false,
	}
	err := k.SetValidatorMeta(ctx, meta)
	c.Assert(err, IsNil)
	vm, err := k.GetValidatorMeta(ctx)
	c.Assert(err, IsNil)
	c.Assert(vm.Ragnarok, Equals, false)
	c.Assert(vm.RotateAtBlockHeight, Equals, int64(constants.RotatePerBlockHeight))
}
