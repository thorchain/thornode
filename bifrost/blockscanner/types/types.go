package types

import "fmt"

var UnavailableBlock error = fmt.Errorf("block is not yet available")
