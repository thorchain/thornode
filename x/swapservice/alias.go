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
	NewMsgSetPoolData  = types.NewMsgSetPoolData
	NewPoolStruct      = types.NewPoolStruct
	NewMsgSetStakeData = types.NewMsgSetStakeData
	NewMsgSwap         = types.NewMsgSwap
	ModuleCdc          = types.ModuleCdc
	RegisterCodec      = types.RegisterCodec
)

type (
	MsgSetPoolData     = types.MsgSetPoolData
	MsgSetStakeData    = types.MsgSetStakeData
	MsgSwap            = types.MsgSwap
	QueryResResolve    = types.QueryResResolve
	QueryResPoolDatas  = types.QueryResPoolDatas
	QueryResStakeDatas = types.QueryResStakeDatas
	PoolStruct         = types.PoolStruct
)
