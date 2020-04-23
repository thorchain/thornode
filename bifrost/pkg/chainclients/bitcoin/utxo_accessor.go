package bitcoin

// UnspentTransactionOutputAccessor define methods to access bitcoin unspent transactional output
type UnspentTransactionOutputAccessor interface {
	GetUTXOs() ([]UnspentTransactionOutput, error)
	AddUTXO(UnspentTransactionOutput) error
	RemoveUTXO(key string) error
	UpsertTransactionFee(fee float64, vSize int64) error
	GetTransactionFee() (float64, int64, error)
}
