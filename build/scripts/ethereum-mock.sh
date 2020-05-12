#!/bin/sh

BLOCK_PERIOD=1

geth --dev --dev.period $BLOCK_PERIOD --verbosity 5 --networkid 15 --datadir "data" -mine --miner.threads 1 -rpc --rpcaddr 0.0.0.0 --rpcport 8545 -nousb --rpcapi "eth,net,web3,miner,personal,admin,ssh,txpool,debug" --rpccorsdomain "*" -nodiscover --rpcvhosts="*"
