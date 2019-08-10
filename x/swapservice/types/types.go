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
	return strings.EqualFold(string(t), string(t2))
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
