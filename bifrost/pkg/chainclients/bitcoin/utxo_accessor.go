package bitcoin

// UnspentTransactionOutputAccessor define methods to access bitcoin unspent transactional output
type UnspentTransactionOutputAccessor interface {
	GetUTXOs() ([]UnspentTransactionOutput, error)
	AddUTXO(UnspentTransactionOutput) error
	RemoveUTXO(key string) error
}
