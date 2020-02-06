package chainclients

import (
	"errors"

	"github.com/binance-chain/go-sdk/common/types"
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
