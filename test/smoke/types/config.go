package types

import (
	"reflect"
)

type BalanceExport struct {
	Tx           int64 `json:"TX"`
	Out          int64 `json:"OUT"`
	MasterRune   int64 `json:"MASTER/RUNE-A1F"`
	MasterLoki   int64 `json:"MASTER/LOK-3CO"`
	MasterBNB    int64 `json:"MASTER/BNB"`
	User1Rune    int64 `json:"USER-1/RUNE-A1F"`
	User1Loki    int64 `json:"USER-1/LOK-3CO"`
	User1BNB     int64 `json:"USER-1/BNB"`
	Staker1Rune  int64 `json:"STAKER-1/RUNE-A1F"`
	Staker1Loki  int64 `json:"STAKER-1/LOK-3CO"`
	Staker1BNB   int64 `json:"STAKER-1/BNB"`
	Staker2Rune  int64 `json:"STAKER-2/RUNE-A1F"`
	Staker2Loki  int64 `json:"STAKER-2/LOK-3CO"`
	Staker2BNB   int64 `json:"STAKER-2/BNB"`
	VaultRune    int64 `json:"VAULT/RUNE-A1F"`
	VaultLoki    int64 `json:"VAULT/LOK-3CO"`
	VaultBNB     int64 `json:"VAULT/BNB"`
	PoolBNBRune  int64 `json:"POOL-BNB/RUNE-A1F"`
	PoolBNBBNB   int64 `json:"POOL-BNB/BNB"`
	PoolLokiRune int64 `json:"POOL-LOK/RUNE-A1F"`
	PoolLokiLoki int64 `json:"POOL-LOK/LOK-3C0"`
}

// TODO: THORNode are hard coding the attributes to the user names due to bad
// json formats. Fix this later
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

func (b BalancesConfig) Export() BalanceExport {
	return BalanceExport{
		Tx:           b.Tx,
		Out:          b.Out,
		MasterRune:   b.Master["RUNE-A1F"],
		MasterLoki:   b.Master["LOK-3C0"],
		MasterBNB:    b.Master["BNB"],
		User1Rune:    b.User1["RUNE-A1F"],
		User1Loki:    b.User1["LOK-3C0"],
		User1BNB:     b.User1["BNB"],
		Staker1Rune:  b.Staker1["RUNE-A1F"],
		Staker1Loki:  b.Staker1["LOK-3C0"],
		Staker1BNB:   b.Staker1["BNB"],
		Staker2Rune:  b.Staker2["RUNE-A1F"],
		Staker2Loki:  b.Staker2["LOK-3C0"],
		Staker2BNB:   b.Staker2["BNB"],
		VaultRune:    b.Vault["RUNE-A1F"],
		VaultLoki:    b.Vault["LOK-3C0"],
		VaultBNB:     b.Vault["BNB"],
		PoolBNBRune:  b.PoolBNB["RUNE-A1F"],
		PoolBNBBNB:   b.PoolBNB["BNB"],
		PoolLokiRune: b.PoolLoki["RUNE-A1F"],
		PoolLokiLoki: b.PoolLoki["LOK-3C0"],
	}
}

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
