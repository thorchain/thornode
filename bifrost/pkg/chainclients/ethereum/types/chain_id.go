package types

// Ethereum chain id type
type ChainID int

const (
	Mainnet ChainID = iota + 1
	_
	Ropsten
	Rinkeby
	Localnet = iota + 15
)
