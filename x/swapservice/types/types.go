package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Initial Starting Price for a pooldata that was never previously owned
var MinPoolDataPrice = sdk.Coins{sdk.NewInt64Coin("atom", 1)}

// PoolStruct is a struct that contains all the metadata of a pooldata
type PoolStruct struct {
	Value string         `json:"value"`
	Owner sdk.AccAddress `json:"owner"`
	Price sdk.Coins      `json:"price"`
}

// Returns a new PoolStruct with the minprice as the price
func NewPoolStruct() PoolStruct {
	return PoolStruct{
		Price: MinPoolDataPrice,
	}
}

// implement fmt.Stringer
func (w PoolStruct) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Owner: %s
Value: %s
Price: %s`, w.Owner, w.Value, w.Price))
}
