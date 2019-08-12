package signer

import (
	types "gitlab.com/thorchain/bepswap/observe/x/signer/types"
)

type Signer struct {
	PoolAddress string
	DexHost string
	ChainHost string
	Binance *Binance
	StateChain *StateChain
}

func NewSigner(poolAddress, dexHost, chainHost string) *Signer {
	binance := NewBinance(poolAddress, dexHost)
	stateChain := NewStateChain(chainHost)

	return &Signer{
		PoolAddress: poolAddress,
		DexHost: dexHost,
		ChainHost: chainHost,
		Binance: binance,
		StateChain: stateChain,
	}
}

func (s *Signer) Start() {
	// go s.QueryTxn()
}

func (s *Signer) QueryTxn() {
	// s.StateChain.Query()
}

func (s *Signer) SignTxn(outTx types.OutTx) ([]byte, map[string]string) {
	hexTx, param := s.Binance.SignTx(outTx)
	return hexTx, param
}

func (s *Signer) BroadcastTxn(hexTx []byte, param map[string]string) {
	s.Binance.BroadcastTx(hexTx, param)
}
