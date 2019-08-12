package swapservice

import (
	"github.com/jpthor/cosmos-swap/x/swapservice/types"
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
)

var (
	NewPoolStruct        = types.NewPoolStruct
	NewAdminConfig       = types.NewAdminConfig
	NewMsgSetTxHash      = types.NewMsgSetTxHash
	NewMsgSetPoolData    = types.NewMsgSetPoolData
	NewMsgSetStakeData   = types.NewMsgSetStakeData
	NewMsgSetUnStake     = types.NewMsgSetUnStake
	NewMsgSwap           = types.NewMsgSwap
	NewMsgSetAdminConfig = types.NewMsgSetAdminConfig
	NewTxOut             = types.NewTxOut
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
)

type (
	MsgSetUnStake       = types.MsgSetUnStake
	MsgUnStakeComplete  = types.MsgUnStakeComplete
	MsgSwapComplete     = types.MsgSwapComplete
	MsgSetPoolData      = types.MsgSetPoolData
	MsgSetStakeData     = types.MsgSetStakeData
	MsgSetTxHash        = types.MsgSetTxHash
	MsgSwap             = types.MsgSwap
	MsgSetAdminConfig   = types.MsgSetAdminConfig
	QueryResPoolStructs = types.QueryResPoolStructs
	TrustAccount        = types.TrustAccount
	SwapRecord          = types.SwapRecord
	UnstakeRecord       = types.UnstakeRecord
	PoolStatus          = types.PoolStatus
	PoolIndex           = types.PoolIndex
	TxHash              = types.TxHash
	TxID                = types.TxID
	PoolStruct          = types.PoolStruct
	TxOutItem           = types.TxOutItem
	TxOut               = types.TxOut
	Coin                = types.Coin
	Ticker              = types.Ticker
	UnitTicker          = types.UnitTicker
	Amount              = types.Amount
	AdminConfigKey      = types.AdminConfigKey
	AdminConfig         = types.AdminConfig
	BnbAddress          = types.BnbAddress
)
