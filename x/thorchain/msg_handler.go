package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/constants"
)

// MsgHandler is an interface expect all handler to implement
type MsgHandler interface {
	Run(_ sdk.Context, _ sdk.Msg, _ constants.Constants, _ semver.Version) sdk.Result
}
