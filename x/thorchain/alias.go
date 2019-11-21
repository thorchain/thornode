package thorchain

import (
	"gitlab.com/thorchain/bepswap/thornode/x/thorchain/types"
)

const (
	ModuleName       = types.ModuleName
	RouterKey        = types.RouterKey
	StoreKey         = types.StoreKey
	DefaultCodespace = types.DefaultCodespace
	PoolEnabled      = types.Enabled
	PoolBootstrap    = types.Bootstrap
	PoolSuspended    = types.Suspended
	EventSuccess     = types.Success
	EventRefund      = types.Refund

	// Admin config keys
	GSLKey                 = types.GSLKey
	StakerAmtIntervalKey   = types.StakerAmtIntervalKey
	MaxWithdrawBasisPoints = types.MaxWithdrawBasisPoints
	MinValidatorBondKey    = types.MinValidatorBondKey
	WhiteListGasAssetKey   = types.WhiteListGasAssetKey
	PoolRefundGasKey       = types.PoolRefundGasKey
	DefaultPoolStatus      = types.DefaultPoolStatus

	NodeActive      = types.Active
	NodeWhiteListed = types.WhiteListed
	NodeDisabled    = types.Disabled
	NodeReady       = types.Ready
	NodeStandby     = types.Standby
	NodeUnknown     = types.Unknown
)

var (
	NewPool                        = types.NewPool
	NewVaultData                   = types.NewVaultData
	NewAdminConfig                 = types.NewAdminConfig
	NewMsgNoOp                     = types.NewMsgNoOp
	NewMsgAdd                      = types.NewMsgAdd
	NewMsgSetPoolData              = types.NewMsgSetPoolData
	NewMsgSetStakeData             = types.NewMsgSetStakeData
	NewMsgSetUnStake               = types.NewMsgSetUnStake
	NewMsgSwap                     = types.NewMsgSwap
	NewMsgSetAdminConfig           = types.NewMsgSetAdminConfig
	NewTxInVoter                   = types.NewTxInVoter
	NewTxOut                       = types.NewTxOut
	NewEvent                       = types.NewEvent
	NewEventPool                   = types.NewEventPool
	NewEventAdd                    = types.NewEventAdd
	NewEventAdminConfig            = types.NewEventAdminConfig
	NewEventSwap                   = types.NewEventSwap
	NewEventStake                  = types.NewEventStake
	NewEmptyRefundEvent            = types.NewEmptyRefundEvent
	NewEventUnstake                = types.NewEventUnstake
	NewMsgOutboundTx               = types.NewMsgOutboundTx
	NewPoolStaker                  = types.NewPoolStaker
	NewStakerPool                  = types.NewStakerPool
	NewMsgEndPool                  = types.NewMsgEndPool
	NewMsgAck                      = types.NewMsgAck
	HasMajority                    = types.HasMajority
	GetAdminConfigKey              = types.GetAdminConfigKey
	ModuleCdc                      = types.ModuleCdc
	RegisterCodec                  = types.RegisterCodec
	NewTrustAccount                = types.NewTrustAccount
	NewNodeAccount                 = types.NewNodeAccount
	NewYggdrasil                   = types.NewYggdrasil
	NewReserveContributor          = types.NewReserveContributor
	NewMsgYggdrasil                = types.NewMsgYggdrasil
	NewMsgReserveContributor       = types.NewMsgReserveContributor
	NewMsgBond                     = types.NewMsgBond
	NewPoolAddresses               = types.NewPoolAddresses
	NewMsgNextPoolAddress          = types.NewMsgNextPoolAddress
	NewMsgLeave                    = types.NewMsgLeave
	GetPoolStatus                  = types.GetPoolStatus
	GetRandomTx                    = types.GetRandomTx
	GetRandomNodeAccount           = types.GetRandomNodeAccount
	GetRandomBNBAddress            = types.GetRandomBNBAddress
	GetRandomTxHash                = types.GetRandomTxHash
	GetRandomBech32Addr            = types.GetRandomBech32Addr
	GetRandomBech32ConsensusPubKey = types.GetRandomBech32ConsensusPubKey
	GetRandomPubKey                = types.GetRandomPubKey
	GetRandomPubkeys               = types.GetRandomPubKeys
	SetupConfigForTest             = types.SetupConfigForTest
	EmptyPoolAddresses             = types.EmptyPoolAddresses
)

type (
	MsgBond                     = types.MsgBond
	MsgNoOp                     = types.MsgNoOp
	MsgAdd                      = types.MsgAdd
	MsgSetUnStake               = types.MsgSetUnStake
	MsgSetPoolData              = types.MsgSetPoolData
	MsgSetStakeData             = types.MsgSetStakeData
	MsgSetTxIn                  = types.MsgSetTxIn
	MsgOutboundTx               = types.MsgOutboundTx
	MsgSwap                     = types.MsgSwap
	MsgSetAdminConfig           = types.MsgSetAdminConfig
	MsgSetVersion               = types.MsgSetVersion
	MsgSetTrustAccount          = types.MsgSetTrustAccount
	MsgNextPoolAddress          = types.MsgNextPoolAddress
	MsgEndPool                  = types.MsgEndPool
	MsgLeave                    = types.MsgLeave
	MsgAck                      = types.MsgAck
	MsgReserveContributor       = types.MsgReserveContributor
	MsgYggdrasil                = types.MsgYggdrasil
	QueryResPools               = types.QueryResPools
	QueryResHeights             = types.QueryResHeights
	QueryResTxOut               = types.QueryResTxOut
	ResTxOut                    = types.ResTxOut
	TrustAccount                = types.TrustAccount
	TrustAccounts               = types.TrustAccounts
	PoolStatus                  = types.PoolStatus
	PoolIndex                   = types.PoolIndex
	TxInIndex                   = types.TxInIndex
	TxIn                        = types.TxIn
	TxInVoter                   = types.TxInVoter
	Pool                        = types.Pool
	PoolStaker                  = types.PoolStaker
	StakerPool                  = types.StakerPool
	StakerUnit                  = types.StakerUnit
	TxOutItem                   = types.TxOutItem
	TxOut                       = types.TxOut
	AdminConfigKey              = types.AdminConfigKey
	AdminConfig                 = types.AdminConfig
	StakerPoolItem              = types.StakerPoolItem
	StakeTxDetail               = types.StakeTxDetail
	Event                       = types.Event
	Events                      = types.Events
	EventSwap                   = types.EventSwap
	EventStake                  = types.EventStake
	EventStatus                 = types.EventStatus
	ReserveContributor          = types.ReserveContributor
	ReserveContributors         = types.ReserveContributors
	Yggdrasil                   = types.Yggdrasil
	Yggdrasils                  = types.Yggdrasils
	NodeAccount                 = types.NodeAccount
	NodeAccounts                = types.NodeAccounts
	NodeAccountsBySlashingPoint = types.NodeAccountsBySlashingPoint
	PoolAddresses               = types.PoolAddresses
	NodeStatus                  = types.NodeStatus
	ValidatorMeta               = types.ValidatorMeta
	VaultData                   = types.VaultData
)
