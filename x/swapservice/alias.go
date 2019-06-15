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
	NewMsgSetAccData  = types.NewMsgSetAccData
	NewAccStruct      = types.NewAccStruct
	ModuleCdc         = types.ModuleCdc
	RegisterCodec     = types.RegisterCodec
)

type (
	MsgSetPoolData    = types.MsgSetPoolData
	MsgSetAccData     = types.MsgSetAccData
	QueryResResolve   = types.QueryResResolve
	QueryResPoolDatas = types.QueryResPoolDatas
	QueryResAccDatas  = types.QueryResAccDatas
	PoolStruct        = types.PoolStruct
	AccStruct         = types.AccStruct
)
