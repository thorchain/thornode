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
	MaxUnstakeBasisPoints = types.MaxUnstakeBasisPoints

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

	BondPaid     = types.BondPaid
	BondReturned = types.BondReturned
	AsgardKeygen = types.AsgardKeygen
)

var (
	NewPool                        = types.NewPool
	NewTxMarker                    = types.NewTxMarker
	NewVaultData                   = types.NewVaultData
	NewObservedTx                  = types.NewObservedTx
	NewTssVoter                    = types.NewTssVoter
	NewErrataTxVoter               = types.NewErrataTxVoter
	NewObservedTxVoter             = types.NewObservedTxVoter
	NewMsgTssPool                  = types.NewMsgTssPool
	NewMsgTssKeysignFail           = types.NewMsgTssKeysignFail
	NewMsgObservedTxIn             = types.NewMsgObservedTxIn
	NewMsgObservedTxOut            = types.NewMsgObservedTxOut
	NewMsgNoOp                     = types.NewMsgNoOp
	NewMsgAdd                      = types.NewMsgAdd
	NewMsgSetStakeData             = types.NewMsgSetStakeData
	NewMsgSetUnStake               = types.NewMsgSetUnStake
	NewMsgSwap                     = types.NewMsgSwap
	NewKeygen                      = types.NewKeygen
	NewKeygenBlock                 = types.NewKeygenBlock
	NewMsgSetNodeKeys              = types.NewMsgSetNodeKeys
	NewTxOut                       = types.NewTxOut
	NewEvent                       = types.NewEvent
	NewEventRewards                = types.NewEventRewards
	NewEventPool                   = types.NewEventPool
	NewEventAdd                    = types.NewEventAdd
	NewEventSwap                   = types.NewEventSwap
	NewEventStake                  = types.NewEventStake
	NewEventUnstake                = types.NewEventUnstake
	NewEventRefund                 = types.NewEventRefund
	NewEventBond                   = types.NewEventBond
	NewEventGas                    = types.NewEventGas
	NewEventSlash                  = types.NewEventSlash
	NewEventReserve                = types.NewEventReserve
	NewMsgRefundTx                 = types.NewMsgRefundTx
	NewMsgOutboundTx               = types.NewMsgOutboundTx
	NewMsgMigrate                  = types.NewMsgMigrate
	NewMsgRagnarok                 = types.NewMsgRagnarok
	NewPoolStaker                  = types.NewPoolStaker
	NewStakerPool                  = types.NewStakerPool
	NewMsgEndPool                  = types.NewMsgEndPool
	HasSuperMajority               = types.HasSuperMajority
	ChooseSignerParty              = types.ChooseSignerParty
	GetThreshold                   = types.GetThreshold
	ModuleCdc                      = types.ModuleCdc
	RegisterCodec                  = types.RegisterCodec
	NewNodeAccount                 = types.NewNodeAccount
	NewVault                       = types.NewVault
	NewReserveContributor          = types.NewReserveContributor
	NewMsgYggdrasil                = types.NewMsgYggdrasil
	NewMsgReserveContributor       = types.NewMsgReserveContributor
	NewMsgBond                     = types.NewMsgBond
	NewMsgErrataTx                 = types.NewMsgErrataTx
	NewMsgLeave                    = types.NewMsgLeave
	NewMsgSetVersion               = types.NewMsgSetVersion
	GetPoolStatus                  = types.GetPoolStatus
	GetRandomVault                 = types.GetRandomVault
	GetRandomTx                    = types.GetRandomTx
	GetRandomObservedTx            = types.GetRandomObservedTx
	GetRandomNodeAccount           = types.GetRandomNodeAccount
	GetRandomBNBAddress            = types.GetRandomBNBAddress
	GetRandomTxHash                = types.GetRandomTxHash
	GetRandomBech32Addr            = types.GetRandomBech32Addr
	GetRandomBech32ConsensusPubKey = types.GetRandomBech32ConsensusPubKey
	GetRandomPubKey                = types.GetRandomPubKey
	GetRandomPubKeySet             = types.GetRandomPubKeySet
	SetupConfigForTest             = types.SetupConfigForTest
	GetEventStatuses               = types.GetEventStatuses
)

type (
	MsgBond                     = types.MsgBond
	MsgNoOp                     = types.MsgNoOp
	MsgAdd                      = types.MsgAdd
	MsgSetUnStake               = types.MsgSetUnStake
	MsgSetStakeData             = types.MsgSetStakeData
	MsgOutboundTx               = types.MsgOutboundTx
	MsgMigrate                  = types.MsgMigrate
	MsgRagnarok                 = types.MsgRagnarok
	MsgRefundTx                 = types.MsgRefundTx
	MsgErrataTx                 = types.MsgErrataTx
	MsgSwap                     = types.MsgSwap
	MsgSetVersion               = types.MsgSetVersion
	MsgSetNodeKeys              = types.MsgSetNodeKeys
	MsgEndPool                  = types.MsgEndPool
	MsgLeave                    = types.MsgLeave
	MsgReserveContributor       = types.MsgReserveContributor
	MsgYggdrasil                = types.MsgYggdrasil
	MsgObservedTxIn             = types.MsgObservedTxIn
	MsgObservedTxOut            = types.MsgObservedTxOut
	MsgTssPool                  = types.MsgTssPool
	MsgTssKeysignFail           = types.MsgTssKeysignFail
	QueryResPools               = types.QueryResPools
	QueryResHeights             = types.QueryResHeights
	QueryResTxOut               = types.QueryResTxOut
	QueryYggdrasilVaults        = types.QueryYggdrasilVaults
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
	ErrataTxVoter               = types.ErrataTxVoter
	TssVoter                    = types.TssVoter
	TssKeysignFailVoter         = types.TssKeysignFailVoter
	TxOutItem                   = types.TxOutItem
	TxOut                       = types.TxOut
	Keygen                      = types.Keygen
	KeygenBlock                 = types.KeygenBlock
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
	NodeStatus                  = types.NodeStatus
	VaultData                   = types.VaultData
	VaultStatus                 = types.VaultStatus
	EventStatuses               = types.EventStatuses
	GasPool                     = types.GasPool
	EventGas                    = types.EventGas
	TxMarker                    = types.TxMarker
	TxMarkers                   = types.TxMarkers
)
