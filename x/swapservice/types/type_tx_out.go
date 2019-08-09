package types

// TxOut is a structure represent all the tx we need to return to client
type TxOut struct {
	Height  int64        `json:"height"`
	Hash    string       `json:"hash"`
	TxArray []*TxOutItem `json:"tx_array"`
}

// NewTxOut create a new item ot TxOut
func NewTxOut(height int64) *TxOut {
	return &TxOut{
		Height:  height,
		TxArray: nil,
	}
}
