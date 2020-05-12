#!/bin/sh
# ./mock-leave.bash <mock binance IP address> <BNB Address>
# ./mock-leave.bash 127.0.0.1 bnbXYXYX

set -ex

if [ -z $1 ]; then
    echo "Missing mock binance address (address:port)"
    exit 1
fi

if [ -z $2 ]; then
    echo "Missing bnb address argument"
    exit 1
fi

POOL_ADDRESS=$(curl -s $1:1317/thorchain/pool_addresses | jq -r ".current[0].address")

curl -vvv -s -X POST -d "[{
  \"from\": \"$2\",
  \"to\": \"$POOL_ADDRESS\",
  \"coins\":[
      {\"denom\": \"RUNE-A1F\", \"amount\": 1}
  ],
  \"memo\": \"LEAVE\"
}]" $1:26660/broadcast/easy
