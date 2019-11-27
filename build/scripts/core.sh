# adds an account node into the genesis file
add_node_account () {
    NODE_ADDRESS=$1
    VALIDATOR=$2
    NODE_PUB_KEY=$3
    VERSION=$4
    BOND_ADDRESS=$5
    POOL_PUB_KEY=$6
    jq --arg VERSION $VERSION --arg BOND_ADDRESS "$BOND_ADDRESS" --arg VALIDATOR "$VALIDATOR" --arg NODE_ADDRESS "$NODE_ADDRESS" --arg NODE_PUB_KEY "$NODE_PUB_KEY" '.app_state.thorchain.node_accounts += [{"node_address": $NODE_ADDRESS, "version": $VERSION, "status":"active", "active_block_height": "0", "bond_address":$BOND_ADDRESS,"validator_cons_pub_key":$VALIDATOR,"node_pub_keys":{"secp256k1":$NODE_PUB_KEY,"ed25519":$NODE_PUB_KEY}}]' <~/.thord/config/genesis.json >/tmp/genesis.json
    mv /tmp/genesis.json ~/.thord/config/genesis.json
}

add_last_event_id () {
    echo "Adding last event id $1"
    jq --arg ID $1 '.app_state.thorchain.last_event_id = $ID' ~/.thord/config/genesis.json > /tmp/genesis.json
    mv /tmp/genesis.json ~/.thord/config/genesis.json
}

# Adds a pool address into the genesis file
add_pool_address () {
    POOL_ADDRESS=$1
    POOL_PUB_KEY=$2
    SEQNO=$3
    jq --arg SEQNO "$SEQNO" --arg POOL_ADDRESS "$POOL_ADDRESS" --arg POOL_PUB_KEY "$POOL_PUB_KEY" '.app_state.thorchain.pool_addresses.rotate_at="28800" | .app_state.thorchain.pool_addresses.rotate_window_open_at="27800" | .app_state.thorchain.pool_addresses.current += [{"chain":"BNB","seq_no":$SEQNO, "address": $POOL_ADDRESS, "pub_key":$POOL_PUB_KEY}]' <~/.thord/config/genesis.json >/tmp/genesis.json
    mv /tmp/genesis.json ~/.thord/config/genesis.json
}

add_admin_config () {
    jq --arg NODE_ADDRESS "$3" --arg KEY "$1" --arg VALUE "$2" '.app_state.thorchain.admin_configs += [{"address": $NODE_ADDRESS ,"key":$KEY, "value":$VALUE}]' ~/.thord/config/genesis.json > /tmp/genesis.json
}

# inits a thorchain with a comman separate list of usernames
init_chain () {
    export IFS=","

    thord init local --chain-id thorchain
    thorcli keys list

    for user in $1; do # iterate over our list of comma separated users "alice,jack"
        # Only create if it doesn't exist!
        EXISTS=""
        for name in $(thorcli keys list --output json | jq -r '.[].name'); do
          if [ "$name" = "$user" ]; then
            EXISTS="true"
          fi
        done

        if [ -z "$EXISTS" ]; then
          echo "$2" | thorcli keys add $user
        fi

        thord add-genesis-account $(thorcli keys show $user -a) 1000thor
    done

    thorcli config chain-id thorchain
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
    thorcli query thorchain version --chain-id thorchain --trust-node --output json | jq -r .version
}
