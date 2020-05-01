package bitcoin

import "fmt"

// DummyBlockMetaAccessor
type DummyBlockMetaAccessor struct {
	storage map[string]*BlockMeta
}

func NewDummyUTXOAccessor() *DummyBlockMetaAccessor {
	return &DummyBlockMetaAccessor{
		storage: make(map[string]*BlockMeta),
	}
}

func (t *DummyBlockMetaAccessor) GetBlockMetas() ([]*BlockMeta, error) {
	blockMetas := make([]*BlockMeta, 0)
	for _, item := range t.storage {
		blockMetas = append(blockMetas, item)
	}
	return blockMetas, nil
}

func (t *DummyBlockMetaAccessor) GetBlockMeta(height int64) (*BlockMeta, error) {
	key := fmt.Sprintf(PrefixBlocMeta+"%d", height)
	blockMeta, ok := t.storage[key]
	if ok {
		return blockMeta, nil
	}
	return nil, nil
}

func (t *DummyBlockMetaAccessor) SaveBlockMeta(height int64, blockMeta *BlockMeta) error {
	key := fmt.Sprintf(PrefixBlocMeta+"%d", height)
	t.storage[key] = blockMeta
	return nil
}

func (t *DummyBlockMetaAccessor) PruneBlockMeta(height int64) error {
	return nil
}

func (t *DummyBlockMetaAccessor) UpsertTransactionFee(fee float64, vSize int32) error {
	return nil
}

func (t *DummyBlockMetaAccessor) GetTransactionFee() (float64, int32, error) {
	return 0.00018385, 166, nil
}
