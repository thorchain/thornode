package constants

type DummyConstants struct {
}

func (dc *DummyConstants) GetInt64Value(name ConstantName) int64 {
	return 0
}
