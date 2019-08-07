#!/bin/bash

set -x
set -e

while true; do

  make install
  ssd init local --chain-id sschain

  ssd add-genesis-account $(sscli keys show jack -a) 1000rune,100000000stake
  ssd add-genesis-account $(sscli keys show alice -a) 1000rune,100000000stake

  sscli config chain-id sschain
  sscli config output json
  sscli config indent true
  sscli config trust-node true

  echo "password" | ssd gentx --name jack
  ssd collect-gentxs
  ssd validate-genesis

  break

done
