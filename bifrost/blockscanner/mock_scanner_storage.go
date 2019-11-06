package blockscanner

import (
	"encoding/binary"
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
)

const MockErrorBlockHeight = 1024

// MockScannerStorage is to mock scanner storage interface
type MockScannerStorage struct {
	l     *sync.Mutex
	store map[string][]byte
}

// NewMockScannerStorage create a new instance of MockScannerStorage
func NewMockScannerStorage() *MockScannerStorage {
	return &MockScannerStorage{
		store: make(map[string][]byte),
		l:     &sync.Mutex{},
	}
}

func (mss *MockScannerStorage) GetScanPos() (int64, error) {
	buf, ok := mss.store[ScanPosKey]
	if !ok {
		return 0, errors.New("scan pos doesn't exist")
	}
	pos, _ := binary.Varint(buf)
	return pos, nil
}
func (mss *MockScannerStorage) SetScanPos(block int64) error {
	mss.l.Lock()
	defer mss.l.Unlock()
	buf := make([]byte, 8)
	n := binary.PutVarint(buf, block)
	mss.store[ScanPosKey] = buf[:n]
	return nil
}
func (mss *MockScannerStorage) SetBlockScanStatus(block int64, status BlockScanStatus) error {
	blockStatusItem := BlockStatusItem{
		Height: block,
		Status: status,
	}
	buf, err := json.Marshal(blockStatusItem)
	if nil != err {
		return errors.Wrap(err, "fail to marshal BlockStatusItem to json")
	}
	mss.l.Lock()
	defer mss.l.Unlock()
	mss.store[getBlockStatusKey(block)] = buf
	return nil
}
func (mss *MockScannerStorage) RemoveBlockStatus(block int64) error {
	mss.l.Lock()
	defer mss.l.Unlock()
	delete(mss.store, getBlockStatusKey(block))
	return nil
}
func (mss *MockScannerStorage) GetBlocksForRetry(failedOnly bool) ([]int64, error) {
	return nil, nil
}
func (mss *MockScannerStorage) Close() error {
	return nil
}
