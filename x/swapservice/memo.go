package swapservice

import (
	"fmt"
	"strconv"
	"strings"
)

// TXTYPE:STATE1:STATE2:STATE3:FINALMEMO

type txType uint8
type adminType uint8

const (
	txCreate txType = iota
	txStake
	txWithdraw
	txSwap
	txAdmin
	unknowTx
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
}

var txToStringMap = map[txType]string{
	txCreate:   "create",
	txStake:    "stake",
	txWithdraw: "withdraw",
	txSwap:     "swap",
	txAdmin:    "admin",
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
	return unknowTx, fmt.Errorf("Invalid tx type: %s", s)
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

	GetSymbol() string
	GetAmount() string
	GetDestination() string
	GetSlipLimit() float64
	GetMemo() string
	GetAdminType() adminType
	GetKey() string
	GetValue() string
}

type MemoBase struct {
	TxType txType
	Symbol string
}

type CreateMemo struct {
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
	Destination string
	SlipLimit   float64
	Memo        string
}

type AdminMemo struct {
	MemoBase
	Key   string
	Value string
	Type  adminType
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

	symbol := strings.ToUpper(parts[1])
	if tx != txAdmin {
		if err := validateSymbol(symbol); err != nil {
			return noMemo, err
		}
	}

	switch tx {
	case txCreate:
		return CreateMemo{
			MemoBase: MemoBase{TxType: txCreate, Symbol: symbol},
		}, nil
	case txStake:
		return StakeMemo{
			MemoBase: MemoBase{TxType: txStake, Symbol: symbol},
		}, nil
	case txWithdraw:
		if len(parts) < 3 {
			return noMemo, fmt.Errorf("Missing withdrawal unit amount")
		}
		// check that amount is parse-able as float64
		_, err := strconv.ParseFloat(parts[2], 64)
		return WithdrawMemo{
			MemoBase: MemoBase{TxType: txWithdraw, Symbol: symbol},
			Amount:   parts[2],
		}, err
	case txSwap:
		max := 5
		parts = strings.SplitN(memo, ":", max)
		if len(parts) < 3 {
			return noMemo, fmt.Errorf("Missing swap parameters: destination address")
		}
		destination := parts[2]
		if err := validateDestination(destination); err != nil {
			return noMemo, fmt.Errorf("Destination address is not valid")
		}
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
			MemoBase:    MemoBase{TxType: txSwap, Symbol: symbol},
			Destination: destination,
			SlipLimit:   slip,
			Memo:        mem,
		}, err
	case txAdmin:
		if len(parts) < 4 {
			return noMemo, fmt.Errorf("Not enough parameters")
		}
		a, err := stringToAdminType(parts[1])
		return AdminMemo{
			Type:  a,
			Key:   parts[2],
			Value: parts[3],
		}, err
	default:
		return noMemo, fmt.Errorf("TxType not supported: %s", tx.String())
	}
}

// Base Functions
func (m MemoBase) GetType() txType         { return m.TxType }
func (m MemoBase) IsType(tx txType) bool   { return m.TxType.Equals(tx) }
func (m MemoBase) GetSymbol() string       { return strings.ToUpper(m.Symbol) }
func (m MemoBase) GetAmount() string       { return "" }
func (m MemoBase) GetDestination() string  { return "" }
func (m MemoBase) GetSlipLimit() float64   { return 0 }
func (m MemoBase) GetMemo() string         { return "" }
func (m MemoBase) GetAdminType() adminType { return adminUnknown }
func (m MemoBase) GetKey() string          { return "" }
func (m MemoBase) GetValue() string        { return "" }

// Transaction Specific Functions
func (m WithdrawMemo) GetAmount() string    { return m.Amount }
func (m SwapMemo) GetDestination() string   { return m.Destination }
func (m SwapMemo) GetSlipLimit() float64    { return m.SlipLimit }
func (m SwapMemo) GetMemo() string          { return m.Memo }
func (m AdminMemo) GetAdminType() adminType { return m.Type }
func (m AdminMemo) GetKey() string          { return m.Key }
func (m AdminMemo) GetValue() string        { return m.Value }

// validates the given symbol
func validateSymbol(sym string) error {
	if len(sym) < 3 {
		return fmt.Errorf("Symbol Error: Not enough characters (%d)", len(sym))
	}

	if len(sym) > 8 {
		return fmt.Errorf("Symbol Error: Too many characters (%d)", len(sym))
	}

	return nil
}

// validates the given binance address
func validateDestination(des string) error {
	// bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6
	prefixes := []string{"bnb", "tbnb"}

	// check if our address has one of the prefixes above
	hasPrefix := false
	for _, pref := range prefixes {
		if strings.HasPrefix(des, pref) {
			hasPrefix = true
			break
		}
	}
	if !hasPrefix {
		return fmt.Errorf("Address prefix is not supported")
	}

	// trim the prefix from our address
	for _, pref := range prefixes {
		if strings.HasPrefix(des, pref) {
			des = strings.TrimLeft(des, pref)
			break
		}
	}

	// check address length is valid
	if len(des) != 39 {
		return fmt.Errorf("Address length is not correct")
	}

	return nil
}
