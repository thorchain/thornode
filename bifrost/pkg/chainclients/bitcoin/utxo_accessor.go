package bitcoin

import "gitlab.com/thorchain/thornode/common"

// UnspentTransactionOutputAccessor define methods to access bitcoin unspent transactional output
type UnspentTransactionOutputAccessor interface {
	GetUTXOs(pubKey common.PubKey) ([]UnspentTransactionOutput, error)
	AddUTXO(UnspentTransactionOutput) error
	RemoveUTXO(key string) error
	UpsertTransactionFee(fee float64, vSize int32) error
	GetTransactionFee() (float64, int32, error)
}
