package thorchain

import (
	"github.com/cosmos/cosmos-sdk/x/bank"

	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

const (
	ModuleName       = types.ModuleName
	ReserveName      = types.ReserveName
	AsgardName       = types.AsgardName
	BondName         = types.BondName
	RouterKey        = types.RouterKey
	StoreKey         = types.StoreKey
	DefaultCodespace = types.DefaultCodespace

	// pool status
	PoolEnabled   = types.Enabled
	PoolBootstrap = types.Bootstrap
	PoolSuspended = types.Suspended

	// event status
	EventSuccess = types.Success
	EventPending = types.Pending
	EventFail    = types.Failed
	RefundStatus = types.Refund

	// Admin config keys
	MaxUnstakeBasisPoints = types.MaxUnstakeBasisPoints

	// Vaults
	AsgardVault    = types.AsgardVault
	YggdrasilVault = types.YggdrasilVault
	ActiveVault    = types.ActiveVault
	InactiveVault  = types.InactiveVault
	RetiringVault  = types.RetiringVault

	// Node status
	NodeActive      = types.Active
	NodeWhiteListed = types.WhiteListed
	NodeDisabled    = types.Disabled
	NodeReady       = types.Ready
	NodeStandby     = types.Standby
	NodeUnknown     = types.Unknown

	// Bond type
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
	NewBanVoter                    = types.NewBanVoter
	NewErrataTxVoter               = types.NewErrataTxVoter
	NewObservedTxVoter             = types.NewObservedTxVoter
	NewMsgMimir                    = types.NewMsgMimir
	NewMsgNativeTx                 = types.NewMsgNativeTx
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
	NewEventErrata                 = types.NewEventErrata
	NewEventFee                    = types.NewEventFee
	NewEventOutbound               = types.NewEventOutbound
	NewPoolMod                     = types.NewPoolMod
	NewMsgRefundTx                 = types.NewMsgRefundTx
	NewMsgOutboundTx               = types.NewMsgOutboundTx
	NewMsgMigrate                  = types.NewMsgMigrate
	NewMsgRagnarok                 = types.NewMsgRagnarok
	NewQueryNodeAccount            = types.NewQueryNodeAccount
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
	NewMsgBan                      = types.NewMsgBan
	NewMsgSwitch                   = types.NewMsgSwitch
	NewMsgLeave                    = types.NewMsgLeave
	NewMsgSetVersion               = types.NewMsgSetVersion
	NewMsgSetIPAddress             = types.NewMsgSetIPAddress
	GetPoolStatus                  = types.GetPoolStatus
	GetRandomVault                 = types.GetRandomVault
	GetRandomTx                    = types.GetRandomTx
	GetRandomObservedTx            = types.GetRandomObservedTx
	GetRandomNodeAccount           = types.GetRandomNodeAccount
	GetRandomTHORAddress           = types.GetRandomTHORAddress
	GetRandomRUNEAddress           = types.GetRandomRUNEAddress
	GetRandomBNBAddress            = types.GetRandomBNBAddress
	GetRandomBTCAddress            = types.GetRandomBTCAddress
	GetRandomTxHash                = types.GetRandomTxHash
	GetRandomBech32Addr            = types.GetRandomBech32Addr
	GetRandomBech32ConsensusPubKey = types.GetRandomBech32ConsensusPubKey
	GetRandomPubKey                = types.GetRandomPubKey
	GetRandomPubKeySet             = types.GetRandomPubKeySet
	SetupConfigForTest             = types.SetupConfigForTest
	GetEventStatuses               = types.GetEventStatuses
)

type (
	MsgSend               = bank.MsgSend
	MsgNativeTx           = types.MsgNativeTx
	MsgSwitch             = types.MsgSwitch
	MsgBond               = types.MsgBond
	MsgNoOp               = types.MsgNoOp
	MsgAdd                = types.MsgAdd
	MsgSetUnStake         = types.MsgSetUnStake
	MsgSetStakeData       = types.MsgSetStakeData
	MsgOutboundTx         = types.MsgOutboundTx
	MsgMimir              = types.MsgMimir
	MsgMigrate            = types.MsgMigrate
	MsgRagnarok           = types.MsgRagnarok
	MsgRefundTx           = types.MsgRefundTx
	MsgErrataTx           = types.MsgErrataTx
	MsgBan                = types.MsgBan
	MsgSwap               = types.MsgSwap
	MsgSetVersion         = types.MsgSetVersion
	MsgSetIPAddress       = types.MsgSetIPAddress
	MsgSetNodeKeys        = types.MsgSetNodeKeys
	MsgLeave              = types.MsgLeave
	MsgReserveContributor = types.MsgReserveContributor
	MsgYggdrasil          = types.MsgYggdrasil
	MsgObservedTxIn       = types.MsgObservedTxIn
	MsgObservedTxOut      = types.MsgObservedTxOut
	MsgTssPool            = types.MsgTssPool
	MsgTssKeysignFail     = types.MsgTssKeysignFail
	QueryResPools         = types.QueryResPools
	QueryResHeights       = types.QueryResHeights
	QueryResTxOut         = types.QueryResTxOut
	QueryYggdrasilVaults  = types.QueryYggdrasilVaults
	QueryNodeAccount      = types.QueryNodeAccount
	ResTxOut              = types.ResTxOut
	NodeKeys              = types.NodeKeys
	NodesKeys             = types.NodesKeys
	PoolStatus            = types.PoolStatus
	Pool                  = types.Pool
	Pools                 = types.Pools
	Staker                = types.Staker
	ObservedTxs           = types.ObservedTxs
	ObservedTx            = types.ObservedTx
	ObservedTxVoter       = types.ObservedTxVoter
	ObservedTxVoters      = types.ObservedTxVoters
	ObservedTxIndex       = types.ObservedTxIndex
	BanVoter              = types.BanVoter
	ErrataTxVoter         = types.ErrataTxVoter
	TssVoter              = types.TssVoter
	TssKeysignFailVoter   = types.TssKeysignFailVoter
	TxOutItem             = types.TxOutItem
	TxOut                 = types.TxOut
	Keygen                = types.Keygen
	KeygenBlock           = types.KeygenBlock
	Event                 = types.Event
	Events                = types.Events
	EventSwap             = types.EventSwap
	EventStake            = types.EventStake
	EventUnstake          = types.EventUnstake
	EventStatus           = types.EventStatus
	EventAdd              = types.EventAdd
	EventRewards          = types.EventRewards
	EventErrata           = types.EventErrata
	EventReserve          = types.EventReserve
	PoolAmt               = types.PoolAmt
	PoolMod               = types.PoolMod
	PoolMods              = types.PoolMods
	ReserveContributor    = types.ReserveContributor
	ReserveContributors   = types.ReserveContributors
	Vault                 = types.Vault
	Vaults                = types.Vaults
	NodeAccount           = types.NodeAccount
	NodeAccounts          = types.NodeAccounts
	NodeStatus            = types.NodeStatus
	VaultData             = types.VaultData
	VaultStatus           = types.VaultStatus
	EventStatuses         = types.EventStatuses
	GasPool               = types.GasPool
	EventGas              = types.EventGas
	TxMarker              = types.TxMarker
	TxMarkers             = types.TxMarkers
	EventPool             = types.EventPool
	EventRefund           = types.EventRefund
	EventBond             = types.EventBond
	EventFee              = types.EventFee
	EventSlash            = types.EventSlash
	EventOutbound         = types.EventOutbound
)
