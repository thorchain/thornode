package bitcoin

import "gitlab.com/thorchain/thornode/common"

// DummyUTXOAccessor
type DummyUTXOAccessor struct {
	storage map[string]UnspentTransactionOutput
}

func NewDummyUTXOAccessor() *DummyUTXOAccessor {
	return &DummyUTXOAccessor{
		storage: make(map[string]UnspentTransactionOutput),
	}
}

func (t *DummyUTXOAccessor) GetUTXOs(pKey common.PubKey) ([]UnspentTransactionOutput, error) {
	result := make([]UnspentTransactionOutput, 0, len(t.storage))
	for _, item := range t.storage {
		result = append(result, item)
	}
	return result, nil
}

func (t *DummyUTXOAccessor) AddUTXO(u UnspentTransactionOutput) error {
	t.storage[u.GetKey()] = u
	return nil
}

func (t *DummyUTXOAccessor) RemoveUTXO(key string) error {
	delete(t.storage, key)
	return nil
}

func (t *DummyUTXOAccessor) UpsertTransactionFee(fee float64, vSize int32) error {
	return nil
}

func (t *DummyUTXOAccessor) GetTransactionFee() (float64, int32, error) {
	return 0.00018385, 166, nil
}
