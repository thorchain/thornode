package swapservice

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/thorchain/statechain/x/swapservice/types"
)

// TXTYPE:STATE1:STATE2:STATE3:FINALMEMO

type txType uint8
type adminType uint8

const (
	txUnknown txType = iota
	txCreate
	txStake
	txWithdraw
	txSwap
	txAdmin
	txOutbound
	txDonate
	txGas
)

const (
	adminUnknown adminType = iota
	adminKey
	adminPoolStatus
)

var stringToTxTypeMap = map[string]txType{
	"create":   txCreate,
	"stake":    txStake,
	"withdraw": txWithdraw,
	"swap":     txSwap,
	"admin":    txAdmin,
	"outbound": txOutbound,
	"donate":   txDonate,
	"gas":      txGas,
}

var txToStringMap = map[txType]string{
	txCreate:   "create",
	txStake:    "stake",
	txWithdraw: "withdraw",
	txSwap:     "swap",
	txAdmin:    "admin",
	txOutbound: "outbound",
	txDonate:   "donate",
	txGas:      "gas",
}

var stringToAdminTypeMap = map[string]adminType{
	"key":        adminKey,
	"poolstatus": adminPoolStatus,
}

// converts a string into a txType
func stringToTxType(s string) (txType, error) {
	sl := strings.ToLower(s)
	if t, ok := stringToTxTypeMap[sl]; ok {
		return t, nil
	}
	return txUnknown, fmt.Errorf("Invalid tx type: %s", s)
}

// converts a string into a adminType
func stringToAdminType(s string) (adminType, error) {
	sl := strings.ToLower(s)
	if t, ok := stringToAdminTypeMap[sl]; ok {
		return t, nil
	}
	return adminUnknown, fmt.Errorf("Invalid admin type: %s", s)
}

// Check if two txTypes are the same
func (tx txType) Equals(tx2 txType) bool {
	return tx.String() == tx2.String()
}

// Converts a txType into a string
func (tx txType) String() string {
	return txToStringMap[tx]
}

type Memo interface {
	IsType(tx txType) bool

	GetTicker() Ticker
	GetAmount() string
	GetDestination() BnbAddress
	GetSlipLimit() float64
	GetMemo() string
	GetAdminType() adminType
	GetKey() string
	GetValue() string
	GetBlockHeight() int64
}

type MemoBase struct {
	TxType txType
	Ticker Ticker
}

type CreateMemo struct {
	MemoBase
}

type GasMemo struct {
	MemoBase
}

type DonateMemo struct {
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
	Destination BnbAddress
	SlipLimit   float64
	Memo        string
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
	tx, err := stringToTxType(strings.ToLower(parts[0]))
	if err != nil {
		return noMemo, err
	}

	var ticker Ticker
	if tx != txGas && tx != txAdmin && tx != txOutbound {
		var err error
		ticker, err = NewTicker(parts[1])
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
			MemoBase: MemoBase{TxType: txCreate, Ticker: ticker},
		}, nil

	case txDonate:
		return DonateMemo{
			MemoBase: MemoBase{TxType: txDonate, Ticker: ticker},
		}, nil

	case txStake:
		return StakeMemo{
			MemoBase: MemoBase{TxType: txStake, Ticker: ticker},
		}, nil

	case txWithdraw:
		if len(parts) < 3 {
			return noMemo, fmt.Errorf("Missing withdrawal unit amount")
		}
		// check that amount is parse-able as float64
		_, err := strconv.ParseFloat(parts[2], 64)
		return WithdrawMemo{
			MemoBase: MemoBase{TxType: txWithdraw, Ticker: ticker},
			Amount:   parts[2],
		}, err

	case txSwap:
		max := 5
		parts = strings.SplitN(memo, ":", max)
		if len(parts) < 2 {
			return noMemo, fmt.Errorf("missing swap parameters: memo should in SWAP:SYMBOLXX-XXX:DESTADDR:TRADE-TARGET format")
		}
		// DESTADDR can be empty , if it is empty , it will swap to the sender address
		destination := types.NoBnbAddress
		if len(parts) > 2 {
			if len(parts[2]) > 0 {
				destination, err = NewBnbAddress(parts[2])
				if err != nil {
					return noMemo, err
				}
			}
		}
		// trade target can be empty , when it is empty , there is no price protection
		var slip float64
		if len(parts) > 3 && len(parts[3]) > 0 {
			slip, err = strconv.ParseFloat(parts[3], 64)
			if err != nil {
				return noMemo, err
			}
		}
		var mem string
		if len(parts) == max {
			mem = parts[4]
		}
		return SwapMemo{
			MemoBase:    MemoBase{TxType: txSwap, Ticker: ticker},
			Destination: destination,
			SlipLimit:   slip,
			Memo:        mem,
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
func (m MemoBase) GetType() txType            { return m.TxType }
func (m MemoBase) IsType(tx txType) bool      { return m.TxType.Equals(tx) }
func (m MemoBase) GetTicker() Ticker          { return m.Ticker }
func (m MemoBase) GetAmount() string          { return "" }
func (m MemoBase) GetDestination() BnbAddress { return "" }
func (m MemoBase) GetSlipLimit() float64      { return 0 }
func (m MemoBase) GetMemo() string            { return "" }
func (m MemoBase) GetAdminType() adminType    { return adminUnknown }
func (m MemoBase) GetKey() string             { return "" }
func (m MemoBase) GetValue() string           { return "" }
func (m MemoBase) GetBlockHeight() int64      { return 0 }

// Transaction Specific Functions
func (m WithdrawMemo) GetAmount() string      { return m.Amount }
func (m SwapMemo) GetDestination() BnbAddress { return m.Destination }
func (m SwapMemo) GetSlipLimit() float64      { return m.SlipLimit }
func (m SwapMemo) GetMemo() string            { return m.Memo }
func (m AdminMemo) GetAdminType() adminType   { return m.Type }
func (m AdminMemo) GetKey() string            { return m.Key }
func (m AdminMemo) GetValue() string          { return m.Value }
func (m OutboundMemo) GetBlockHeight() int64  { return m.BlockHeight }
