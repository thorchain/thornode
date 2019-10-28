Multinode Thorchain
===================

Files in this folder will allow you to run fou4 nodes in docker composer

## Build image
```bash
make build
```

## Run it
```bash
make up NODE_ID=xxx MASTER_API=yyy GENESIS_URL=zzz
```

## Stop it
```bash
make down
```

After this , you will have four nodes running in docker composer , they will talk to each other

# Single node

If you wnat to run a single node
```bash
make single
```
