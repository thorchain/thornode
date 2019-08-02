package swapservice

import (
	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

const (
	ModuleName = types.ModuleName
	RouterKey  = types.RouterKey
	StoreKey   = types.StoreKey
)

var (
	ModuleCdc     = types.ModuleCdc
	RegisterCodec = types.RegisterCodec
)

type (
	MsgSetPool      = types.MsgSetPool
	MsgSetTxHash    = types.MsgSetTxHash
	QueryResResolve = types.QueryResResolve
	Pool            = types.Pool
	TxHash          = types.TxHash
)
