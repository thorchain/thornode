# adds an account node into the genesis file
add_node_account () {
    NODE_ADDRESS=$1
    VALIDATOR=$2
    OBSERVER_ADDRESS=$3
    VERSION=$4
    POOL_ADDRESS=$5
    jq --arg VERSION "$VERSION" --arg POOL_ADDRESS "$POOL_ADDRESS" --arg VALIDATOR "$VALIDATOR" --arg NODE_ADDRESS "$NODE_ADDRESS" --arg OBSERVER_ADDRESS "$OBSERVER_ADDRESS" '.app_state.swapservice.node_accounts[0] = {"node_address": $NODE_ADDRESS, "version": $VERSION, "status":"active","bond_address":$POOL_ADDRESS,"accounts":{"bnb_signer_acc": $POOL_ADDRESS, "bepv_validator_acc": $VALIDATOR, "bep_observer_acc": $OBSERVER_ADDRESS}}' <~/.thord/config/genesis.json >/tmp/genesis.json
    mv /tmp/genesis.json ~/.thord/config/genesis.json
}

# Adds a pool address into the genesis file
add_pool_address () {
    POOL_ADDRESS=$1
    SEQNO=$2
    jq --arg SEQNO "$SEQNO" --arg POOL_ADDRESS "$POOL_ADDRESS" '.app_state.swapservice.pool_addresses.rotate_at="28800" | .app_state.swapservice.pool_addresses.rotate_window_open_at="27800" | .app_state.swapservice.pool_addresses.current[0] = {"chain":"BNB","seq_no":$SEQNO,"pub_key":$POOL_ADDRESS}' <~/.thord/config/genesis.json >/tmp/genesis.json
    mv /tmp/genesis.json ~/.thord/config/genesis.json
}

# inits a statechain with a comman separate list of usernames
init_chain () {
    export IFS=","

    thord init local --chain-id statechain

    for user in $1; do # iterate over our list of comma separated users "alice,jack"
        echo "password" | thorcli keys add $user
        thord add-genesis-account $(thorcli keys show $user -a) 1000thor
    done

    thorcli config chain-id statechain
    thorcli config output json
    thorcli config indent true
    thorcli config trust-node true
}
