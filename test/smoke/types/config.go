package types

import (
	"reflect"
)

// TODO: THORNode are hard coding the attributes to the user names due to bad json
// formats. Fix this later
type BalancesConfig struct {
	Tx       int64            `json:"TX"`
	Out      int64            `json:"OUT"`
	Master   map[string]int64 `json:"MASTER"`
	User1    map[string]int64 `json:"USER-1"`
	Staker1  map[string]int64 `json:"STAKER-1"`
	Staker2  map[string]int64 `json:"STAKER-2"`
	Vault    map[string]int64 `json:"VAULT"`
	PoolBNB  map[string]int64 `json:"POOL-BNB"`
	PoolLoki map[string]int64 `json:"POOL-LOK"`
}

type BalancesConfigs []BalancesConfig

func (b1 BalancesConfig) Equals(b2 BalancesConfig) (bool, string, map[string]int64, map[string]int64) {
	comp := func(b1, b2 map[string]int64) bool {
		if len(b1) == 0 {
			b1 = make(map[string]int64, 0)
		}
		if len(b2) == 0 {
			b2 = make(map[string]int64, 0)
		}
		for k := range b2 {
			if _, ok := b1[k]; !ok {
				b1[k] = 0
			}
		}
		return reflect.DeepEqual(b1, b2)
	}

	if !comp(b1.Master, b2.Master) {
		return false, "Master", b1.Master, b2.Master
	}
	if !comp(b1.User1, b2.User1) {
		return false, "User1", b1.User1, b2.User1
	}
	if !comp(b1.Staker1, b2.Staker1) {
		return false, "Staker1", b1.Staker1, b2.Staker1
	}
	if !comp(b1.Staker2, b2.Staker2) {
		return false, "Staker2", b1.Staker2, b2.Staker2
	}
	if !comp(b1.Vault, b2.Vault) {
		return false, "Vault", b1.Vault, b2.Vault
	}
	if !comp(b1.PoolBNB, b2.PoolBNB) {
		return false, "BNB Pool", b1.PoolBNB, b2.PoolBNB
	}
	if !comp(b1.PoolLoki, b2.PoolLoki) {
		return false, "Loki Pool", b1.PoolLoki, b2.PoolLoki
	}
	return true, "", nil, nil
}

func (b BalancesConfigs) GetByTx(i int64) BalancesConfig {
	for _, bal := range b {
		if bal.Tx == i {
			return bal
		}
	}
	return BalancesConfig{}
}

type Result struct {
	Success     bool
	Transaction TransactionConfig
	Obtained    BalancesConfig
}

type Results []Result

func NewResult(success bool, txn TransactionConfig, bal BalancesConfig) Result {
	return Result{
		Success:     success,
		Transaction: txn,
		Obtained:    bal,
	}
}

func (rs Results) Success() bool {
	for _, r := range rs {
		if !r.Success {
			return false
		}
	}
	return true
}

type TransactionConfig struct {
	Tx    int64            `json:"TX"`
	From  string           `json:"FROM"`
	To    string           `json:"TO"`
	Memo  string           `json:"MEMO"`
	Coins map[string]int64 `json:"COINS"`
}
