#!/bin/sh
# ./refund.sh localhost:1317

# set -ex

if [ -z $1 ]; then
	echo "Missing thornode api address (address:port)"
	exit 1
fi

# NOTE: the tx ID doesn't matter at all, just no blank as it would trigger ragnarok
TX_ID=91311A8951EEFC1C84B09338738BC0154E488778CBB3CF47143B3B96D18230C2
POOLS=$(curl -s $1/thorchain/pools)
USERNAME=thorchain
PASSWORD=password

refund_pool () {
	asset=$(echo $1 | jq -r ".asset");
	address=$(echo $1 | jq -r ".pool_address");
	echo "Refunding pool $asset"
	echo $PASSWORD | thorcli tx thorchain set-end-pool $asset $address $TX_ID --from $USERNAME --chain-id thorchain  --home /root/.thorcli -y 2>&1
}

for pool in $(echo $POOLS | jq -c '.[]'); do
	asset=$(echo $pool | jq -r ".asset");
	if [ "$asset" = "BNB.BNB" ]; then
		bnb=$pool
		continue
	fi
	refund_pool $pool
done
refund_pool $bnb
