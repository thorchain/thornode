package types

import (
	"time"

	sdk "github.com/binance-chain/go-sdk/client"
	"github.com/binance-chain/go-sdk/keys"
)

type Tests struct {
	WithActors  bool `json:"with_actors"`
	StakerCount int  `json:"staker_count"`
	SweepOnExit bool `json:"sweep_on_exit"`
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
	SendTo      string   `json:"send_to"`
	SlipLimit   int64    `json:"slip_limit"`
	Check       Check    `json:"check"`
}

type Coin struct {
	Symbol string  `json:"symbol"`
	Amount int64 `json:"amount"`
}

type Check struct {
	Delay      time.Duration `json:"delay"`
	Binance    Binance       `json:"binance"`
	Statechain []Statechain  `json:"statechain"`
}

type Binance struct {
	Target string `json:"target"`
	Coins  []Coin `json:"coins"`
}

type Statechain struct {
	Units       int64       `json:"units"`
	Symbol      string        `json:"symbol"`
	Rune        int64       `json:"rune"`
	Asset       int64       `json:"asset"`
	Status      string        `json:"status"`
	StakerUnits []StakerUnits `json:"staker_units,omitempty"`
}

type StakerUnits struct {
	Actor string  `json:"actor"`
	Units int64 `json:"units"`
}
