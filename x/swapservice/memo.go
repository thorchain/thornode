package swapservice

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

// TXTYPE:STATE1:STATE2:STATE3:FINALMEMO

type TxType uint8
type adminType uint8

const (
	txUnknown TxType = iota
	txCreate
	txStake
	txWithdraw
	txSwap
	txOutbound
	txAdd
	txGas
	txBond
	txNextPool
	txLeave
	txAck
)

var stringToTxTypeMap = map[string]TxType{
	"create":   txCreate,
	"c":        txCreate,
	"#":        txCreate,
	"stake":    txStake,
	"st":       txStake,
	"+":        txStake,
	"withdraw": txWithdraw,
	"wd":       txWithdraw,
	"-":        txWithdraw,
	"swap":     txSwap,
	"s":        txSwap,
	"=":        txSwap,
	"outbound": txOutbound,
	"add":      txAdd,
	"a":        txAdd,
	"%":        txAdd,
	"gas":      txGas,
	"g":        txGas,
	"$":        txGas,
	"bond":     txBond,
	"nextpool": txNextPool,
	"leave":    txLeave,
	"ack":      txAck,
}

var txToStringMap = map[TxType]string{
	txCreate:   "create",
	txStake:    "stake",
	txWithdraw: "withdraw",
	txSwap:     "swap",
	txOutbound: "outbound",
	txAdd:      "add",
	txGas:      "gas",
	txBond:     "bond",
	txNextPool: "nextpool",
	txLeave:    "leave",
	txAck:      "ack",
}

// converts a string into a txType
func stringToTxType(s string) (TxType, error) {
	// we can support Abbreviated MEMOs , usually it is only one character
	sl := strings.ToLower(s)
	if t, ok := stringToTxTypeMap[sl]; ok {
		return t, nil
	}
	return txUnknown, fmt.Errorf("invalid tx type: %s", s)
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

	GetAsset() common.Asset
	GetAmount() string
	GetDestination() common.Address
	GetSlipLimit() sdk.Uint
	GetKey() string
	GetValue() string
	GetBlockHeight() uint64
	GetNodeAddress() sdk.AccAddress
	GetNextPoolAddress() common.Address
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
	TokenAmount string
}

type WithdrawMemo struct {
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
	BlockHeight uint64
}

type BondMemo struct {
	MemoBase
	NodeAddress sdk.AccAddress
}

type NextPoolMemo struct {
	MemoBase
	NextPoolAddr common.Address
}

type LeaveMemo struct {
	MemoBase
}
type AckMemo struct {
	MemoBase
}

func ParseMemo(memo string) (Memo, error) {
	var err error
	noMemo := MemoBase{}
	if len(memo) == 0 {
		return noMemo, fmt.Errorf("memo can't be empty")
	}
	parts := strings.Split(memo, ":")
	tx, err := stringToTxType(parts[0])
	if err != nil {
		return noMemo, err
	}

	var asset common.Asset
	if tx != txGas && tx != txOutbound && tx != txBond && tx != txLeave && tx != txAck && tx != txNextPool {
		if len(parts) < 2 {
			return noMemo, fmt.Errorf("cannot parse given memo: length %d", len(parts))
		}
		var err error
		fmt.Printf("Parts: %+v\n", parts)
		asset, err = common.NewAsset(parts[1])
		if err != nil {
			return noMemo, err
		}
	}

	switch tx {
	case txCreate:
		return CreateMemo{
			MemoBase: MemoBase{TxType: txCreate, Asset: asset},
		}, nil

	case txGas:
		return GasMemo{
			MemoBase: MemoBase{TxType: txGas},
		}, nil
	case txLeave:
		return LeaveMemo{
			MemoBase: MemoBase{TxType: txLeave},
		}, nil
	case txAck:
		return AckMemo{
			MemoBase: MemoBase{
				TxType: txAck,
			},
		}, nil
	case txAdd:
		return AddMemo{
			MemoBase: MemoBase{TxType: txAdd, Asset: asset},
		}, nil

	case txStake:
		return StakeMemo{
			MemoBase: MemoBase{TxType: txStake, Asset: asset},
		}, nil

	case txWithdraw:
		if len(parts) < 2 {
			return noMemo, fmt.Errorf("invalid unstake memo")
		}
		var withdrawAmount string
		if len(parts) > 2 {
			withdrawAmount = parts[2]
			wa, err := strconv.ParseFloat(withdrawAmount, 10)
			if nil != err || wa < 0 || wa > MaxWithdrawBasisPoints {
				return noMemo, fmt.Errorf("withdraw amount :%s is invalid", withdrawAmount)
			}
		}
		return WithdrawMemo{
			MemoBase: MemoBase{TxType: txWithdraw, Asset: asset},
			Amount:   withdrawAmount,
		}, err

	case txSwap:
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
			if nil != err {
				return noMemo, fmt.Errorf("swap price limit:%s is invalid", parts[3])
			}

			slip = amount
		}
		return SwapMemo{
			MemoBase:    MemoBase{TxType: txSwap, Asset: asset},
			Destination: destination,
			SlipLimit:   slip,
		}, err

	case txOutbound:
		if len(parts) < 2 {
			return noMemo, fmt.Errorf("not enough parameters")
		}
		height, err := strconv.ParseUint(parts[1], 0, 64)

		return OutboundMemo{
			BlockHeight: height,
		}, err
	case txBond:
		if len(parts) < 2 {
			return noMemo, fmt.Errorf("not enough parameters")
		}
		addr, err := sdk.AccAddressFromBech32(parts[1])
		if nil != err {
			return noMemo, errors.Wrapf(err, "%s is an invalid bep address", parts[1])
		}
		return BondMemo{
			MemoBase:    MemoBase{TxType: txBond},
			NodeAddress: addr,
		}, nil

	case txNextPool:
		nextPoolAddr, err := common.NewAddress(parts[1])
		if nil != err {
			return noMemo, fmt.Errorf("%s is an invalid bnb address,err: %w", parts[1], err)
		}
		return NextPoolMemo{
			MemoBase: MemoBase{
				TxType: txNextPool,
				Asset:  common.Asset{},
			},
			NextPoolAddr: nextPoolAddr,
		}, nil
	default:
		return noMemo, fmt.Errorf("TxType not supported: %s", tx.String())
	}
}

// Base Functions
func (m MemoBase) GetType() TxType                    { return m.TxType }
func (m MemoBase) IsType(tx TxType) bool              { return m.TxType.Equals(tx) }
func (m MemoBase) GetAsset() common.Asset             { return m.Asset }
func (m MemoBase) GetAmount() string                  { return "" }
func (m MemoBase) GetDestination() common.Address     { return "" }
func (m MemoBase) GetSlipLimit() sdk.Uint             { return sdk.ZeroUint() }
func (m MemoBase) GetKey() string                     { return "" }
func (m MemoBase) GetValue() string                   { return "" }
func (m MemoBase) GetBlockHeight() uint64             { return 0 }
func (m MemoBase) GetNodeAddress() sdk.AccAddress     { return sdk.AccAddress{} }
func (m MemoBase) GetNextPoolAddress() common.Address { return "" }

// Transaction Specific Functions
func (m WithdrawMemo) GetAmount() string                  { return m.Amount }
func (m SwapMemo) GetDestination() common.Address         { return m.Destination }
func (m SwapMemo) GetSlipLimit() sdk.Uint                 { return m.SlipLimit }
func (m AdminMemo) GetKey() string                        { return m.Key }
func (m AdminMemo) GetValue() string                      { return m.Value }
func (m OutboundMemo) GetBlockHeight() uint64             { return m.BlockHeight }
func (m BondMemo) GetNodeAddress() sdk.AccAddress         { return m.NodeAddress }
func (m NextPoolMemo) GetNextPoolAddress() common.Address { return m.NextPoolAddr }
