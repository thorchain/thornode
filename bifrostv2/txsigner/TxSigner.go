package txsigner

type TxSigner struct {
}

func NewTxSigner() (*TxSigner, error) {
	return &TxSigner{}, nil
}

func (s *TxSigner) Start() error {
	return nil
}

func (s *TxSigner) Stop() error {
	return nil
}
