package types

import (
	sdk "github.com/binance-chain/go-sdk/client"
	"github.com/binance-chain/go-sdk/keys"
)

type Tests struct {
	StakerCount int `json:"staker_count"`
	Actors      Actors
	Rules       []Rule `json:"rules"`
}

type Actors struct {
	Faucet  Keys
	Master  Keys
	Admin   Keys
	Stakers []Keys
	User    Keys
	Pool    Keys
}

type Keys struct {
	Key    keys.KeyManager
	Client sdk.DexClient
}

type Rule struct {
	Description string   `json:"description"`
	From        string   `json:"from"`
	To          []string `json:"to"`
	Coins       []Coin   `json:"coins"`
	Memo        string   `json:"memo"`
	Check       Check    `json:"check"`
}

type Coin struct {
	Symbol string  `json:"symbol"`
	Amount float64 `json:"amount"`
}

type Check struct {
	Target     string     `json:"target"`
	Binance    []Coin     `json:"binance"`
	Statechain Statechain `json:"statechain"`
}

type Statechain struct {
	Units       float64       `json:"units"`
	Symbol      string        `json:"symbol"`
	Rune        float64       `json:"rune"`
	Token       float64       `json:"token"`
	Status      string        `json:"status"`
	StakerUnits []StakerUnits `json:"staker_units,omitempty"`
}

type StakerUnits struct {
	Actor string  `json:"actor"`
	Units float64 `json:"units"`
}
