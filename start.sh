#!/bin/bash

while true; do

  appHome="$HOME/go/src/github.com"
  appPath="jpthor/cosmos-swap"

  cd ${appHome}/${appPath}
  make install
  ssd init local --chain-id sschain

  # sscli keys add jack
  # sscli keys add alice

  ssd add-genesis-account $(sscli keys show jack -a) 1000atom,100000000stake
  ssd add-genesis-account $(sscli keys show alice -a) 1000atom,100000000stake

  sscli config chain-id sschain
  sscli config output json
  sscli config indent true
  sscli config trust-node true

  ssd gentx --name jack
  ssd collect-gentxs
  ssd validate-genesis

  ssd start & sscli rest-server --chain-id sschain --trust-node && fg

  break

done
