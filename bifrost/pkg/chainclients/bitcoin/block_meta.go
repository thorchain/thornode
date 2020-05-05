package bitcoin

import (
	"fmt"
	"strings"

	"gitlab.com/thorchain/thornode/common"
)

// BlockMeta is a structure to store the blocks bifrost scanned
type BlockMeta struct {
	PreviousHash              string                     `json:"previous_hash"`
	Height                    int64                      `json:"height"`
	BlockHash                 string                     `json:"block_hash"`
	UnspentTransactionOutputs []UnspentTransactionOutput `json:"utxos"`
	TxIDs                     []string                   `json:"tx_ids"`
}

// NewBlockMeta create a new instance of BlockMeta
func NewBlockMeta(previousHash string, height int64, blockHash string) *BlockMeta {
	return &BlockMeta{
		PreviousHash: previousHash,
		Height:       height,
		BlockHash:    blockHash,
	}
}

// GetUTXOs that match the given pubkey
func (b *BlockMeta) GetUTXOs(pubKey common.PubKey) []UnspentTransactionOutput {
	utxos := make([]UnspentTransactionOutput, 0, len(b.UnspentTransactionOutputs))
	for _, item := range b.UnspentTransactionOutputs {
		if item.VaultPubKey.Equals(pubKey) {
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
			fmt.Println("===============remove utxo found==============")
			fmt.Println(key)
			fmt.Println("===============remove utxo found==============")
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

// AddTxID add tx id on a block meta to track which tx we processed
func (b *BlockMeta) AddTxID(txID string) {
	for _, blockTxID := range b.TxIDs {
		if txID == blockTxID {
			return
		}
	}
	b.TxIDs = append(b.TxIDs, txID)
}
