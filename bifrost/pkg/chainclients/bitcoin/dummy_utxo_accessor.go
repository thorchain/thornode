package bitcoin

// DummyUTXOAccessor
type DummyUTXOAccessor struct {
	storage map[string]UnspentTransactionOutput
}

func (t *DummyUTXOAccessor) GetUTXOs() ([]UnspentTransactionOutput, error) {
	result := make([]UnspentTransactionOutput, len(t.storage))
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
