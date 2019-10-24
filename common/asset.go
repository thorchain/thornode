package common

import (
	"fmt"
	"strings"
)

var (
	BNBAsset     = Asset{"BNB", "BNB", "BNB"}
	RuneA1FAsset = Asset{"BNB", "RUNE-A1F", "RUNE"}
	RuneB1AAsset = Asset{"BNB", "RUNE-B1A", "RUNE"}
)

type Asset struct {
	Chain  Chain  `json:"chain"`
	Symbol Symbol `json:"symbol"`
	Ticker Ticker `json:"ticker"`
}

func NewAsset(input string) (Asset, error) {
	var err error

	asset := Asset{}
	parts := strings.Split(input, ".")
	if len(parts) != 2 {
		return Asset{}, fmt.Errorf("bad asset format")
	}
	asset.Chain, err = NewChain(parts[0])
	if err != nil {
		return Asset{}, err
	}

	asset.Symbol, err = NewSymbol(parts[1])
	if err != nil {
		return Asset{}, err
	}

	parts = strings.Split(parts[1], "-")
	asset.Ticker, err = NewTicker(parts[0])
	if err != nil {
		return Asset{}, err
	}

	return asset, nil
}

func (a Asset) Equals(a2 Asset) bool {
	return a.Chain.Equals(a2.Chain) && a.Symbol.Equals(a2.Symbol) && a.Ticker.Equals(a2.Ticker)
}

func (a Asset) IsEmpty() bool {
	return a.Chain.IsEmpty() || a.Symbol.IsEmpty() || a.Ticker.IsEmpty()
}

func (a Asset) String() string {
	return fmt.Sprintf("%s.%s", a.Chain.String(), a.Symbol.String())
}

func IsBNBAsset(a Asset) bool {
	return a.Equals(BNBAsset)
}

func IsRuneAsset(a Asset) bool {
	return a.Equals(RuneA1FAsset) || a.Equals(RuneB1AAsset)
}
