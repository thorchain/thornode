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

Unit tests (where we've broken the SWAPs and STAKEs into their own test definitions) still follow a variant of the above (as THORNode still need to SEED the actors; GAS and WITHDRAW the pool).

### Scenarios

The test scenarios are all written in JSON and follow a fairly simple format, that should be easy to read.

At the top level THORNode define how many stakers THORNode wish to create, other runtime options as well as our main rules array. 

```json
{
  "actor_list": ["master", "admin", "user", "staker_1", "staker_2"],
  "sweep_on_exit": true,
  "rules": [...]
}
```

Where:

* `actor_list` is a list of all the actors to create
* and `sweep_on_exit` will sweep up the pool (and return to the faucet) on completion. We only ever set this to `false` when performing an actual seed of the pools on the `dev` and `staging` environments.

Each rule will have:

```json
{
  {
    "description": "SEED",
    "from": "faucet",
    "to": [
      {
        "actor": "master",
        "coins": [
          {
            "symbol": "BNB",
            "amount": 100000000
          }
        ]
      }
    ],
    "send_to": "staker_1",
    "slip_limit": 1234567,
    "memo": "MEMO",
    "check_delay": 10
  }
}
```

Where:

* `description` is a simple description to describe the definition,
* `from` is the actor performing the transaction (e.g: `master`, `admin`, `user`, `staker_N` or `pool`),
* `to` is an array of actors and the coins to send (an array means that THORNode can support multi-send),
* `send_to` is the actor to send to, when performing a swap and send (appended to the memo sent),
* `slip_limit` is to set the slip limit (appended to the memo sent)
* `memo` is the memo to use for the transaction
* and `check_delay` is the delay between broadcasting the transaction to Binance, and checking the balances (Binance and Statechain).

#### Validation

After a transaction has been executed, THORNode check Binance and the Statechain. The output is saved as JSON into `/tmp/smoke.json` by default.

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

* `MASTER_KEY` is the private key of the wallet THORNode wish to transfer assets to
* and `KEY_LIST` is a comma-separated list of private keys THORNode wish to sweep up the assets from.
