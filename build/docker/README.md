Thornode Docker
===============

This directory contains helper commands and docker image to run a complete
thornode.

### Standalone Node
To run a single isolated node...
```bash
make run-standalone
```

### Genesis Ceremony
To run a 4 node setup conducting a genesis ceremony...

```bash
make run-genesis
```

### Run Validator
To run a single node to join an already existing blockchain...

```bash
PEER=<SEED IP ADDRESS> make run-validator
```
