package signer

type TxStatus int

const (
	TxUnknown TxStatus = iota
	TxAvailable
	TxUnavailable
	TxSpent
)
