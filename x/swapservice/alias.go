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
	PoolSuspended    = types.Suspended
	PoolBootstrap    = types.Bootstrap
	EventSuccess     = types.Success
	EventRefund      = types.Refund

	// Admin config keys
	GSLKey                 = types.GSLKey
	TSLKey                 = types.TSLKey
	StakerAmtIntervalKey   = types.StakerAmtIntervalKey
	PoolAddressKey         = types.PoolAddressKey
	PoolExpiryKey          = types.PoolExpiryKey
	MRRAKey                = types.MRRAKey
	MaxWithdrawBasisPoints = types.MaxWithdrawBasisPoints
)

var (
	NewPool               = types.NewPool
	NewAdminConfig        = types.NewAdminConfig
	NewMsgNoOp            = types.NewMsgNoOp
	NewMsgAdd             = types.NewMsgAdd
	NewMsgSetPoolData     = types.NewMsgSetPoolData
	NewMsgSetStakeData    = types.NewMsgSetStakeData
	NewMsgSetUnStake      = types.NewMsgSetUnStake
	NewMsgSwap            = types.NewMsgSwap
	NewMsgSetAdminConfig  = types.NewMsgSetAdminConfig
	NewTxOut              = types.NewTxOut
	NewEvent              = types.NewEvent
	NewEventSwap          = types.NewEventSwap
	NewEventStake         = types.NewEventStake
	NewEmptyRefundEvent   = types.NewEmptyRefundEvent
	NewEventUnstake       = types.NewEventUnstake
	NewMsgOutboundTx      = types.NewMsgOutboundTx
	NewPoolStaker         = types.NewPoolStaker
	NewStakerPool         = types.NewStakerPool
	NewMsgEndPool         = types.NewMsgEndPool
	HasMajority           = types.HasMajority
	GetPoolStatus         = types.GetPoolStatus
	GetAdminConfigKey     = types.GetAdminConfigKey
	ModuleCdc             = types.ModuleCdc
	RegisterCodec         = types.RegisterCodec
	PoolAddressExpiryDate = types.PoolAddressExpiryDate
)

type (
	MsgNoOp            = types.MsgNoOp
	MsgAdd             = types.MsgAdd
	MsgSetUnStake      = types.MsgSetUnStake
	MsgSetPoolData     = types.MsgSetPoolData
	MsgSetStakeData    = types.MsgSetStakeData
	MsgSetTxIn         = types.MsgSetTxIn
	MsgOutboundTx      = types.MsgOutboundTx
	MsgSwap            = types.MsgSwap
	MsgSetAdminConfig  = types.MsgSetAdminConfig
	MsgAddTrustAccount = types.MsgAddTrustAccount
	MsgEndPool         = types.MsgEndPool
	QueryResPools      = types.QueryResPools
	QueryResHeights    = types.QueryResHeights
	TrustAccount       = types.TrustAccount
	TrustAccounts      = types.TrustAccounts
	PoolStatus         = types.PoolStatus
	PoolIndex          = types.PoolIndex
	TxInIndex          = types.TxInIndex
	TxIn               = types.TxIn
	TxInVoter          = types.TxInVoter
	Pool               = types.Pool
	PoolStaker         = types.PoolStaker
	StakerPool         = types.StakerPool
	StakerUnit         = types.StakerUnit
	TxOutItem          = types.TxOutItem
	TxOut              = types.TxOut
	AdminConfigKey     = types.AdminConfigKey
	AdminConfig        = types.AdminConfig
	StakerPoolItem     = types.StakerPoolItem
	StakeTxDetail      = types.StakeTxDetail
	Event              = types.Event
	Events             = types.Events
	EventStatus        = types.EventStatus
)
