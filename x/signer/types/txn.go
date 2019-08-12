package types

type TxArr struct {
	RequestTxHash string
	From          string
	To            string
	Token         string
	Amount        string
}

type OutTx struct {
	TxOutID string
	Hash    string
	TxArray []TxArr
}
