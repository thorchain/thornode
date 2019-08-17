[![pipeline status](https://gitlab.com/thorchain/bepswap/observe/badges/master/pipeline.svg)](https://gitlab.com/thorchain/bepswap/observe/commits/master)
[![coverage report](https://gitlab.com/thorchain/bepswap/observe/badges/master/coverage.svg)](https://gitlab.com/thorchain/bepswap/observe/commits/master)

BEPSwap Observe
===============

Observe events that take place on Binance.

### Environment
Export the following environment variables (set based on your environment and/or the net being used [test/prod]):
```bash
export DEX_HOST=<DEX Hostname>
export POOL_ADDRESS=<Binance Address>
export REDIS_PASSWORD=""
export REDIS_URL=localhost:6379
export PORT=<HTTP Port>
export RPC_HOST=<RPC Hostname>
export PRIVATE_KEY=<Binance Private Key>
export CHAIN_HOST=<Statechain REST Server Address>
export RUNE_ADDRESS=<Rune Address>
export SIGNER_PASSWD=<Password>
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
