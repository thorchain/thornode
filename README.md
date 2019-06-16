Cosmos Swap
===========

Swap any two tokens on the [Cosmos network](cosmos.network)

An account can stake coins into a pool that is composed of `ATOM` and any
other single token. Once a pool has been created by an account(s) staking
their coins, people can than swap/trade within that pool to gain one token
over another.

How many tokens an individual gets from another token is relative to how many
tokens of each exist in the pool. Say you have a pool of two tokens, `ATOM`
and `CANYA`. There are 20 `CANYA` tokens, and 50 `ATOM` tokens. If you swap 10
`CANYA` tokens, we add those to the pool (for a total of 30), see that our
addition makes up for 1/3rd of the present tokens, which means we get 1/3rd of
the `ATOM` tokens, which is 16.6666666667. This ensures that we never run out
of tokens in a pool from swapping.

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
