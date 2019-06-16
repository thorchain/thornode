Cosmos Swap
===========

Swap any two tokens on the [Cosmos network](cosmos.network)

An account can stake coins into a pool that is composed of `ATOM` and any
other single token. Once a pool has been created by an account(s) staking
their coins, people can than swap/trade within that pool to gain one token
over another.

### Development
Setup a local server
```bash
make start
```

See [test](https://github.com/jpthor/cosmos-swap/blob/master/scripts/test.sh) and [demo](https://github.com/jpthor/cosmos-swap/blob/master/scripts/demo.sh) scripts for how to use the API

### Test
Run tests
```bash
make test
```

## TODO

 * Stakers gain fees when people swap tokens
 * Use real accAddresses instead of hacky made up accounts stored without this
   module's KVstore.
