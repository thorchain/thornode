package chainclients

import (
	"errors"

	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
)

var kaboom = errors.New("Kaboom!!!")

// This is a full implementation of a dummy chain, intended for testing purposes

type DummyChain struct{}

func (DummyChain) SignTx(tx stypes.TxOutItem, height int64) ([]byte, error) {
	return nil, kaboom
}
func (DummyChain) BroadcastTx(_ stypes.TxOutItem, tx []byte) error { return kaboom }
func (DummyChain) CheckIsTestNet() (string, bool)                  { return "", false }
func (DummyChain) GetHeight() (int64, error)                       { return 0, kaboom }
func (DummyChain) GetAddress(poolPubKey common.PubKey) string      { return "" }
func (DummyChain) GetAccount(addr string) (common.Account, error) {
	return common.Account{}, kaboom
}
func (DummyChain) GetChain() common.Chain                { return "" }
func (DummyChain) GetGasFee(count uint64) common.Gas     { return nil }
func (DummyChain) Start(globalTxsQueue chan stypes.TxIn) {}
func (DummyChain) Stop()                                 {}
