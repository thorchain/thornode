package swapservice

import (
	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

const (
	ModuleName       = types.ModuleName
	RouterKey        = types.RouterKey
	StoreKey         = types.StoreKey
	DefaultCodespace = types.DefaultCodespace
	PoolActive       = types.Active
	PoolSuspended    = types.Suspended
)

var (
	NewPoolStruct           = types.NewPoolStruct
	NewMsgSetTxHash         = types.NewMsgSetTxHash
	NewMsgSetTxHashComplete = types.NewMsgSetTxHashComplete
	NewMsgSetPoolData       = types.NewMsgSetPoolData
	NewMsgSetStakeData      = types.NewMsgSetStakeData
	NewMsgSetUnStake        = types.NewMsgSetUnStake
	NewMsgSwap              = types.NewMsgSwap
	GetPoolNameFromTicker   = types.GetPoolNameFromTicker
	NewTxHash               = types.NewTxHash
	ModuleCdc               = types.ModuleCdc
	RegisterCodec           = types.RegisterCodec
)

type (
	MsgSetPoolData       = types.MsgSetPoolData
	MsgSetStakeData      = types.MsgSetStakeData
	MsgSetTxHash         = types.MsgSetTxHash
	MsgSetTxHashComplete = types.MsgSetTxHashComplete
	MsgSwap              = types.MsgSwap
	QueryResPoolStructs  = types.QueryResPoolStructs
	TxHash               = types.TxHash
	PoolStruct           = types.PoolStruct
	PoolStaker           = types.PoolStaker
	StakerPool           = types.StakerPool
	StakerUnit           = types.StakerUnit
)
