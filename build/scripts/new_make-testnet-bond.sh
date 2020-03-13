#!/usr/bin/env bash

echo "about to start making testnet bond"

BOND_FILE=bond.txt
FAUCET_FILE=faucet.txt
export BOND_WALLET=${THORNODE_ENV}-bond-wallet
NODE_ACCOUNT=$(docker exec thor-daemon thorcli keys show thorchain -a)
BOND_MEMO=BOND:$NODE_ACCOUNT
FAUCET_WALLET=faucet
CHAIN_ID=Binance-Chain-Nile
TENDERMINT_NODE="18.208.208.172:26657"
FUND_MEMO="fund validator"
BOND_AMOUNT=100000000:RUNE-A1F
GAS_FEE=37500

####################################
# restore faucet wallet only on CI
####################################
if [ ! -z "${CI}" ]; then
wget https://media.githubusercontent.com/media/binance-chain/node-binary/master/cli/testnet/0.6.2/linux/tbnbcli
chmod +x tbnbcli
mv tbnbcli /usr/local/bin/.

cat <<EOF > ${FAUCET_FILE}
${FAUCET_PASSWORD}
${FAUCET_PASSWORD}
${FAUCET_MNEMONIC}
EOF

while read -r password password_confirmation mnemonic
do
        tbnbcli keys add $FAUCET_WALLET --recover 2>/dev/null
done < $FAUCET_FILE
fi

################################
# restore bond wallet locally
################################
MNEMONIC=$(docker exec thor-daemon cat /root/.bond/mnemonic.txt)

# first delete the wallet if it does exist
BOND_ADDRESS=$(tbnbcli keys list --output json | jq '.[] | select(.name | contains(env.BOND_WALLET))'.address | sed -e 's/"//g')
if [ -z "${BOND_ADDRESS}" ]; then
    echo "no need to delete locally recovered bond wallet"
else
    if [ ! -z "${BOND_WALLET_PASSWORD}" ]; then
        echo $BOND_WALLET_PASSWORD| tbnbcli keys delete bond-wallet 2>/dev/null
    else
        echo "please export your BOND_WALLET_PASSWORD"
        exit 1
    fi
fi

cat <<EOF > $BOND_FILE
${BOND_WALLET_PASSWORD}
${BOND_WALLET_PASSWORD}
${MNEMONIC}
EOF

while read -r password password_confirmation mnemonic
do
        tbnbcli keys add $BOND_WALLET --recover 2>/dev/null
done < $BOND_FILE
BOND_ADDRESS=$(tbnbcli keys list --output json | jq '.[] | select(.name | contains(env.BOND_WALLET))'.address | sed -e 's/"//g')

##############################
# fund bond wallet from faucet
##############################
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

######################
# make bond to Asgard
#####################
IP=$(docker-machine ip $DOCKER_SERVER)
ASGARD=$(curl -s http://${PEER}:1317/thorchain/pool_addresses | jq '.current[]'.address | sed -e 's/"//g')
echo ${BOND_WALLET_PASSWORD} | tbnbcli send \
                                --from $BOND_WALLET \
                                --to $ASGARD \
                                --amount "$BOND_AMOUNT" \
                                --chain-id=$CHAIN_ID \
                                --node=$TENDERMINT_NODE \
                                --memo $BOND_MEMO \
                                --json \

echo "just finished making bond"

#############
# clean up ##
#############
if [ ! -z "${CI}" ]; then
    echo "no need to unset docker variables"
else
    eval $(docker-machine env -u)
fi
rm -f $BOND_FILE $FAUCET_FILE

# delete local bond-wallet
echo ${BOND_WALLET_PASSWORD} | tbnbcli keys delete $BOND_WALLET

# delete local faucet-wallet
if [ ! -z "${CI}" ]; then
    echo ${FAUCET_PASSWORD} | tbnbcli keys delete $FAUCET_WALLET
fi