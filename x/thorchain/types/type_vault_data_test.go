package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type VaultDataSuite struct{}

var _ = Suite(&VaultDataSuite{})

func (s *VaultDataSuite) TestCalcNodeRewards(c *C) {
	vault := VaultData{
		TotalBondUnits: sdk.NewUint(100),
		BondRewardRune: sdk.NewUint(3000),
	}
	reward := vault.CalcNodeRewards(sdk.NewUint(5))
	c.Check(reward.Uint64(), Equals, uint64(150))

	vault = VaultData{
		TotalBondUnits: sdk.NewUint(7357),
		BondRewardRune: sdk.NewUint(275.357 * common.One),
	}
	reward = vault.CalcNodeRewards(sdk.NewUint(78))
	c.Check(reward.Uint64(), Equals, uint64(291937556))

	vault = VaultData{
		TotalBondUnits: sdk.NewUint(7357),
		BondRewardRune: sdk.ZeroUint(),
	}
	reward = vault.CalcNodeRewards(sdk.NewUint(78))
	c.Check(reward.Uint64(), Equals, uint64(0))
}
