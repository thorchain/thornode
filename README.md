[![pipeline status](https://gitlab.com/thorchain/thornode/badges/master/pipeline.svg)](https://gitlab.com/thorchain/thornode/commits/master)
[![coverage report](https://gitlab.com/thorchain/thornode/badges/master/coverage.svg)](https://gitlab.com/thorchain/thornode/commits/master)
[![Build Status](https://gitlab.com/thorchain/thornode/badges/master/build.svg)](https://gitlab.com/thorchain/thornode/commits/master)

# BEPSwap Statechain
=======================

Swap any two coins (BEP2 Assets) on Binance Chain using a statechain built with [CosmosSDK.](cosmos.network)

The BEPSwap statechain comes to consensus about events observed on Binance Chain, and then applies logic to these finalised events. Each event causes a state change in the statechain, and some events also result in an output transaction which require assets to be moved. These output transactions are then batched, signed by a threshold signature scheme protocol and broadcast back to Binance Chain. 

The BEPSwap statechain can be thought of an elaborate multi-signature wallet on Binance Chain, which has joint custody of assets and strict, deterministic caveats on how to spend. All BEPSwap validators have a co-located Observer and Signer, which together mean they can all agree on what assets they control, what requests for spending are made by users, and how to perform these spends. 

### Transactions 
The BEPSwap Statechain facilitates the following transactions, which are made on Binance Chain, and replayed into the statechain via the Observers:
- **CREATE POOL**: Anyone can create a new BEP2 Pool
- **STAKE**: Anyone can stake assets in those pools
- **WITHDRAW**: Anyone who is staking can withdraw their claim on the pool
- **SWAP**: Anyone can send in assets and swap to another, including sending to a destination address, and including optional price protection. 
- **ADD**: Anyone can add assets into the pool, which can be claimed by stakers
- **GAS**: Anyone can add `BNB` gas to ensure transactions can be processed
- **ADMIN**: Whitelisted admins can request to change global and pool-based parameters, but majority consensus is required. 

The Staking, Withdrawing and Swapping logic is based on the `XYK` continuous liquidity pool concept. 

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

Once you've installed `thorcli` and `thord`, check that they are there.

```bash
thorcli help
thord help
```

### Configuration

Next configure your chain.
```bash
# Initialize configuration files and genesis file
# moniker is the name of your node
thord init <moniker> --chain-id sschain

# Copy the Address output here and save it for later use
# [optional] add "--ledger" at the end to use a Ledger Nano S
thorcli keys add jack

# Copy the Address output here and save it for later use
thorcli keys add alice

# Add both accounts, with coins to the genesis file
thord add-genesis-account $(thorcli keys show jack -a) 1000rune,100000000stake
rune add-genesis-account $(thorcli keys show alice -a) 1000rune,100000000stake

# Configure your CLI to eliminate need for chain-id flag
thorcli config chain-id sschain
thorcli config output json
thorcli config indent true
thorcli config trust-node true

thord gentx --name jack
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

BEPSwap Observe
===============

Observe events that take place on Binance.

### Environment
Please create a config.json file
```json
{
  "chain_id": "statechain",
  "pool_address": "pool address",
  "rune_address": "rune address",
  "dex_host": "testnet-dex.binance.org",
  "observer_db_path": "signerdata",
  "block_scanner": {
    "rpc_host": "binance node host",
    "start_block_height": 34912415,
    "block_scan_processors": 2,
    "block_height_discover_back_off": "1s"
  },
  "state_chain": {
    "chain_id": "statechain",
    "chain_host": "localhost:1317",
    "signer_name": "signer name",
    "signer_passwd": "signer password"
  }
}
You could overwrite the above configs using environment variables as well
Export the following environment variables (set based on your environment and/or the net being used [test/prod]):
```bash
export CHAIN_ID=<chain id>
export POOL_ADDRESS=<pool address>
export RUNE_ADDRESS=<rune address>,
export DEX_HOST=<DEX Hostname>
export BLOCK_SCANNER_RPC_HOST=<RPC HOSTNAME>
export BLOCK_SCANNER_START_BLOCK_HEIGHT=34912415
export BLOCK_SCANNER_BLOCK_SCAN_PROCESSORS=2
export BLOCK_SCANNER_BLOCK_HEIGHT_DISCOVER_BACK_OFF=1s
export STATE_CHAIN_CHAIN_ID=STATECHAIN
export STATE_CHAIN_CHAIN_HOST=localhost:1317
export STATE_CHAIN_SIGNER_NAME=signer name
export STATE_CHAIN_SIGNER_PASSWD=signer password
```


### Development
Setup a local server
```bash
make install
make start
```

### Test
Run tests
```bash
make test
```

To run test live when you change a file, use...
```
go get -u github.com/mitranim/gow
make test-watch
```
