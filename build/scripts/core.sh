# adds an account node into the genesis file
add_node_account () {
    NODE_ADDRESS=$1
    VALIDATOR=$2
    OBSERVER_ADDRESS=$3
    VERSION=$4
    POOL_ADDRESS=$5
    jq --arg VERSION "$VERSION" --arg POOL_ADDRESS "$POOL_ADDRESS" --arg VALIDATOR "$VALIDATOR" --arg NODE_ADDRESS "$NODE_ADDRESS" --arg OBSERVER_ADDRESS "$OBSERVER_ADDRESS" '.app_state.swapservice.node_accounts += [{"node_address": $NODE_ADDRESS, "version": $VERSION, "status":"active","bond_address":$POOL_ADDRESS,"accounts":{"bnb_signer_acc": $POOL_ADDRESS, "bepv_validator_acc": $VALIDATOR, "bep_observer_acc": $OBSERVER_ADDRESS}}]' <~/.thord/config/genesis.json >/tmp/genesis.json
    mv /tmp/genesis.json ~/.thord/config/genesis.json
}

# Adds a pool address into the genesis file
add_pool_address () {
    POOL_ADDRESS=$1
    SEQNO=$2
    jq --arg SEQNO "$SEQNO" --arg POOL_ADDRESS "$POOL_ADDRESS" '.app_state.swapservice.pool_addresses.rotate_at="28800" | .app_state.swapservice.pool_addresses.rotate_window_open_at="27800" | .app_state.swapservice.pool_addresses.current += [{"chain":"BNB","seq_no":$SEQNO,"pub_key":$POOL_ADDRESS}]' <~/.thord/config/genesis.json >/tmp/genesis.json
    mv /tmp/genesis.json ~/.thord/config/genesis.json
}

add_admin_config () {
    jq --arg NODE_ADDRESS "$3" --arg KEY "$1" --arg VALUE "$2" '.app_state.swapservice.admin_configs += [{"address": $NODE_ADDRESS ,"key":$KEY, "value":$VALUE}]' ~/.thord/config/genesis.json > /tmp/genesis.json
}

# inits a statechain with a comman separate list of usernames
init_chain () {
    export IFS=","

    thord init local --chain-id statechain

    for user in $1; do # iterate over our list of comma separated users "alice,jack"
        echo "$2" | thorcli keys add $user
        thord add-genesis-account $(thorcli keys show $user -a) 1000thor
    done

    thorcli config chain-id statechain
    thorcli config output json
    thorcli config indent true
    thorcli config trust-node true
}

peer_list () {
    PEER="$1@$2:26656"
    ADDR='addr_book_strict = true'
    ADDR_STRICT_FALSE='addr_book_strict = false'
    PEERSISTENT_PEER_TARGET='persistent_peers = ""'

    sed -i -e "s/$ADDR/$ADDR_STRICT_FALSE/g" ~/.thord/config/config.toml
    sed -i -e "s/$PEERSISTENT_PEER_TARGET/persistent_peers = \"$PEER\"/g" ~/.thord/config/config.toml
}

gen_bnb_address () {
    if [ ! -f ~/.signer/private_key.txt ]; then
        echo "GENERATING BNB ADDRESSES"
        # because the generate command can get API rate limited, we may need to retry
        n=0
        until [ $n -ge 60 ]; do
            generate > /tmp/bnb && break
            n=$[$n+1]
            sleep 1
        done
        ADDRESS=$(cat /tmp/bnb | grep MASTER= | awk -F= '{print $NF}')
        echo $ADDRESS > ~/.signer/address.txt
        BINANCE_PRIVATE_KEY=$(cat /tmp/bnb | grep MASTER_KEY= | awk -F= '{print $NF}')
        echo $BINANCE_PRIVATE_KEY > ~/.signer/private_key.txt
        PUBKEY=$(cat /tmp/bnb | grep MASTER_PUBKEY= | awk -F= '{print $NF}')
        echo $PUBKEY > ~/.signer/pubkey.txt
    fi
}

fetch_genesis () {
    until curl -s "$1:26657" > /dev/null; do
        sleep 3
    done
    curl -s $1:26657/genesis | jq .result.genesis > ~/.thord/config/genesis.json
    thord validate-genesis
    cat ~/.thord/config/genesis.json
}

fetch_node_id () {
    until curl -s "$1:26657" > /dev/null; do
        sleep 3
    done
    curl -s $1:26657/status | jq -r .result.node_info.id
}

fetch_version () {
    thorcli query swapservice version --chain-id statechain --trust-node --output json | jq -r .version
}
