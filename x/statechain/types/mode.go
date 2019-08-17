package types

import (
	"fmt"
	"strings"
)

type TxMode uint8

const (
	TxUnknown TxMode = iota
	TxAsync
	TxSync
	TxBlock
)

var stringToTxModeMap = map[string]TxMode{
	"unknown": TxUnknown,
	"async":   TxAsync,
	"sync":    TxSync,
	"block":   TxBlock,
}

var txModeToString = map[TxMode]string{
	TxUnknown: "unknown",
	TxAsync:   "async",
	TxSync:    "sync",
	TxBlock:   "block",
}

// converts a string into a TxMode
func stringToTxMode(s string) (TxMode, error) {
	sl := strings.ToLower(s)
	if t, ok := stringToTxModeMap[sl]; ok {
		return t, nil
	}
	return TxUnknown, fmt.Errorf("Invalid tx mode: %s", s)
}

func NewMode(mode string) (TxMode, error) {
	return stringToTxMode(mode)
}

func (tx TxMode) IsValid() bool {
	return tx != TxUnknown
}

func (tx TxMode) String() string {
	return txModeToString[tx]
}
