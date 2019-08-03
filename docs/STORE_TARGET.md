Store Target
============

The purpose of this document is to outline the list of data to be stored in
Cosmos.

## Pool
```golang
// Pool contains metadata about a liquity pool.
type Pool struct {
    PoolID       string `json:"pool_id"`       // ie "pool-BNB"
    BalanceRune  string `json:"balance_rune"`  // how many RUNE in the pool
    BalanceToken string `json:"balance_token"` // how many token in the pool
    Ticker       string `json:"ticker"`        // what's the token's ticker
(ie "BNB" or "BTC")
    TokenName    string `json:"token_name"`    // what's the token's name, ie
"Binance" or "Bitcoin"
    PoolUnits    string `json:"pool_units"`    // total units of the pool
    PoolAddress  string `json:"pool_address"`  // pool address on binance chain
    Status       string `json:"status"`        // TODO: add description
}
```

## PoolIndex
```golang
// a list of all pools by pool ID
type PoolIndex []string
```

## PoolStaker
```golang
// PoolStaker stores the amount of units each staker has in a given pool
type PoolStaker struct {
    PoolID     string       `json:"pool_id"`     // ie pool-BNB
    TotalUnits string       `json:"total_units"` // total units in the pool
    Stakers    []StakerUnit `json:"stakers"`     // key will be staker id , which is the address on binane chain value will be UNITS
}
```
This struct can hold a lot of data (if there are thousands of stakers in a
chain), and it will get updated often (every time someone stakes or unstake in
a pool). This could be a problem with storage because I believe everytime we
update the contents of a key in the kvstore, it writes the entire contents to
disk. Maybe each staker should be its own kv pair so that writes aren't
expensive in terms of disk activity and storage? Then we could either
implement an index (like `PoolIndex`), or use an iterator with a prefix.
Considering the number of stakers can be really large, we again run into the
same problem or being disk expensive every update.

Is `TotalUnits` a duplication of `Pool.PoolUnits`? Storing the same data in
two different places without a source of truth can be dangerous.

### StakerUnit
Stores how much units a staker holds
```golang
type StakerUnit struct {
    StakerID string `json:"staker_id"` // Staker bnb address
    Units    string `json:"units"` // number of units the stake owns
}
```

## StakerPool
Stores all of the pools a staker is holds for
```golang
type StakerPool struct {
    StakerID  string           `json:"staker_id"`      // this will be staker's address on binance chain
    PoolUnits []StakerPoolItem `json:"pool_and_units"` // the key of this map will be the pool id , value will bt [UNIT,RUNE,TOKEN]
}
```

### StakerPoolItem
Stores how many units, a stake has for a specific pool.
```golang
type StakerPoolItem struct {
    PoolID       string `json:"pool_id"`        // ie pool-BNB
    Units        string `json:"units"`          // number of units a staker has in this
pool
    RuneBalance  string `json:"rune_balance"`   // number of rune coins he
staked in the pool
    TokenBalance string `json:"token_balance"`  // number of token coins he
staked in the pool
}
```

## SwapRecord
```golang
type SwapRecord struct {
    RequestTxHash   string `json:"request_tx_hash"`  // The TxHash on binance chain represent user send token to the pool
    SourceTicker    string `json:"source_ticker"`    // Source ticker
    TargetTicker    string `json:"target_ticker"`    // Target ticker
    Requester       string `json:"requester"`        // Requester , should be the address on binance chain
    Destination     string `json:"destination"`      // destination, address to
send the tokens
    AmountRequested string `json:"amount_requested"` // amount of source token in
    AmountPaidBack  string `json:"amount_paid_back"` // amount of target token pay out to user
    PayTxHash       string `json:"pay_tx_hash"`      // TxHash on binance chain represent our pay to user
}
```
