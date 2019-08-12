Statechain Integration Tests
============================

## Run within Docker
```
make test
```

## Run manually locally

## Setup
Clean your environment and setup...

```bash
make -C ../.. clean setup install
```

## Start Service

Once you have your chain setup, start the daemon in a terminal window.

```bash
make -C ../.. install reset start-daemon
```

Then also start the rest server in another terminal window
```bash
make -C ../.. start-rest
```

## Run Tests
Once you have start the service, run tests...

```bash
make test-internal
```

## Reset
Between each run, you must start from a clean state...

```bash
make -C ../.. reset
```
