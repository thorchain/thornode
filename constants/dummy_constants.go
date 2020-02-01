package constants

type DummyConstants struct {
	int64values  map[ConstantName]int64
	boolValues   map[ConstantName]bool
	stringValues map[ConstantName]string
}

// NewDummyConstants create a new instance of DummyConstants for test purpose
func NewDummyConstants(int64Values map[ConstantName]int64, boolValues map[ConstantName]bool, stringValues map[ConstantName]string) *DummyConstants {
	return &DummyConstants{
		int64values:  int64Values,
		boolValues:   boolValues,
		stringValues: stringValues,
	}
}

func (dc *DummyConstants) GetInt64Value(name ConstantName) int64 {
	v, ok := dc.int64values[name]
	if !ok {
		return 0
	}
	return v
}

func (dc *DummyConstants) GetBoolValue(name ConstantName) bool {
	v, ok := dc.boolValues[name]
	if !ok {
		return false
	}
	return v
}

func (dc *DummyConstants) GetStringValue(name ConstantName) string {
	v, ok := dc.stringValues[name]
	if !ok {
		return ""
	}
	return v
}

func (dc *DummyConstants) String() string {
	return ""
}
