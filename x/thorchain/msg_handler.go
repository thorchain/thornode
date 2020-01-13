package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

// MsgHandler is an interface expect all handler to implement
type MsgHandler interface {
	Run(ctx sdk.Context, msg sdk.Msg, version semver.Version, constAccessor constants.ConstantValues) sdk.Result
}
