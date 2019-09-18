## 
Files in this folder will allow you to run two nodes in docker composer
## Build docker image
```shell script
docker build  -t thorchain/statechainnode .

```

## setup
```shell script
export SIGNER_PASSWD = "passwd"
export POOL_ADDRESS = "pool address"
export $TRUSTED_BNB_ADDRESS = ""
./setup.sh
```

## run it
```shell script
docker-compose up
```

After this , you will have two nodes running in docker composer , they will talk to each other