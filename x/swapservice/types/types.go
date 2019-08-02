package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Pool - metadata about a pool
type Pool struct {
	Address     sdk.AccAddress `json:"address"`      // unique BNB address to store staked tokens
	TokenName   string         `json:"token_name"`   // display name of token (ie "Bitcoin")
	TokenTicker string         `json:"token_ticker"` // ticker name of token (ie "BTC")
}

// TODO: create address
func NewPool(name, ticker string) Pool {
	return Pool{
		TokenName:   name,
		TokenTicker: strings.ToUpper(ticker),
	}
}

func (p Pool) Key() string {
	return p.TokenTicker
}

func (p Pool) Empty() bool {
	return p.Address.Empty() || p.TokenTicker == ""
}

func (p Pool) String() string {
	return fmt.Sprintf("Pool: %s (%s) %s", p.TokenName, p.TokenTicker, p.Address)
}

type TxHash struct {
	TxHash string `json:"tx_hash"` // binance chain tx hash of coins sent to our address
	// TODO: allow this field to be updated
	Refunded bool `json:"refund"` // if tx has been refunded back to original wallet
}

func NewTxHash(txHash string) TxHash {
	return TxHash{
		TxHash: txHash,
	}
}

func (tx TxHash) Key() string {
	return tx.TxHash
}

func (tx TxHash) Empty() bool {
	return tx.TxHash == ""
}

func (tx TxHash) String() string {
	return fmt.Sprintf("TxHash: %s", tx.TxHash)
}
