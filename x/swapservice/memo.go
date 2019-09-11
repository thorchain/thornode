package swapservice

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/bepswap/common"
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
	txAdmin
	txOutbound
	txAdd
	txGas
)

const (
	adminUnknown adminType = iota
	adminKey
	adminPoolStatus
)

var stringToTxTypeMap = map[string]TxType{
	"create":   txCreate,
	"stake":    txStake,
	"withdraw": txWithdraw,
	"swap":     txSwap,
	"admin":    txAdmin,
	"outbound": txOutbound,
	"add":      txAdd,
	"gas":      txGas,
}

// Swap: >:
// Stake: +:
// Withdraw: -:
// Create: &:
// Admin: !:
// Gas: $:
// Add: %:
// symbolToTxTypeMap a map from symbol to txType
// https://gitlab.com/thorchain/bepswap/statechain/issues/64
var symbolToTxTypeMap = map[string]TxType{
	"&": txCreate,
	"+": txStake,
	"-": txWithdraw,
	">": txSwap,
	"!": txAdmin,
	"$": txGas,
	"%": txAdd,
}

var txToStringMap = map[TxType]string{
	txCreate:   "create",
	txStake:    "stake",
	txWithdraw: "withdraw",
	txSwap:     "swap",
	txAdmin:    "admin",
	txOutbound: "outbound",
	txAdd:      "add",
	txGas:      "gas",
}

var stringToAdminTypeMap = map[string]adminType{
	"key":        adminKey,
	"poolstatus": adminPoolStatus,
}

// converts a string into a txType
func stringToTxType(s string) (TxType, error) {
	// we can support Abbreviated MEMOs , usually it is only one character
	if len(s) == 1 {
		if t, ok := symbolToTxTypeMap[s]; ok {
			return t, nil
		}
	}
	sl := strings.ToLower(s)
	if t, ok := stringToTxTypeMap[sl]; ok {
		return t, nil
	}
	return txUnknown, fmt.Errorf("invalid tx type: %s", s)
}

// converts a string into a adminType
func stringToAdminType(s string) (adminType, error) {
	sl := strings.ToLower(s)
	if t, ok := stringToAdminTypeMap[sl]; ok {
		return t, nil
	}
	return adminUnknown, fmt.Errorf("invalid admin type: %s", s)
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

	GetTicker() common.Ticker
	GetAmount() string
	GetDestination() common.BnbAddress
	GetSlipLimit() sdk.Uint
	GetAdminType() adminType
	GetKey() string
	GetValue() string
	GetBlockHeight() int64
}

type MemoBase struct {
	TxType TxType
	Ticker common.Ticker
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
	Destination common.BnbAddress
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
	BlockHeight int64
}

func ParseMemo(memo string) (Memo, error) {
	var err error
	noMemo := MemoBase{}
	parts := strings.Split(memo, ":")
	if len(parts) < 2 {
		return noMemo, fmt.Errorf("Cannot parse given memo: length %d", len(parts))
	}
	tx, err := stringToTxType(parts[0])
	if err != nil {
		return noMemo, err
	}

	var ticker common.Ticker
	if tx != txGas && tx != txAdmin && tx != txOutbound {
		var err error
		ticker, err = common.NewTicker(parts[1])
		if err != nil {
			return noMemo, err
		}
	}

	switch tx {
	case txCreate:
		return CreateMemo{
			MemoBase: MemoBase{TxType: txCreate, Ticker: ticker},
		}, nil

	case txGas:
		return GasMemo{
			MemoBase: MemoBase{TxType: txGas, Ticker: ticker},
		}, nil

	case txAdd:
		return AddMemo{
			MemoBase: MemoBase{TxType: txAdd, Ticker: ticker},
		}, nil

	case txStake:
		return StakeMemo{
			MemoBase: MemoBase{TxType: txStake, Ticker: ticker},
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
			MemoBase: MemoBase{TxType: txWithdraw, Ticker: ticker},
			Amount:   withdrawAmount,
		}, err

	case txSwap:
		if len(parts) < 2 {
			return noMemo, fmt.Errorf("missing swap parameters: memo should in SWAP:SYMBOLXX-XXX:DESTADDR:TRADE-TARGET format")
		}
		// DESTADDR can be empty , if it is empty , it will swap to the sender address
		destination := common.NoBnbAddress
		if len(parts) > 2 {
			if len(parts[2]) > 0 {
				destination, err = common.NewBnbAddress(parts[2])
				if err != nil {
					return noMemo, err
				}
			}
		}
		// trade target can be empty , when it is empty , there is no price protection
		slip := sdk.ZeroUint()
		if len(parts) > 3 && len(parts[3]) > 0 {
			amount, err := common.NewAmount(parts[3])
			if nil != err {
				return noMemo, fmt.Errorf("swap price limit:%s is invalid", parts[3])
			}

			slip = sdk.NewUint(uint64(math.Round(amount.Float64() * float64(One))))
		}
		return SwapMemo{
			MemoBase:    MemoBase{TxType: txSwap, Ticker: ticker},
			Destination: destination,
			SlipLimit:   slip,
		}, err

	case txAdmin:
		if len(parts) < 4 {
			return noMemo, fmt.Errorf("not enough parameters")
		}
		a, err := stringToAdminType(parts[1])
		return AdminMemo{
			MemoBase: MemoBase{TxType: txAdmin},
			Type:     a,
			Key:      parts[2],
			Value:    parts[3],
		}, err

	case txOutbound:
		if len(parts) < 2 {
			return noMemo, fmt.Errorf("Not enough parameters")
		}
		height, err := strconv.ParseInt(parts[1], 0, 64)
		return OutboundMemo{
			BlockHeight: height,
		}, err
	default:
		return noMemo, fmt.Errorf("TxType not supported: %s", tx.String())
	}
}

// Base Functions
func (m MemoBase) GetType() TxType                   { return m.TxType }
func (m MemoBase) IsType(tx TxType) bool             { return m.TxType.Equals(tx) }
func (m MemoBase) GetTicker() common.Ticker          { return m.Ticker }
func (m MemoBase) GetAmount() string                 { return "" }
func (m MemoBase) GetDestination() common.BnbAddress { return "" }
func (m MemoBase) GetSlipLimit() sdk.Uint            { return sdk.ZeroUint() }
func (m MemoBase) GetAdminType() adminType           { return adminUnknown }
func (m MemoBase) GetKey() string                    { return "" }
func (m MemoBase) GetValue() string                  { return "" }
func (m MemoBase) GetBlockHeight() int64             { return 0 }

// Transaction Specific Functions
func (m WithdrawMemo) GetAmount() string             { return m.Amount }
func (m SwapMemo) GetDestination() common.BnbAddress { return m.Destination }
func (m SwapMemo) GetSlipLimit() sdk.Uint            { return m.SlipLimit }
func (m AdminMemo) GetAdminType() adminType          { return m.Type }
func (m AdminMemo) GetKey() string                   { return m.Key }
func (m AdminMemo) GetValue() string                 { return m.Value }
func (m OutboundMemo) GetBlockHeight() int64         { return m.BlockHeight }
