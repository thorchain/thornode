package types

import "strings"

const (
	// module pooldata
	ModuleName = "swapservice"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	RouterKey = ModuleName // this was defined in your key.go file
)

func IsRune(ticker string) bool {
	return strings.EqualFold(ticker, RuneTicker)
}
