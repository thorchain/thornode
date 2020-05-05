#!/bin/sh

SIGNER_NAME="${SIGNER_NAME:=thorchain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"
MASTER_ADDR="${ETH_MASTER_ADDR:=0x3fd2d4ce97b082d4bce3f9fee2a3d60668d2f473}"

cd "opt"
mkdir "data"
echo "{
    \"config\": {
        \"chainId\": 15,
        \"homesteadBlock\": 0,
    	\"eip150Block\": 0,
        \"eip155Block\": 0,
        \"eip158Block\": 0
    },
    \"difficulty\": \"1\",
    \"gasLimit\": \"21000000\",
    \"alloc\": {
        \"3fd2d4ce97b082d4bce3f9fee2a3d60668d2f473\": { \"balance\": \"2000000000000000000000\" },
        \"970e8128ab834e8eac17ab8e3812f010678cf791\": { \"balance\": \"0\" },
        \"f6da288748ec4c77642f6c5543717539b3ae001b\": { \"balance\": \"0\" },
        \"fabb9cc6ec839b1214bb11c53377a56a6ed81762\": { \"balance\": \"0\" },
        \"1f30a82340f08177aba70e6f48054917c74d7d38\": { \"balance\": \"0\" }
    }
}" >> "genesis.json"

geth --datadir "data" init "genesis.json"
geth --etherbase 0x3fd2d4ce97b082d4bce3f9fee2a3d60668d2f473 --verbosity 5 --networkid 15 --datadir "data" -mine --miner.threads 1 -rpc --rpcaddr 0.0.0.0 --rpcport 8545 -nousb --rpcapi "eth,net,web3,miner,personal,admin,ssh,txpool,debug" --rpccorsdomain "*" -nodiscover --rpcvhosts="*"
