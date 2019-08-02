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
	MsgSetPool   = types.MsgSetPool
	MsgSetTxHash = types.MsgSetTxHash
	QueryPool    = types.QueryPool
	QueryPools   = types.QueryPools
	Pool         = types.Pool
	TxHash       = types.TxHash
)
