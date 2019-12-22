package thorchain

import (
	"gitlab.com/thorchain/thornode/x/thorchain/types"
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
	EventPending     = types.Pending
	EventFail        = types.Failed
	EventRefund      = types.Refund

	// Admin config keys
	MaxWithdrawBasisPoints = types.MaxWithdrawBasisPoints
	WhiteListGasAssetKey   = types.WhiteListGasAssetKey
	PoolRefundGasKey       = types.PoolRefundGasKey
	DefaultPoolStatus      = types.DefaultPoolStatus

	// Vaults
	AsgardVault    = types.AsgardVault
	YggdrasilVault = types.YggdrasilVault
	ActiveVault    = types.ActiveVault
	InactiveVault  = types.InactiveVault
	RetiringVault  = types.RetiringVault

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
	NewObservedTx                  = types.NewObservedTx
	NewTssVoter                    = types.NewTssVoter
	NewObservedTxVoter             = types.NewObservedTxVoter
	NewMsgTssPool                  = types.NewMsgTssPool
	NewMsgObservedTxIn             = types.NewMsgObservedTxIn
	NewMsgObservedTxOut            = types.NewMsgObservedTxOut
	NewMsgNoOp                     = types.NewMsgNoOp
	NewMsgAdd                      = types.NewMsgAdd
	NewMsgSetPoolData              = types.NewMsgSetPoolData
	NewMsgSetStakeData             = types.NewMsgSetStakeData
	NewMsgSetUnStake               = types.NewMsgSetUnStake
	NewMsgSwap                     = types.NewMsgSwap
	NewMsgSetAdminConfig           = types.NewMsgSetAdminConfig
	NewKeygens                     = types.NewKeygens
	NewMsgSetNodeKeys              = types.NewMsgSetNodeKeys
	NewTxOut                       = types.NewTxOut
	NewEvent                       = types.NewEvent
	NewEventRewards                = types.NewEventRewards
	NewEventPool                   = types.NewEventPool
	NewEventAdd                    = types.NewEventAdd
	NewEventAdminConfig            = types.NewEventAdminConfig
	NewEventSwap                   = types.NewEventSwap
	NewEventStake                  = types.NewEventStake
	NewEventUnstake                = types.NewEventUnstake
	NewMsgOutboundTx               = types.NewMsgOutboundTx
	NewPoolStaker                  = types.NewPoolStaker
	NewStakerPool                  = types.NewStakerPool
	NewMsgEndPool                  = types.NewMsgEndPool
	HasMajority                    = types.HasMajority
	GetAdminConfigKey              = types.GetAdminConfigKey
	ModuleCdc                      = types.ModuleCdc
	RegisterCodec                  = types.RegisterCodec
	NewNodeAccount                 = types.NewNodeAccount
	NewVault                       = types.NewVault
	NewReserveContributor          = types.NewReserveContributor
	NewMsgYggdrasil                = types.NewMsgYggdrasil
	NewMsgReserveContributor       = types.NewMsgReserveContributor
	NewMsgBond                     = types.NewMsgBond
	NewPoolAddresses               = types.NewPoolAddresses
	NewMsgLeave                    = types.NewMsgLeave
	NewMsgSetVersion               = types.NewMsgSetVersion
	GetPoolStatus                  = types.GetPoolStatus
	GetRandomTx                    = types.GetRandomTx
	GetRandomObservedTx            = types.GetRandomObservedTx
	GetRandomNodeAccount           = types.GetRandomNodeAccount
	GetRandomBNBAddress            = types.GetRandomBNBAddress
	GetRandomTxHash                = types.GetRandomTxHash
	GetRandomBech32Addr            = types.GetRandomBech32Addr
	GetRandomBech32ConsensusPubKey = types.GetRandomBech32ConsensusPubKey
	GetRandomPubKey                = types.GetRandomPubKey
	GetRandomPubkeys               = types.GetRandomPubKeys
	GetRandomPoolPubKeys           = types.GetRandomPoolPubKeys
	SetupConfigForTest             = types.SetupConfigForTest
)

type (
	MsgBond                     = types.MsgBond
	MsgNoOp                     = types.MsgNoOp
	MsgAdd                      = types.MsgAdd
	MsgSetUnStake               = types.MsgSetUnStake
	MsgSetPoolData              = types.MsgSetPoolData
	MsgSetStakeData             = types.MsgSetStakeData
	MsgOutboundTx               = types.MsgOutboundTx
	MsgRefundTx                 = types.MsgRefundTx
	MsgSwap                     = types.MsgSwap
	MsgSetAdminConfig           = types.MsgSetAdminConfig
	MsgSetVersion               = types.MsgSetVersion
	MsgSetNodeKeys              = types.MsgSetNodeKeys
	MsgEndPool                  = types.MsgEndPool
	MsgLeave                    = types.MsgLeave
	MsgReserveContributor       = types.MsgReserveContributor
	MsgYggdrasil                = types.MsgYggdrasil
	MsgObservedTxIn             = types.MsgObservedTxIn
	MsgObservedTxOut            = types.MsgObservedTxOut
	MsgTssPool                  = types.MsgTssPool
	QueryResPools               = types.QueryResPools
	QueryResHeights             = types.QueryResHeights
	QueryResTxOut               = types.QueryResTxOut
	ResTxOut                    = types.ResTxOut
	NodeKeys                    = types.NodeKeys
	NodesKeys                   = types.NodesKeys
	PoolStatus                  = types.PoolStatus
	PoolIndex                   = types.PoolIndex
	Pool                        = types.Pool
	Pools                       = types.Pools
	PoolStaker                  = types.PoolStaker
	StakerPool                  = types.StakerPool
	StakerUnit                  = types.StakerUnit
	ObservedTxs                 = types.ObservedTxs
	ObservedTx                  = types.ObservedTx
	ObservedTxVoter             = types.ObservedTxVoter
	ObservedTxVoters            = types.ObservedTxVoters
	ObservedTxIndex             = types.ObservedTxIndex
	TssVoter                    = types.TssVoter
	TxOutItem                   = types.TxOutItem
	TxOut                       = types.TxOut
	Keygens                     = types.Keygens
	Keygen                      = types.Keygen
	AdminConfigKey              = types.AdminConfigKey
	AdminConfig                 = types.AdminConfig
	StakerPoolItem              = types.StakerPoolItem
	StakeTxDetail               = types.StakeTxDetail
	Event                       = types.Event
	Events                      = types.Events
	EventSwap                   = types.EventSwap
	EventStake                  = types.EventStake
	EventStatus                 = types.EventStatus
	EventRewards                = types.EventRewards
	PoolAmt                     = types.PoolAmt
	ReserveContributor          = types.ReserveContributor
	ReserveContributors         = types.ReserveContributors
	Vault                       = types.Vault
	Vaults                      = types.Vaults
	NodeAccount                 = types.NodeAccount
	NodeAccounts                = types.NodeAccounts
	NodeAccountsBySlashingPoint = types.NodeAccountsBySlashingPoint
	PoolAddresses               = types.PoolAddresses
	NodeStatus                  = types.NodeStatus
	VaultData                   = types.VaultData
	VaultStatus                 = types.VaultStatus
)
