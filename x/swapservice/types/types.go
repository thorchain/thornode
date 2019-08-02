package types

import (
	"fmt"
	"strings"
)

// Pool - metadata about a pool
type Pool struct {
	Address     string `json:"address"`      // unique BNB address to store staked tokens
	TokenName   string `json:"token_name"`   // display name of token (ie "Bitcoin")
	TokenTicker string `json:"token_ticker"` // ticker name of token (ie "BTC")
}

func NewPool(name, ticker string) Pool {
	// TODO add address
	return Pool{
		TokenName:   name,
		TokenTicker: strings.ToUpper(ticker),
	}
}

func (p Pool) Key() string {
	return p.TokenTicker
}

func (p Pool) String() string {
	return fmt.Sprintf("Pool: %s (%s) %s", p.TokenName, p.TokenTicker, p.Address)
}

type StakeTx struct {
	TxHash string `json:"tx_hash"` // binance chain tx hash of coins sent to our address
}

func (tx StakeTx) Key() string {
	return tx.TxHash
}

func (tx StakeTx) String() string {
	return fmt.Sprintf("TxHash: %s", tx.TxHash)
}
