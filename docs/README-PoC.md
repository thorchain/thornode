Tokenize Units PoC
==================

The goal of this PoC is to prove that we can mint/burn tokens within cosmos to
represent the amount of "units" a staker has within a given pool

## Workflow
The following lays out the entire procedure of staking/unstaking tokens. It
may make references to structs (S) or API Endpoints (E) documented below.

 1. Jack creates a liquity pool between `RUNE` and `BNB`. All liquity pools must
    contain `RUNE` as one of the two tokens. He calls "Create A Pool" (E),
specifying the details of the pool he is create. This pool is unique, and
cannot be a duplicate of another pool. A struct "Pool" (S) is saved to the
statechain.
 2. Jack wants to stake coins in this new pool and called "List Pools" (E) and
    gets a list of pools including the one he is looking for. Jack stakes `100 BNB` and `100 RUNE`, by sending (via multisend) those coins to the pool address. Once he has a tx hash, he calls "Stake Event" (E) putting his tx details in the request (testnet), or having the cosmos node pulling tx details from binance API). A struct "Tx Event" (S) is saved to the statechain to ensure we cannot register the same event twice
 3. Jack now want to unstake coins. Jack called "Unstake Event" (E) with the
    amount of uToken ("unit tokens") he wishes to remove. Those tokens are
burned from his wallet and the appropriate amount of `RUNE` and `BNB` coins
are sent from the binance pool address to his address.

### Store Structs (S)
#### Pool
We do not need to store much in the pool since most data is publicly
available, like balances and uTokens
```golang
type Pool struct {
  Address string `json:"address"` // unique `BNB` Address to store staked tokens
  TokenName string `json:"token_name"` // display name of token (ie "Bitcoin")
  TokenTicker string `json:"token_ticker"` // ticker name of token (ie "BTC").
Must be uppercase
}
```

#### TX Event
The purpose of this struct is to record when we have successfully processed a
transaction on binance, and ensure we don't duplicate the processing of a
transaction.
```golang
type TxEvent {
  TxHash string `json:"tx_hash" // hash of tx on binance chain
}
```

### API Endpoints (E)
**NOTE** Some endpoints are already implemented by cosmos, so we don't need to
do it again, like account balances

#### Pools
##### Create A Pool `PUT /pool`
Ensure uniqueness and not allow duplicate pools (or overwrite)
```
{
  "token_name": "Binance Coin",
  "token_ticker": "BNB"
}
```

##### Get Pool `GET /pool/<id>`
```json
{
  "token_name": "Binance Coin",
  "token_ticker": "BNB",
  "total_supply": "3028.090" // total amount of uToken supply for this pool
}
```

##### List Pools `GET /pools`
Same as "Get Pool" but listing all pools instead of one.

#### Chain Events
Endpoints to register events on binance chain.

##### Stake Event `POST /binance/tx`
**[TESTNET ONLY]** This endpoint is only for testnet to easily mock out binance
transactions. On mainnet, this endpoint wouldn't exist, and the daemon would
watch for binance txs and process as needed based on memo and to address.
```
{
  "tx_hash": "XXXX", // tx hash ID from a binance transaction
}
```

##### Unstake Event `POST /unstake`
Unstakes tokens, by burning the tokens specified and sending binance tokens to
the same address as the authenticating user. This ensures we can't
accidentally send the unstaked tokens to another address.
```
{
  "to": "bnbXXXXXXXXXXXXX", // bnb address to send coins to
  "coins": [
    "denom": "BNBU", // uToken ticker
    "amount": "34.39857", // amount to unstake
  ]
}
```
