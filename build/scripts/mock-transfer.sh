#!/bin/sh
# ./mock-bond.bash <mock binance IP address> <BNB Address> <BNB Address> <coin amount> <memo>
# ./mock-bond.bash 127.0.0.1 bnbXYXYX bnbZZZZZ 5 BNB mymemo

set -ex

if [ -z $1 ]; then
    echo "Missing mock binance address (address:port)"
    exit 1
fi

if [ -z $2 ]; then
    echo "Missing bnb from address argument"
    exit 1
fi

if [ -z $3 ]; then
    echo "Missing bnb to address argument"
    exit 1
fi

if [ -z $4 ]; then
    echo "Missing coin amount argument"
    exit 1
fi

if [ -z $5 ]; then
    echo "Missing coin asset argument"
    exit 1
fi

if [ -z $6 ]; then
    echo "Missing memo argument"
    exit 1
fi


# POOL_ADDRESS=$(curl -s $1:1317/thorchain/pool_addresses | jq -r ".current[0].address")

curl -vvv -s -X POST -d "{
  \"from\": \"$2\",
  \"to\": \"$3\",
  \"coins\":[
      {\"denom\": \"$5\", \"amount\": $4}
  ],
  \"memo\": \"$6\"
}" $1:26660/broadcast/easy
