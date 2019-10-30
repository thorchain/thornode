# Testing

## Definitions

#### Actors
The smoke tests generate the following `Actors` for executing transactions against the statechain:

##### Faucet
The Binance faucet that funds the master account.

##### Master
The master is funded by the faucet. The master account then seeds all other actors. There is only a single master account.

##### Admin
An admin is what performs all admin transactions (memos prefixed with `ADMIN:`).

##### User
A user simply performs swaps.

##### Staker
A staker simply stakes funds into the pool or pools.

## Tests

For further information on the tests being run, please see [here](https://docs.google.com/spreadsheets/d/1sLK0FE-s6LInWijqKgxAzQk2RiSDZO1GL58kAD62ch0). The purpose of these tests are simply to determine if the Statechain, and its dependant components, are operating as expected, whenever new code is deployed. 

### Lifecycle

A full smoke test lifecycle is as follows:

* Generate the actors;
* SEED the master with funds from the faucet (faucet);
* then SEED the other accounts (admin, user and staker(s));
* then GAS the pool(s);
* then STAKE;
* then SWAP;
* then WITHDRAW;
* then SWEEP all assets back to the faucet from the various actors.

Unit tests (where we've broken the SWAPs and STAKEs into their own test definitions) still follow a variant of the above (as we still need to SEED the actors; GAS, END and ENABLE the pool).

### Scenarios

The test scenarios are all written in JSON and follow a fairly simple format, that should be easy to read.

At the top level we define how many stakers we wish to create, other runtime options as well as our main rules array. 

```json
{
  "with_actors": true,
  "staker_count": 2,
  "sweep_on_exit": true,
  "rules": [...]
}
```

Where:

* `with_actors` create the actors or not (this will override `staker_count`),
* `staker_count` the number of stakers to create,
* `sweep_on_exit` sweep up the pool (and return to the faucet) on completion. We only ever set this to `false` when performing an actual seed of the pools on the `dev` and `staging` environments.

Each rule will have:

```json
{
  {
    "description": "SEED",
    "from": "from",
    "to": [
      "to"
    ],
    "send_to": "staker_1",
    "slip_limit": 1234567,
    "coins": [
      {
        "symbol": "BNB",
        "amount": 1.00000000
      }
    ],
    "memo": "MEMO",
    "check": {}
  }
}
```

Where:

* `description` is a simple description to describe the definition,
* `from` is the actor performing the transaction (e.g: `master`, `admin`, `user`, `staker_N` or `pool`),
* `to` is an array of actors the transaction is for (by using an array, we can support multi-send),
* `coins` is an array of coin objects containing the `symbol` and the `amount` to send,
* `send_to` is the actor to send to, when performing a swap and send (appended to the memo sent),
* `slip_limit` is to set the slip limit (appended to the memo sent)
* `memo` is the memo to use for the transaction
* and `check` defines the rules for validating the transaction (see blow).

#### Validation

After a transaction has been executed, we either check Binance or the Statechain (or sometimes both), to ensure that the resulting balances are inline with our business rules. If this is empty, then the transaction will still be executed, but the result won't be validated.

```json
{
  "delay": 10,
  "binance": {
    "target": "from",
    "coins": [...]
  },
  "statechain": [
    {
      "units": 1.00000000,
      "symbol": "BNB",
      "rune": 1.00000000,
      "asset": 1.00000000,
      "staker_units": [
        {
          "actor": "staker_1",
          "units": 1.00000000
        }
      ]
    }
  ]
}
```

Where:

* `delay` is the number of second to delay running the checks (to ensure ample time is given to both Binance and the Statechain),
* `binance` is an object that contains the `target` actor Binance wallet to check and an array of coin objects (the expected balances - follows the same structure as above)
* and `statechain` is an array of objects that contains the pool `units`, `rune` and `asset` balances to check for a given pool (determined by the `symbol` supplied) as well as a `staker_units` array for validating an actor's share of the pool.

### Running the Tests

The tests are all run via `make`.

#### Main test suite

Please see the test specs [here](https://docs.google.com/spreadsheets/d/1sLK0FE-s6LInWijqKgxAzQk2RiSDZO1GL58kAD62ch0)

```shell script
make FAUCET_KEY=<faucet key> POOL_KEY=<pool key> ENV=<env> smoke-test-refund
make FAUCET_KEY=<faucet key> POOL_KEY=<pool key> ENV=<env> smoke-test-audit-1p
make FAUCET_KEY=<faucet key> POOL_KEY=<pool key> ENV=<env> smoke-test-audit-2p
```

#### Individual (Unit) Tests

These are really only intended to be run when debugging locally - e.g.: you wish to generate noise (without running the entire suite) to see what the Chain Service or other components within the stack observe/report.

##### Gas

```shell script
make FAUCET_KEY=<faucet key> POOL_KEY=<pool key> ENV=<env> gas
```

##### Stake

```shell script
make FAUCET_KEY=<faucet key> POOL_KEY=<pool key> ENV=<env> stake
```

##### Swap

```shell script
make FAUCET_KEY=<faucet key> POOL_KEY=<pool key> ENV=<env> swap
```

For each of the tests you must provide:

* `FAUCET_KEY` this is the private key of the faucet. Without this, the tests will fail as nothing will be funded,
* `POOL_KEY` this is the private key of the pool that that Statechain Observer is observing
* and `ENV` is the environment to run the tests against (can be one of `local`, `develop`, `staging` or `production`).

#### Sweep

While all assets are swept up and returned to the faucet (faucet) on completion of the tests, you can manually perform a sweep by running:

```shell script
make MASTER_KEY=<master key> KEY_LIST=<key list> sweep
```

Where:

* `MASTER_KEY` is the private key of the wallet we wish to transfer assets to
* and `KEY_LIST` is a comma-separated list of private keys we wish to sweep up the assets from.
