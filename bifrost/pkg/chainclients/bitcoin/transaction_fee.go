package bitcoin

// TransactionFee on bitcoin
type TransactionFee struct {
	Fee   float64 `json:"fee"`
	VSize int64   `json:"v_size"`
}
