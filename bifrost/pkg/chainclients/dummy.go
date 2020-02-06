package chainclients

import (
	"errors"

	"github.com/binance-chain/go-sdk/common/types"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
)

var kaboom = errors.New("Kaboom!!!")

// This is a full implementation of a dummy chain, intended for testing purposes

type DummyChain struct{}

func (DummyChain) SignTx(tx stypes.TxOutItem, height int64) ([]byte, error) {
	return nil, kaboom
}
func (DummyChain) BroadcastTx(tx []byte) error                { return kaboom }
func (DummyChain) CheckIsTestNet() (string, bool)             { return "", false }
func (DummyChain) GetHeight() (int64, error)                  { return 0, kaboom }
func (DummyChain) GetAddress(poolPubKey common.PubKey) string { return "" }
func (DummyChain) GetAccount(addr types.AccAddress) (types.BaseAccount, error) {
	return types.BaseAccount{}, kaboom
}
func (DummyChain) GetChain() common.Chain              { return "" }
func (DummyChain) GetGasFee(count uint64) common.Gas   { return nil }
func (DummyChain) ValidateMetadata(_ interface{}) bool { return false }
func (DummyChain) Start()                              {}
func (DummyChain) Stop() error                         { return kaboom }
func (DummyChain) InitBlockScanner(_ string, _ config.BlockScannerConfiguration, _ pubkeymanager.PubKeyValidator, _ *metrics.Metrics) error {
	return kaboom
}
func (DummyChain) GetMessages() <-chan stypes.TxIn                                { return nil }
func (DummyChain) GetTxInForRetry(failedOnly bool) ([]stypes.TxIn, error)         { return nil, kaboom }
func (DummyChain) SetTxInStatus(txIn stypes.TxIn, status stypes.TxInStatus) error { return kaboom }
func (DummyChain) RemoveTxIn(txin stypes.TxIn) error                              { return kaboom }
