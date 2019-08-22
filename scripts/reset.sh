#!/bin/sh

set -x
set -e

while true; do

  ssd init local --chain-id sschain

  ssd add-genesis-account $(sscli keys show jack -a) 1000rune,100000000stake
  ssd add-genesis-account $(sscli keys show alice -a) 1000rune,100000000stake

  sscli config chain-id sschain
  sscli config output json
  sscli config indent true
  sscli config trust-node true

  echo "password" | ssd gentx --name jack
  ssd collect-gentxs

  # add jack as a trusted account
  cat ~/.ssd/config/genesis.json | jq ".app_state.swapservice.trust_accounts[0] = {\"name\":\"Jack\", \"bnb_address\": \"bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlYYY\", \"rune_address\": \"$(sscli keys show jack -a)\"}" > /tmp/genesis.json
  mv /tmp/genesis.json ~/.ssd/config/genesis.json
  cat ~/.ssd/config/genesis.json | jq ".app_state.swapservice.admin_configs[0] = {\"key\":\"PoolAddress\", \"value\": \"bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6\", \"address\": \"bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlYYY\", \"rune_address\": \"$(sscli keys show jack -a)\"}" > /tmp/genesis.json
  mv /tmp/genesis.json ~/.ssd/config/genesis.json

  ssd validate-genesis

  break

done
