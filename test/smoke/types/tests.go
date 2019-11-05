package types

import (
	"time"

	sdk "github.com/binance-chain/go-sdk/client"
	"github.com/binance-chain/go-sdk/keys"
)

type Tests struct {
	ActorList   []string `json:"actor_list"`
	ActorKeys   map[string]Keys
	SweepOnExit bool   `json:"sweep_on_exit"`
	Rules       []Rule `json:"rules"`
}

type Keys struct {
	Key    keys.KeyManager
	Client sdk.DexClient
}

type Rule struct {
	Description string        `json:"description"`
	From        string        `json:"from"`
	To          []To          `json:"to"`
	Memo        string        `json:"memo"`
	SendTo      string        `json:"send_to"`
	SlipLimit   int64         `json:"slip_limit"`
	CheckDelay  time.Duration `json:"check_delay"`
}

type To struct {
	Actor string `json:"actor"`
	Coins []Coin `json:"coins"`
}

type Coin struct {
	Symbol string `json:"symbol"`
	Amount int64  `json:"amount"`
}
