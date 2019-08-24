[![pipeline status](https://gitlab.com/thorchain/bepswap/observe/badges/master/pipeline.svg)](https://gitlab.com/thorchain/bepswap/observe/commits/master)
[![coverage report](https://gitlab.com/thorchain/bepswap/observe/badges/master/coverage.svg)](https://gitlab.com/thorchain/bepswap/observe/commits/master)

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
