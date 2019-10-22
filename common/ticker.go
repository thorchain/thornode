package common

import (
	"fmt"
	"strings"
)

const (
	BNBTicker     = Ticker("BNB")
	RuneTicker    = Ticker("RUNE")
	RuneA1FTicker = Ticker("RUNE-A1F")
	RuneB1ATicker = Ticker("RUNE-B1A")
)

type Ticker string
type Tickers []Ticker

func NewTicker(ticker string) (Ticker, error) {
	noTicker := Ticker("")
	if len(ticker) < 3 {
		return noTicker, fmt.Errorf("Ticker Error: Not enough characters")
	}

	if len(ticker) > 13 {
		return noTicker, fmt.Errorf("Ticker Error: Too many characters")
	}
	return Ticker(strings.ToUpper(ticker)), nil
}

func (t Ticker) Equals(t2 Ticker) bool {
	return strings.EqualFold(t.String(), t2.String())
}

func (t Ticker) IsEmpty() bool {
	return strings.TrimSpace(t.String()) == ""
}

func (t Ticker) String() string {
	// uppercasing again just incase someon created a ticker via Ticker("rune")
	return strings.ToUpper(string(t))
}

func IsBNB(ticker Ticker) bool {
	return ticker.Equals(BNBTicker)
}

func IsRune(ticker Ticker) bool {
	return ticker.Equals(RuneTicker) || ticker.Equals(RuneA1FTicker) || ticker.Equals(RuneB1ATicker)
}
