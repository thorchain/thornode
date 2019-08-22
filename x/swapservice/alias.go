package swapservice

import (
	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

const (
	ModuleName       = types.ModuleName
	RouterKey        = types.RouterKey
	StoreKey         = types.StoreKey
	DefaultCodespace = types.DefaultCodespace
	PoolEnabled      = types.Enabled
	PoolBootstrap    = types.Bootstrap
	PoolSuspended    = types.Suspended

	// Admin config keys
	GSLKey                 = types.GSLKey
	TSLKey                 = types.TSLKey
	StakerAmtIntervalKey   = types.StakerAmtIntervalKey
	PoolAddressKey         = types.PoolAddressKey
	MRRAKey                = types.MRRAKey
	MaxWithdrawBasisPoints = types.MaxWithdrawBasisPoints
)

var (
	NewPool              = types.NewPool
	NewAdminConfig       = types.NewAdminConfig
	NewMsgNoOp           = types.NewMsgNoOp
	NewMsgAdd            = types.NewMsgAdd
	NewMsgSetTxIn        = types.NewMsgSetTxIn
	NewMsgSetPoolData    = types.NewMsgSetPoolData
	NewMsgSetStakeData   = types.NewMsgSetStakeData
	NewMsgSetUnStake     = types.NewMsgSetUnStake
	NewMsgSwap           = types.NewMsgSwap
	NewMsgSetAdminConfig = types.NewMsgSetAdminConfig
	NewTxOut             = types.NewTxOut
	NewEvent             = types.NewEvent
	NewEventSwap         = types.NewEventSwap
	NewEventStake        = types.NewEventStake
	NewEventUnstake      = types.NewEventUnstake
	NewMsgOutboundTx     = types.NewMsgOutboundTx
	NewPoolStaker        = types.NewPoolStaker
	NewStakerPool        = types.NewStakerPool
	GetPoolStatus        = types.GetPoolStatus
	GetAdminConfigKey    = types.GetAdminConfigKey
	ModuleCdc            = types.ModuleCdc
	RegisterCodec        = types.RegisterCodec
)

type (
	MsgNoOp           = types.MsgNoOp
	MsgAdd            = types.MsgAdd
	MsgSetUnStake     = types.MsgSetUnStake
	MsgSetPoolData    = types.MsgSetPoolData
	MsgSetStakeData   = types.MsgSetStakeData
	MsgSetTxIn        = types.MsgSetTxIn
	MsgOutboundTx     = types.MsgOutboundTx
	MsgSwap           = types.MsgSwap
	MsgSetAdminConfig = types.MsgSetAdminConfig
	QueryResPools     = types.QueryResPools
	TrustAccount      = types.TrustAccount
	PoolStatus        = types.PoolStatus
	PoolIndex         = types.PoolIndex
	TxInIndex         = types.TxInIndex
	TxIn              = types.TxIn
	TxInVoter         = types.TxInVoter
	Pool              = types.Pool
	PoolStaker        = types.PoolStaker
	StakerPool        = types.StakerPool
	StakerUnit        = types.StakerUnit
	TxOutItem         = types.TxOutItem
	TxOut             = types.TxOut
	AdminConfigKey    = types.AdminConfigKey
	AdminConfig       = types.AdminConfig
	StakerPoolItem    = types.StakerPoolItem
	StakeTxDetail     = types.StakeTxDetail
	Event             = types.Event
	Events            = types.Events
)
