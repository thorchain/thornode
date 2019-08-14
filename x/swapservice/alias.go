package swapservice

import (
	"gitlab.com/thorchain/statechain/x/swapservice/types"
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
	GSLKey               = types.GSLKey
	TSLKey               = types.TSLKey
	StakerAmtIntervalKey = types.StakerAmtIntervalKey
	PoolAddressKey       = types.PoolAddressKey
	MRRAKey              = types.MRRAKey
)

var (
	NewPool              = types.NewPool
	NewAdminConfig       = types.NewAdminConfig
	NewMsgSetTxHash      = types.NewMsgSetTxHash
	NewMsgSetPoolData    = types.NewMsgSetPoolData
	NewMsgSetStakeData   = types.NewMsgSetStakeData
	NewMsgSetUnStake     = types.NewMsgSetUnStake
	NewMsgSwap           = types.NewMsgSwap
	NewMsgSetAdminConfig = types.NewMsgSetAdminConfig
	NewTxOut             = types.NewTxOut
	NewPoolStaker        = types.NewPoolStaker
	NewStakerPool        = types.NewStakerPool
	NewSwapRecord        = types.NewSwapRecord
	IsRune               = types.IsRune
	GetPoolStatus        = types.GetPoolStatus
	GetAdminConfigKey    = types.GetAdminConfigKey
	NewTxID              = types.NewTxID
	NewTicker            = types.NewTicker
	RuneTicker           = types.RuneTicker
	NewAmount            = types.NewAmount
	NewAmountFromFloat   = types.NewAmountFromFloat
	ZeroAmount           = types.ZeroAmount
	NewBnbAddress        = types.NewBnbAddress
	ModuleCdc            = types.ModuleCdc
	RegisterCodec        = types.RegisterCodec
	NewCoin              = types.NewCoin
)

type (
	MsgSetUnStake      = types.MsgSetUnStake
	MsgUnStakeComplete = types.MsgUnStakeComplete
	MsgSwapComplete    = types.MsgSwapComplete
	MsgSetPoolData     = types.MsgSetPoolData
	MsgSetStakeData    = types.MsgSetStakeData
	MsgSetTxHash       = types.MsgSetTxHash
	MsgSwap            = types.MsgSwap
	MsgSetAdminConfig  = types.MsgSetAdminConfig
	QueryResPools      = types.QueryResPools
	TrustAccount       = types.TrustAccount
	SwapRecord         = types.SwapRecord
	UnstakeRecord      = types.UnstakeRecord
	PoolStatus         = types.PoolStatus
	PoolIndex          = types.PoolIndex
	TxHash             = types.TxHash
	TxID               = types.TxID
	Pool               = types.Pool
	PoolStaker         = types.PoolStaker
	StakerPool         = types.StakerPool
	StakerUnit         = types.StakerUnit
	TxOutItem          = types.TxOutItem
	TxOut              = types.TxOut
	Coin               = types.Coin
	Coins              = types.Coins
	Ticker             = types.Ticker
	Amount             = types.Amount
	AdminConfigKey     = types.AdminConfigKey
	AdminConfig        = types.AdminConfig
	BnbAddress         = types.BnbAddress
)
