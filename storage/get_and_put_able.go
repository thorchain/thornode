package storage

// GetAndPutable is a interface
type GetAndPutable interface {
	Get(key []byte) (value []byte, err error)
	Put(key, value []byte) error
}
