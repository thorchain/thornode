package types

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName
)

const floatPrecision = 8

var RuneTicker Ticker = Ticker("RUNE")

type Ticker string

func NewTicker(ticker string) (Ticker, error) {
	noTicker := Ticker("")
	if len(ticker) < 3 {
		return noTicker, fmt.Errorf("Ticker Error: Not enough characters")
	}
	if len(ticker) > 8 {
		return noTicker, fmt.Errorf("Ticker Error: Too many characters")
	}
	return Ticker(strings.ToUpper(ticker)), nil
}

func (t Ticker) Equals(t2 Ticker) bool {
	return strings.EqualFold(t.String(), t2.String())
}

func (t Ticker) Empty() bool {
	return strings.TrimSpace(t.String()) == ""
}

func (t Ticker) String() string {
	// uppercasing again just incase someon created a ticker via Ticker("rune")
	return strings.ToUpper(string(t))
}

func IsRune(ticker Ticker) bool {
	return ticker.Equals(RuneTicker)
}

type TxID string

func NewTxID(hash string) (TxID, error) {
	if len(hash) != 64 {
		return TxID(""), fmt.Errorf("TxID Error: Must be 64 characters (got %d)", len(hash))
	}
	return TxID(strings.ToUpper(hash)), nil
}

func (tx TxID) Equals(tx2 TxID) bool {
	return strings.EqualFold(tx.String(), tx2.String())
}

func (tx TxID) Empty() bool {
	return strings.TrimSpace(tx.String()) == ""
}

func (tx TxID) String() string {
	return string(tx)
}

type Amount string

var ZeroAmount Amount = Amount("0")

func NewAmount(amount string) (Amount, error) {
	_, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return Amount("o"), err
	}
	return Amount(amount), nil
}

func NewAmountFromFloat(f float64) Amount {
	return Amount(strconv.FormatFloat(f, 'f', floatPrecision, 64))
}

func (a Amount) Equals(a2 Amount) bool {
	return strings.EqualFold(a.String(), a2.String())
}

func (a Amount) Empty() bool {
	return strings.TrimSpace(a.String()) == ""
}

func (a Amount) Zero() bool {
	return a.Equals(ZeroAmount)
}

func (a Amount) Float64() float64 {
	amt, _ := strconv.ParseFloat(a.String(), 64)
	return amt
}

func (a Amount) String() string {
	return string(a)
}

type BnbAddress string

var NoBnbAddress BnbAddress = BnbAddress("")

func NewBnbAddress(address string) (BnbAddress, error) {
	// Sample: bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6

	prefixes := []string{"bnb", "tbnb"}

	// check if our address has one of the prefixes above
	hasPrefix := false
	for _, pref := range prefixes {
		if strings.HasPrefix(address, pref) {
			hasPrefix = true
			break
		}
	}
	if !hasPrefix {
		return "", fmt.Errorf("Address prefix is not supported")
	}

	// trim the prefix from our address
	for _, pref := range prefixes {
		if strings.HasPrefix(address, pref) {
			address = strings.TrimLeft(address, pref)
			break
		}
	}

	// check address length is valid
	if len(address) != 39 {
		return "", fmt.Errorf("Address length is not correct")
	}

	return BnbAddress(address), nil
}

func (bnb BnbAddress) Equals(bnb2 BnbAddress) bool {
	return strings.EqualFold(bnb.String(), bnb2.String())
}

func (bnb BnbAddress) Empty() bool {
	return strings.TrimSpace(bnb.String()) == ""
}

func (bnb BnbAddress) String() string {
	return string(bnb)
}
