package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName
)

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
