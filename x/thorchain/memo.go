package thorchain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// TXTYPE:STATE1:STATE2:STATE3:FINALMEMO

type (
	TxType    uint8
	adminType uint8
)

const (
	TxUnknown TxType = iota
	TxStake
	TxUnstake
	TxSwap
	TxOutbound
	TxAdd
	TxBond
	TxLeave
	TxYggdrasilFund
	TxYggdrasilReturn
	TxReserve
	TxRefund
	TxMigrate
	TxRagnarok
	TxSwitch
)

var stringToTxTypeMap = map[string]TxType{
	"stake":      TxStake,
	"st":         TxStake,
	"+":          TxStake,
	"withdraw":   TxUnstake,
	"unstake":    TxUnstake,
	"wd":         TxUnstake,
	"-":          TxUnstake,
	"swap":       TxSwap,
	"s":          TxSwap,
	"=":          TxSwap,
	"outbound":   TxOutbound,
	"add":        TxAdd,
	"a":          TxAdd,
	"%":          TxAdd,
	"bond":       TxBond,
	"leave":      TxLeave,
	"yggdrasil+": TxYggdrasilFund,
	"yggdrasil-": TxYggdrasilReturn,
	"reserve":    TxReserve,
	"refund":     TxRefund,
	"migrate":    TxMigrate,
	"ragnarok":   TxRagnarok,
	"switch":     TxSwitch,
}

var txToStringMap = map[TxType]string{
	TxStake:           "stake",
	TxUnstake:         "unstake",
	TxSwap:            "swap",
	TxOutbound:        "outbound",
	TxRefund:          "refund",
	TxAdd:             "add",
	TxBond:            "bond",
	TxLeave:           "leave",
	TxYggdrasilFund:   "yggdrasil+",
	TxYggdrasilReturn: "yggdrasil-",
	TxReserve:         "reserve",
	TxMigrate:         "migrate",
	TxRagnarok:        "ragnarok",
	TxSwitch:          "switch",
}

// converts a string into a txType
func StringToTxType(s string) (TxType, error) {
	// THORNode can support Abbreviated MEMOs , usually it is only one character
	sl := strings.ToLower(s)
	if t, ok := stringToTxTypeMap[sl]; ok {
		return t, nil
	}
	return TxUnknown, fmt.Errorf("invalid tx type: %s", s)
}

func (tx TxType) IsInbound() bool {
	switch tx {
	case TxStake, TxUnstake, TxSwap, TxAdd, TxBond, TxLeave, TxSwitch, TxReserve:
		return true
	default:
		return false
	}
}

func (tx TxType) IsOutbound() bool {
	switch tx {
	case TxOutbound, TxRefund:
		return true
	default:
		return false
	}
}

func (tx TxType) IsInternal() bool {
	switch tx {
	case TxYggdrasilFund, TxYggdrasilReturn, TxMigrate, TxRagnarok:
		return true
	default:
		return false
	}
}

func (tx TxType) IsEmpty() bool {
	return tx == TxUnknown
}

// Check if two txTypes are the same
func (tx TxType) Equals(tx2 TxType) bool {
	return tx.String() == tx2.String()
}

// Converts a txType into a string
func (tx TxType) String() string {
	return txToStringMap[tx]
}

type Memo interface {
	IsType(tx TxType) bool
	GetType() TxType
	IsEmpty() bool
	IsInbound() bool
	IsOutbound() bool
	IsInternal() bool

	String() string
	GetAsset() common.Asset
	GetAmount() string
	GetDestination() common.Address
	GetSlipLimit() sdk.Uint
	GetKey() string
	GetValue() string
	GetTxID() common.TxID
	GetAccAddress() sdk.AccAddress
	GetBlockHeight() int64
}

type MemoBase struct {
	TxType TxType
	Asset  common.Asset
}

type CreateMemo struct {
	MemoBase
}

type GasMemo struct {
	MemoBase
}

type AddMemo struct {
	MemoBase
}

type StakeMemo struct {
	MemoBase
	RuneAmount  string
	AssetAmount string
	Address     common.Address
}

type UnstakeMemo struct {
	MemoBase
	Amount string
}

type SwapMemo struct {
	MemoBase
	Destination common.Address
	SlipLimit   sdk.Uint
}

type AdminMemo struct {
	MemoBase
	Key   string
	Value string
	Type  adminType
}

type OutboundMemo struct {
	MemoBase
	TxID common.TxID
}

type RefundMemo struct {
	MemoBase
	TxID common.TxID
}

type BondMemo struct {
	MemoBase
	NodeAddress sdk.AccAddress
}

type LeaveMemo struct {
	MemoBase
}

type YggdrasilFundMemo struct {
	MemoBase
	BlockHeight int64
}

type YggdrasilReturnMemo struct {
	MemoBase
	BlockHeight int64
}

type ReserveMemo struct {
	MemoBase
}

type MigrateMemo struct {
	MemoBase
	BlockHeight int64
}

type RagnarokMemo struct {
	MemoBase
	BlockHeight int64
}

type SwitchMemo struct {
	MemoBase
	Destination common.Address
}

func NewSwitchMemo(addr common.Address) SwitchMemo {
	return SwitchMemo{
		MemoBase:    MemoBase{TxType: TxSwitch},
		Destination: addr,
	}
}

func NewLeaveMemo() LeaveMemo {
	return LeaveMemo{
		MemoBase: MemoBase{TxType: TxLeave},
	}
}

func NewAddMemo(asset common.Asset) AddMemo {
	return AddMemo{
		MemoBase: MemoBase{TxType: TxAdd, Asset: asset},
	}
}

func NewRagnarokMemo(blockHeight int64) RagnarokMemo {
	return RagnarokMemo{
		MemoBase:    MemoBase{TxType: TxRagnarok},
		BlockHeight: blockHeight,
	}
}

func NewStakeMemo(asset common.Asset, addr common.Address) StakeMemo {
	return StakeMemo{
		MemoBase: MemoBase{TxType: TxStake, Asset: asset},
		Address:  addr,
	}
}

func NewUnstakeMemo(asset common.Asset, amt string) UnstakeMemo {
	return UnstakeMemo{
		MemoBase: MemoBase{TxType: TxUnstake, Asset: asset},
		Amount:   amt,
	}
}

func NewReserveMemo() ReserveMemo {
	return ReserveMemo{
		MemoBase: MemoBase{TxType: TxReserve},
	}
}

func NewMigrateMemo(blockHeight int64) MigrateMemo {
	return MigrateMemo{
		MemoBase:    MemoBase{TxType: TxMigrate},
		BlockHeight: blockHeight,
	}
}

func NewYggdrasilFund(blockHeight int64) YggdrasilFundMemo {
	return YggdrasilFundMemo{
		MemoBase:    MemoBase{TxType: TxYggdrasilFund},
		BlockHeight: blockHeight,
	}
}

func NewYggdrasilReturn(blockHeight int64) YggdrasilReturnMemo {
	return YggdrasilReturnMemo{
		MemoBase:    MemoBase{TxType: TxYggdrasilReturn},
		BlockHeight: blockHeight,
	}
}

func NewOutboundMemo(txID common.TxID) OutboundMemo {
	return OutboundMemo{
		MemoBase: MemoBase{TxType: TxOutbound},
		TxID:     txID,
	}
}

// NewRefundMemo create a new RefundMemo
func NewRefundMemo(txID common.TxID) RefundMemo {
	return RefundMemo{
		MemoBase: MemoBase{TxType: TxRefund},
		TxID:     txID,
	}
}

func NewBondMemo(addr sdk.AccAddress) BondMemo {
	return BondMemo{
		MemoBase:    MemoBase{TxType: TxBond},
		NodeAddress: addr,
	}
}

func NewSwapMemo(asset common.Asset, dest common.Address, slip sdk.Uint) SwapMemo {
	return SwapMemo{
		MemoBase:    MemoBase{TxType: TxSwap, Asset: asset},
		Destination: dest,
		SlipLimit:   slip,
	}
}

func ParseMemo(memo string) (Memo, error) {
	var err error
	noMemo := MemoBase{}
	if len(memo) == 0 {
		return noMemo, fmt.Errorf("memo can't be empty")
	}
	parts := strings.Split(memo, ":")
	tx, err := StringToTxType(parts[0])
	if err != nil {
		return noMemo, err
	}

	// list of memo types that do not contain an asset in their memo
	noAssetMemos := []TxType{
		TxOutbound, TxBond, TxLeave, TxRefund,
		TxYggdrasilFund, TxYggdrasilReturn, TxReserve,
		TxMigrate, TxRagnarok, TxSwitch,
	}
	hasAsset := true
	for _, memoType := range noAssetMemos {
		if tx == memoType {
			hasAsset = false
		}
	}

	var asset common.Asset
	if hasAsset {
		if len(parts) < 2 {
			return noMemo, fmt.Errorf("cannot parse given memo: length %d", len(parts))
		}
		var err error
		asset, err = common.NewAsset(parts[1])
		if err != nil {
			return noMemo, err
		}
	}

	switch tx {
	case TxLeave:
		return NewLeaveMemo(), nil
	case TxAdd:
		return NewAddMemo(asset), nil
	case TxStake:
		var addr common.Address
		if !asset.Chain.IsBNB() {
			if len(parts) < 3 {
				// cannot stake into a non BNB-based pool when THORNode don't have an
				// associated address
				return noMemo, fmt.Errorf("invalid stake. Cannot stake to a non BNB-based pool without providing an associated address")
			}
			addr, err = common.NewAddress(parts[2])
			if err != nil {
				return noMemo, err
			}
		}
		return NewStakeMemo(asset, addr), nil

	case TxUnstake:
		if len(parts) < 2 {
			return noMemo, fmt.Errorf("invalid unstake memo")
		}
		var withdrawAmount string
		if len(parts) > 2 {
			withdrawAmount = parts[2]
			wa, err := sdk.ParseUint(withdrawAmount)
			if err != nil {
				return noMemo, err
			}
			if !wa.GT(sdk.ZeroUint()) || wa.GT(sdk.NewUint(MaxUnstakeBasisPoints)) {
				return noMemo, fmt.Errorf("withdraw amount :%s is invalid", withdrawAmount)
			}
		}
		return NewUnstakeMemo(asset, withdrawAmount), nil

	case TxSwap:
		if len(parts) < 2 {
			return noMemo, fmt.Errorf("missing swap parameters: memo should in SWAP:SYMBOLXX-XXX:DESTADDR:TRADE-TARGET format")
		}
		// DESTADDR can be empty , if it is empty , it will swap to the sender address
		destination := common.NoAddress
		if len(parts) > 2 {
			if len(parts[2]) > 0 {
				destination, err = common.NewAddress(parts[2])
				if err != nil {
					return noMemo, err
				}
			}
		}
		// price limit can be empty , when it is empty , there is no price protection
		slip := sdk.ZeroUint()
		if len(parts) > 3 && len(parts[3]) > 0 {
			amount, err := sdk.ParseUint(parts[3])
			if err != nil {
				return noMemo, fmt.Errorf("swap price limit:%s is invalid", parts[3])
			}

			slip = amount
		}
		return NewSwapMemo(asset, destination, slip), nil
	case TxOutbound:
		if len(parts) < 2 {
			return noMemo, fmt.Errorf("not enough parameters")
		}
		txID, err := common.NewTxID(parts[1])
		return NewOutboundMemo(txID), err
	case TxRefund:
		if len(parts) < 2 {
			return noMemo, fmt.Errorf("not enough parameters")
		}
		txID, err := common.NewTxID(parts[1])
		return NewRefundMemo(txID), err
	case TxBond:
		if len(parts) < 2 {
			return noMemo, fmt.Errorf("not enough parameters")
		}
		addr, err := sdk.AccAddressFromBech32(parts[1])
		if err != nil {
			return noMemo, fmt.Errorf("%s is an invalid thorchain address: %w", parts[1], err)
		}
		return NewBondMemo(addr), nil
	case TxYggdrasilFund:
		if len(parts) < 2 {
			return noMemo, errors.New("not enough parameters")
		}
		blockHeight, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return noMemo, fmt.Errorf("fail to convert (%s) to a valid block height: %w", parts[1], err)
		}
		return NewYggdrasilFund(blockHeight), nil
	case TxYggdrasilReturn:
		if len(parts) < 2 {
			return noMemo, errors.New("not enough parameters")
		}
		blockHeight, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return noMemo, fmt.Errorf("fail to convert (%s) to a valid block height: %w", parts[1], err)
		}
		return NewYggdrasilReturn(blockHeight), nil
	case TxReserve:
		return NewReserveMemo(), nil
	case TxMigrate:
		if len(parts) < 2 {
			return noMemo, errors.New("not enough parameters")
		}
		blockHeight, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return noMemo, fmt.Errorf("fail to convert (%s) to a valid block height: %w", parts[1], err)
		}
		return NewMigrateMemo(blockHeight), nil
	case TxRagnarok:
		if len(parts) < 2 {
			return noMemo, errors.New("not enough parameters")
		}
		blockHeight, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return noMemo, fmt.Errorf("fail to convert (%s) to a valid block height: %w", parts[1], err)
		}
		return NewRagnarokMemo(blockHeight), nil
	case TxSwitch:
		if len(parts) < 2 {
			return noMemo, errors.New("not enough parameters")
		}
		destination, err := common.NewAddress(parts[1])
		if err != nil {
			return noMemo, err
		}
		if destination.IsEmpty() {
			return noMemo, errors.New("address cannot be empty")
		}
		return NewSwitchMemo(destination), nil
	default:
		return noMemo, fmt.Errorf("TxType not supported: %s", tx.String())
	}
}

// Base Functions
func (m MemoBase) String() string                 { return "" }
func (m MemoBase) GetType() TxType                { return m.TxType }
func (m MemoBase) IsType(tx TxType) bool          { return m.TxType.Equals(tx) }
func (m MemoBase) GetAsset() common.Asset         { return m.Asset }
func (m MemoBase) GetAmount() string              { return "" }
func (m MemoBase) GetDestination() common.Address { return "" }
func (m MemoBase) GetSlipLimit() sdk.Uint         { return sdk.ZeroUint() }
func (m MemoBase) GetKey() string                 { return "" }
func (m MemoBase) GetValue() string               { return "" }
func (m MemoBase) GetTxID() common.TxID           { return "" }
func (m MemoBase) GetAccAddress() sdk.AccAddress  { return sdk.AccAddress{} }
func (m MemoBase) GetBlockHeight() int64          { return 0 }
func (m MemoBase) IsOutbound() bool               { return m.TxType.IsOutbound() }
func (m MemoBase) IsInbound() bool                { return m.TxType.IsInbound() }
func (m MemoBase) IsInternal() bool               { return m.TxType.IsInternal() }
func (m MemoBase) IsEmpty() bool                  { return m.TxType.IsEmpty() }

// Transaction Specific Functions
func (m UnstakeMemo) GetAmount() string            { return m.Amount }
func (m SwapMemo) GetDestination() common.Address  { return m.Destination }
func (m SwapMemo) GetSlipLimit() sdk.Uint          { return m.SlipLimit }
func (m AdminMemo) GetKey() string                 { return m.Key }
func (m AdminMemo) GetValue() string               { return m.Value }
func (m BondMemo) GetAccAddress() sdk.AccAddress   { return m.NodeAddress }
func (m StakeMemo) GetDestination() common.Address { return m.Address }
func (m OutboundMemo) GetTxID() common.TxID        { return m.TxID }
func (m OutboundMemo) String() string {
	return fmt.Sprintf("OUTBOUND:%s", m.TxID.String())
}

// GetTxID return the relevant tx id in refund memo
func (m RefundMemo) GetTxID() common.TxID { return m.TxID }

// String implement fmt.Stringer
func (m RefundMemo) String() string {
	return fmt.Sprintf("REFUND:%s", m.TxID.String())
}

func (m YggdrasilFundMemo) String() string {
	return fmt.Sprintf("YGGDRASIL+:%d", m.BlockHeight)
}

func (m YggdrasilFundMemo) GetBlockHeight() int64 {
	return m.BlockHeight
}

func (m YggdrasilReturnMemo) String() string {
	return fmt.Sprintf("YGGDRASIL-:%d", m.BlockHeight)
}

func (m YggdrasilReturnMemo) GetBlockHeight() int64 {
	return m.BlockHeight
}

func (m MigrateMemo) String() string {
	return fmt.Sprintf("MIGRATE:%d", m.BlockHeight)
}

func (m MigrateMemo) GetBlockHeight() int64 {
	return m.BlockHeight
}

func (m RagnarokMemo) String() string {
	return fmt.Sprintf("RAGNAROK:%d", m.BlockHeight)
}

func (m RagnarokMemo) GetBlockHeight() int64 {
	return m.BlockHeight
}

func (m SwitchMemo) GetDestination() common.Address {
	return m.Destination
}
