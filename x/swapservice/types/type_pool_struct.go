package types

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

// PoolStatus is an indication of what the pool state is
type PoolStatus int

//| State | ADMIN-MEMO | Swapping | Staking | Withdrawing | Refunding |
//| ------ | ------ | ------ | ------ | ------ | ------ |
//| `bootstrap` |  `ADMIN:POOL:BOOTSTRAP` | no | yes | yes | Refund Invalid Stakes && all Swaps |
//| `enabled` |  `ADMIN:POOL:ENABLE` | yes | yes | yes | Refund Invalid Tx |
//| `suspended` | `ADMIN:POOL:SUSPEND` | no | no | no | Refund all |
const (
	Enabled PoolStatus = iota
	Bootstrap
	Suspended
)

var poolStatusStr = map[string]PoolStatus{
	"Enabled":   Enabled,
	"Bootstrap": Bootstrap,
	"Suspended": Suspended,
}

// String implement stringer
func (ps PoolStatus) String() string {
	for key, item := range poolStatusStr {
		if item == ps {
			return key
		}
	}
	return ""
}

// MarshalJSON marshal PoolStatus to JSON in string form
func (ps PoolStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(ps.String())
}

// UnmarshalJSON convert string form back to PoolStatus
func (ps *PoolStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); nil != err {
		return err
	}
	*ps = GetPoolStatus(s)
	return nil
}

// GetPoolStatus from string
func GetPoolStatus(ps string) PoolStatus {
	for key, item := range poolStatusStr {
		if strings.EqualFold(key, ps) {
			return item
		}
	}

	return Suspended
}

// PoolStruct is a struct that contains all the metadata of a pooldata
// This is the structure we will saved to the key value store
type PoolStruct struct {
	BalanceRune  string     `json:"balance_rune"`  // how many RUNE in the pool
	BalanceToken string     `json:"balance_token"` // how many token in the pool
	Ticker       Ticker     `json:"ticker"`        // what's the token's ticker
	PoolUnits    string     `json:"pool_units"`    // total units of the pool
	PoolAddress  string     `json:"pool_address"`  // pool address on binance chain
	Status       PoolStatus `json:"status"`        // status
}

// NewPoolStruct Returns a new PoolStruct
func NewPoolStruct() PoolStruct {
	return PoolStruct{
		BalanceRune:  "0",
		BalanceToken: "0",
		PoolUnits:    "0",
		Status:       Bootstrap,
	}
}

func (ps PoolStruct) Empty() bool {
	return ps.Ticker == ""
}

// String implement fmt.Stringer
func (ps PoolStruct) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintln("rune-balance: " + ps.BalanceRune))
	sb.WriteString(fmt.Sprintln("token-balance: " + ps.BalanceToken))
	sb.WriteString(fmt.Sprintln("ticker: " + ps.Ticker))
	sb.WriteString(fmt.Sprintln("pool-units: " + ps.PoolUnits))
	sb.WriteString(fmt.Sprintln("pool-address: " + ps.PoolAddress))
	sb.WriteString(fmt.Sprintln("status: " + ps.Status.String()))
	return sb.String()
}

// EnsureValidPoolStatus
func (ps PoolStruct) EnsureValidPoolStatus(msg sdk.Msg) error {
	switch ps.Status {
	case Enabled:
		return nil
	case Bootstrap:
		switch msg.(type) {
		case MsgSwap:
			return errors.New("pool is in bootstrap status, can't swap")
		default:
			return nil
		}
	case Suspended:
		return errors.New("pool suspended")
	default:
		return errors.Errorf("unknown pool status,%s", ps.Status)
	}
}
