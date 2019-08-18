[![pipeline status](https://gitlab.com/thorchain/bepswap/observe/badges/master/pipeline.svg)](https://gitlab.com/thorchain/bepswap/observe/commits/master)
[![coverage report](https://gitlab.com/thorchain/bepswap/observe/badges/master/coverage.svg)](https://gitlab.com/thorchain/bepswap/observe/commits/master)

BEPSwap Observe
===============

Observe events that take place on Binance.

### Environment
Please create a config.json file
```json
{
  "pool_address": "my pool address",
  "rune_address": "my statechain address",
  "dex_host": "testnet-dex.binance.org",
  "rpc_host": "",
  "private_key": "my private key",
  "chain_host": "localhost:1317",
  "signer_passwd": "mysupersecretpassword",
  "observer_db_path": "data",
  "signer_db_path": "signerdata"
}
You could overwrite the above configs using environment variables as well 
Export the following environment variables (set based on your environment and/or the net being used [test/prod]):
```bash
export POOL_ADDRESS=<PoolAddress>
export RUNE_ADDRESS=<rune address>
export DEX_HOST=<DEX Hostname>
export RPC_HOST=<RPC Hostname>
export PRIVATE_KEY=<Binance Private Key>
export CHAIN_HOST=<Statechain REST Server Address>
export SIGNER_NAME=<SIGNER_NAME>
export SIGNER_PASSWD=<Password>
export LEVEL_DB_OBSERVER_PATH=<LEVEL_DB_OBSERVER_PATH>
export LEVEL_DB_SIGNER_PATH=<LEVEL_DB_SIGNER_PATH>
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
