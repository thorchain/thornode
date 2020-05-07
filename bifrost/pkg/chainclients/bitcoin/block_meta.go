package bitcoin

import (
	"strings"

	"gitlab.com/thorchain/thornode/common"
)

// BlockMeta is a structure to store the blocks bifrost scanned
type BlockMeta struct {
	PreviousHash              string                     `json:"previous_hash"`
	Height                    int64                      `json:"height"`
	BlockHash                 string                     `json:"block_hash"`
	UnspentTransactionOutputs []UnspentTransactionOutput `json:"utxos"`
}

// NewBlockMeta create a new instance of BlockMeta
func NewBlockMeta(previousHash string, height int64, blockHash string) *BlockMeta {
	return &BlockMeta{
		PreviousHash: previousHash,
		Height:       height,
		BlockHash:    blockHash,
	}
}

// GetUTXOs that match the given pubkey and are unspent
func (b *BlockMeta) GetUTXOs(pubKey common.PubKey) []UnspentTransactionOutput {
	utxos := make([]UnspentTransactionOutput, 0, len(b.UnspentTransactionOutputs))
	for _, item := range b.UnspentTransactionOutputs {
		if item.VaultPubKey.Equals(pubKey) && !item.Spent {
			utxos = append(utxos, item)
		}
	}
	return utxos
}

// RemoveUTXO - remove a given UTXO from the storage ,because we already spent it
func (b *BlockMeta) RemoveUTXO(key string) {
	idxToDelete := -1
	for idx, item := range b.UnspentTransactionOutputs {
		if strings.EqualFold(item.GetKey(), key) {
			idxToDelete = idx
			break
		}
	}
	if idxToDelete != -1 {
		b.UnspentTransactionOutputs = append(b.UnspentTransactionOutputs[:idxToDelete], b.UnspentTransactionOutputs[idxToDelete+1:]...)
	}
}

// AddUTXO add the given utxo to blockmeta
func (b *BlockMeta) AddUTXO(utxo UnspentTransactionOutput) {
	for _, u := range b.UnspentTransactionOutputs {
		if u.GetKey() == utxo.GetKey() {
			return
		}
	}
	b.UnspentTransactionOutputs = append(b.UnspentTransactionOutputs, utxo)
}

// SpendUTXO mark a utxo as spent
func (b *BlockMeta) SpendUTXO(key string) {
	for idx, utxo := range b.UnspentTransactionOutputs {
		if key != utxo.GetKey() {
			continue
		}
		b.UnspentTransactionOutputs[idx].Spent = true
		break
	}
}

// UnspendUTXO mark utxo as unspent
func (b *BlockMeta) UnspendUTXO(key string) {
	for idx, utxo := range b.UnspentTransactionOutputs {
		if key != utxo.GetKey() {
			continue
		}
		b.UnspentTransactionOutputs[idx].Spent = false
		break
	}
}
