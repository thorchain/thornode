package types

import (
	"fmt"
	"strings"

	"gitlab.com/thorchain/thornode/common"
)

// Query Result Payload for a pools query
type QueryResPools []Pool

// implement fmt.Stringer
func (n QueryResPools) String() string {
	var assets []string
	for _, record := range n {
		assets = append(assets, record.Asset.String())
	}
	return strings.Join(assets, "\n")
}

type QueryResHeights struct {
	Chain            common.Chain `json:"chain"`
	LastChainHeight  int64        `json:"lastobservedin"`
	LastSignedHeight int64        `json:"lastsignedout"`
	Statechain       int64        `json:"statechain"`
}

func (h QueryResHeights) String() string {
	return fmt.Sprintf("Chain: %d, Signed: %d, Statechain: %d", h.LastChainHeight, h.LastSignedHeight, h.Statechain)
}

type ResTxOut struct {
	Height  uint64       `json:"height"`
	Hash    common.TxID  `json:"hash"`
	Chain   common.Chain `json:"chain"`
	TxArray []TxOutItem  `json:"tx_array"`
}

type QueryResTxOut struct {
	Chains map[common.Chain]ResTxOut `json:"chains"`
}
