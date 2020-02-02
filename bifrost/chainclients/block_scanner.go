package chainclients

import (
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

// BlockScanner defines all public methods used in client and observer
type BlockScanner interface {
	GetMessages() <-chan types.TxIn
	Start()
	MatchedAddress(txInItem types.TxInItem) bool
	Stop() error
}
