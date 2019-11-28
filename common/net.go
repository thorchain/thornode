package common

import (
	"os"
	"strings"
)

func IsTestNet() bool {
	return strings.EqualFold(os.Getenv("NET"), "testnet")
}
