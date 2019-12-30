Thornode Docker
===============

This directory contains helper commands and docker image to run a complete
thornode.

When using "run" commands it runs the service with the data available in
`~/.thornode`. When using the "reset" commands, its the same thing but deletes
all the data before starting the service so we start with a fresh instance.

#### Environments
 * MockNet - is a testnet but using a mock binance server instead of the
   binance testnet
 * TestNet - is a testnet
 * MainNet - is a mainnet

### Standalone Node
To run a single isolated node...
```bash
make run-mocknet-standalone
```

### Genesis Ceremony
To run a 4 node setup conducting a genesis ceremony...

```bash
make run-mocknet-genesis
```

### Run Validator
To run a single node to join an already existing blockchain...

```bash
PEER=<SEED IP ADDRESS> make run-validator
```
