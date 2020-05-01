package bitcoin

import "fmt"

// DummyUTXOAccessor
type DummyUTXOAccessor struct {
	storage map[string]*BlockMeta
}

func NewDummyUTXOAccessor() *DummyUTXOAccessor {
	return &DummyUTXOAccessor{
		storage: make(map[string]*BlockMeta),
	}
}

func (t *DummyUTXOAccessor) GetBlockMetas() ([]*BlockMeta, error) {
	blockMetas := make([]*BlockMeta, 0)
	for _, item := range t.storage {
		blockMetas = append(blockMetas, item)
	}
	return blockMetas, nil
}
func (t *DummyUTXOAccessor) GetBlockMeta(height int64) (*BlockMeta, error) {
	return nil, nil
}
func (t *DummyUTXOAccessor) SaveBlockMeta(height int64, blockMeta *BlockMeta) error {
	key := fmt.Sprintf(PrefixBlocMeta+"%d", height)
	t.storage[key] = blockMeta
	return nil
}
func (t *DummyUTXOAccessor) PruneBlockMeta(height int64) error {
	return nil
}

func (t *DummyUTXOAccessor) UpsertTransactionFee(fee float64, vSize int32) error {
	return nil
}

func (t *DummyUTXOAccessor) GetTransactionFee() (float64, int32, error) {
	return 0.00018385, 166, nil
}
