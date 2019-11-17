#!/bin/sh
set -ex

. $(dirname "$0")/core.sh

SIGNER_NAME="${SIGNER_NAME:=statechain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"
NODES="${NODES:=1}"
SEED="${SEED:=thor-daemon}" # the hostname of the master node
ROTATE_BLOCK_HEIGHT="${ROTATE_BLOCK_HEIGHT:=5}" # how often the pools in statechain should rotate

# find or generate our BNB address
gen_bnb_address
ADDRESS=$(cat ~/.signer/address.txt)

# create statechain user
echo $SIGNER_PASSWD | thorcli keys add $SIGNER_NAME

VALIDATOR=$(thord tendermint show-validator)
NODE_ADDRESS=$(thorcli keys show statechain -a)
NODE_PUB_KEY=$(thorcli keys show statechain -p)
VERSION=$(fetch_version)

if [ "$SEED" = "$(hostname)" ]; then
    echo "I AM THE SEED NODE"
    thord tendermint show-node-id > /tmp/shared/node.txt
fi

# write node account data to json file in shared directory
echo "$NODE_ADDRESS $VALIDATOR $NODE_PUB_KEY $VERSION $ADDRESS" > /tmp/shared/node_$NODE_ADDRESS.json

# write rotate block height as config file
if [ "$ROTATE_BLOCK_HEIGHT" != "0" ]; then
    echo "$KEY $VALUE $NODE_ADDRESS" > /tmp/shared/config_$NODE_ADDRESS.json
fi

# wait until we have the correct number of nodes in our directory before continuing
while [ "$(ls -1 /tmp/shared/node_*.json | wc -l | tr -d '[:space:]')" != "$NODES" ]; do
    sleep 1
done

# wait for TSS keysign agent to become available
$(dirname "$0")/wait-for-tss-keygen.sh $TSSKEYSIGN

KEYCLIENT="/usr/bin/keygenclient --name $SIGNER_NAME --password $SIGNER_PASSWD -u http://$TSSKEYSIGN:8322/keygen"
for f in /tmp/shared/node_*.json; do
  KEYCLIENT="$KEYCLIENT --pubkey $(cat $f | awk '{print $3}')"
done
sh -c "$KEYCLIENT > /tmp/keygenclient.output"

PUBKEY=$(cat /tmp/keygenclient.output | tail -2 | head -1)
POOL_ADDRESS=$(cat /tmp/keygenclient.output | tail -1)

if [ "$SEED" = "$(hostname)" ]; then
    if [ ! -f ~/.thord/config/genesis.json ]; then
        init_chain $SIGNER_NAME $SIGNER_PASSWD

        # add node accounts to genesis file
        for f in /tmp/shared/node_*.json; do 
            add_node_account $(cat $f | awk '{print $1}') $(cat $f | awk '{print $2}') $(cat $f | awk '{print $3}') $(cat $f | awk '{print $4}') $(cat $f | awk '{print $5}')
        done

        add_pool_address $PUBKEY "0"

        for f in /tmp/shared/config_*.json; do
          add_admin_config $(cat $f | awk '{print $1}') $(cat $f | awk '{print $2}') $(cat $f | awk '{print $3}')
        done

        cat ~/.thord/config/genesis.json
        thord validate-genesis
    fi
fi

# setup peer connection
if [ "$SEED" != "$(hostname)" ]; then
    if [ ! -f ~/.thord/config/genesis.json ]; then
        echo "I AM NOT THE SEED"
        
        init_chain
        fetch_genesis $SEED
        NODE_ID=$(fetch_node_id $SEED)
        echo "NODE ID: $NODE_ID"
        peer_list $NODE_ID $SEED

        cat ~/.thord/config/genesis.json
    fi
fi

echo "POOL ADDRESS: $POOL_ADDRESS"

exec "$@"
