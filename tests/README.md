# Testing

## Definitions

### Actors
The smoke tests generate the following `Actors` for executing transactions against the statechain:

#### Bank
The Binance faucet that funds the master account.

#### Master
The master is funded by the bank. The master account then seeds all other actors. There is only a single master account.

#### Admin
An admin is what performs all admin transactions (memos prefixed with `ADMIN:`).

#### User
A user simply performs swaps.

#### Staker
A staker simply stakes funds into the pool or pools.
