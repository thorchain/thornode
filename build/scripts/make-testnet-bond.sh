#!/usr/bin/env bash

set -ex

echo "about to start making testnet bond"

INPUT=input.txt
export BOND_WALLET=bond-wallet
NODE_ACCOUNT=$(docker exec -it thor-daemon thorcli keys show thorchain -a | sed -e 's/[^A-Za-z0-9._-]//g')
BOND_MEMO=BOND:$NODE_ACCOUNT
FAUCET_WALLET=faucet
CHAIN_ID=Binance-Chain-Nile
TENDERMINT_NODE="data-seed-pre-2-s1.binance.org:80"
FUND_MEMO="fund validator"
BOND_AMOUNT=100000000:RUNE-A1F
GAS_FEE=37500

################################
# restore bond wallet locally
################################
MNEMONIC=$(docker exec thor-daemon cat /root/.bond/mnemonic.txt)

# first delete the wallet if it does exist
BOND_ADDRESS=$(tbnbcli keys list --output json | jq '.[] | select(.name | contains(env.BOND_WALLET))'.address | sed -e 's/"//g')
if [ -z "${BOND_ADDRESS}" ]; then
    echo "no need to delete wallet"
else
    echo $BOND_WALLET_PASSWORD| tbnbcli keys delete bond-wallet 2>/dev/null
fi

cat <<EOF > input.txt
${BOND_WALLET_PASSWORD}
${BOND_WALLET_PASSWORD}
${MNEMONIC}
EOF


while read -r password password_confirmation mnemonic
do
        tbnbcli keys add $BOND_WALLET --recover 2>/dev/null
done < $INPUT

BOND_ADDRESS=$(tbnbcli keys list --output json | jq '.[] | select(.name | contains(env.BOND_WALLET))'.address | sed -e 's/"//g')

# fund bond wallet
if [ ! -z "${FAUCET_PASSWORD}" ]; then
    echo $FAUCET_PASSWORD | tbnbcli token multi-send \
                                --from $FAUCET_WALLET \
                                --chain-id=$CHAIN_ID \
                                --node=$TENDERMINT_NODE \
                                --memo=$FUND_MEMO \
                                --transfers "[{\"to\":\"$BOND_ADDRESS\",\"amount\":\"$BOND_AMOUNT\"}, {\"to\":\"$BOND_ADDRESS\",\"amount\":\"$GAS_FEE:BNB\"}]" --json
else
    echo "please export your FAUCET_PASSWORD"
    exit 1
fi

# make bond
IP=$(docker-machine ip $DOCKER_SERVER)
ASGARD=$(curl -s http://${PEER}:1317/thorchain/pool_addresses | jq '.current[]'.address | sed -e 's/"//g')

echo $PASSWORD | tbnbcli send \
                    --from $BOND_WALLET \
                    --to $ASGARD \
                    --amount "$BOND_AMOUNT" \
                    --chain-id=$CHAIN_ID \
                    --node=$TENDERMINT_NODE \
                    --memo $BOND_MEMO \
                    --json \

echo "just finished making testnet bond"

eval $(docker-machine env -u)
docker-machine ssh ${DOCKER_SERVER} touch /tmp/bonded

#############
# clean up ##
#############
rm -f $INPUT

# delete bond-wallet
echo $PASSWORD | tbnbcli keys delete $BOND_WALLET







