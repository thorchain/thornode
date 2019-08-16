package types

import (
	"fmt"
	"strings"
)

type TxMode uint8

const (
	txUnknown TxMode = iota
	txAsync
	txSync
	txBlock
)

var stringToTxModeMap = map[string]TxMode{
	"unknown": txUnknown,
	"async":   txAsync,
	"sync":    txSync,
	"block":   txBlock,
}

var txModeToString = map[TxMode]string{
	txUnknown: "unknown",
	txAsync:   "async",
	txSync:    "sync",
	txBlock:   "block",
}

// converts a string into a TxMode
func stringToTxMode(s string) (TxMode, error) {
	sl := strings.ToLower(s)
	if t, ok := stringToTxModeMap[sl]; ok {
		return t, nil
	}
	return txUnknown, fmt.Errorf("Invalid tx mode: %s", s)
}

func NewMode(mode string) (TxMode, error) {
	return stringToTxMode(mode)
}

func (tx TxMode) IsValid() bool {
	return tx != txUnknown
}

func (tx TxMode) String() string {
	return txModeToString[tx]
}
