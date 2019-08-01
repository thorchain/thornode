BEPSwap Statechain
=======================
Swaps any two coins on the Binance Exchange on top of the [Cosmos network](cosmos.network)
[![Build Status](https://gitlab.com/thorchain/bepswap/statechain/badges/master/build.svg)](https://gitlab.com/thorchain/bepswap/statechain/commits/master)

An account can stake coins into a pool that is composed of `ATOM` and any
other single token. Once a pool has been created by an account(s) staking
their coins, people can than swap/trade within that pool to gain one token
over another.

How many tokens an individual gets from another token is relative to how many
tokens of each exist in the pool. Say you have a pool of two tokens, `ATOM`
and `RUNE`. There are 20 `RUNE` tokens, and 50 `ATOM` tokens. If you swap 10
`RUNE` tokens, we add those to the pool (for a total of 30), see that our
addition makes up for 1/3rd of the present tokens, which means we get 1/3rd of
the `ATOM` tokens, which is 16.6666666667. This ensures that we never run out
of tokens in a pool from swapping.

## Setup
Ensure you have a recent version of go (ie `1.121) and enabled go modules
```
export GO111MODULE=on
```
And have `GOBIN` in your `PATH`
```
export GOBIN=$GOPATH/bin
```

### Automated Install
To install easily, run the following command...
```bash
make setup
```

### Manual Install
Install via this `make` command.

```bash
make install
```

Once you've installed `sscli` and `ssd`, check that they are there.

```bash
sscli help
ssd help
```

### Configuration

Next configure your chain.
```bash
# Initialize configuration files and genesis file
# moniker is the name of your node
ssd init <moniker> --chain-id sschain

# Copy the Address output here and save it for later use
# [optional] add "--ledger" at the end to use a Ledger Nano S
sscli keys add jack

# Copy the Address output here and save it for later use
sscli keys add alice

# Add both accounts, with coins to the genesis file
ssd add-genesis-account $(sscli keys show jack -a) 1000rune,100000000stake
rune add-genesis-account $(sscli keys show alice -a) 1000rune,100000000stake

# Configure your CLI to eliminate need for chain-id flag
sscli config chain-id sschain
sscli config output json
sscli config indent true
sscli config trust-node true

ssd gentx --name jack
```

## Start
There are three services you may want to start.

#### Daemon
This runs the backend
```bash
make start
```

#### REST API Service
Starts an HTTP service to service requests to the backend.
```bash
make start-rest
```

#### CORS Proxy
For making requests in a browser to the API backend, you'll need to start a
proxy in front of the API service to give proper CORS headers. This is because
CORS support in Cosmos was removed. In the meantime, use
[cors.io](http://cors.io) as a proxy.
