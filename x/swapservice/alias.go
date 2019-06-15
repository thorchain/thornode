package swapservice

import (
	"github.com/jpthor/test/x/swapservice/types"
)

const (
	ModuleName = types.ModuleName
	RouterKey  = types.RouterKey
	StoreKey   = types.StoreKey
)

var (
	NewMsgSetPoolData = types.NewMsgSetPoolData
	NewPoolStruct     = types.NewPoolStruct
	ModuleCdc         = types.ModuleCdc
	RegisterCodec     = types.RegisterCodec
)

type (
	MsgSetPoolData    = types.MsgSetPoolData
	QueryResResolve   = types.QueryResResolve
	QueryResPoolDatas = types.QueryResPoolDatas
	PoolStruct        = types.PoolStruct
)
