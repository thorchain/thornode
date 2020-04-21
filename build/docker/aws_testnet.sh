#!/bin/sh

export USER=$(hostname -s)
export DOCKER_SERVER="${THORNODE_ENV}-${THORNODE_SERVICE}-$(date +%s)" # must be unique
export SEED_ENDPOINT=https://${THORNODE_ENV}-seed.thorchain.info
export BUCKET_NAME=thorchain.info
export SEED_BUCKET="${THORNODE_ENV}-seed.${BUCKET_NAME}"
export S3_FILE="bonded_nodes.json"
export DATE=$(date +%F)
export BOOTSTRAP="/opt/${THORNODE_ENV}/${THORNODE_SERVICE}-bootstrap"
export FAUCET_FILE=/tmp/faucet.txt
export BOND_FILE=/tmp/bond.txt
export BOND_WALLET=bond-wallet
export RESTORE_FILE=/opt/mnemonic_phrase.txt
export FAUCET_WALLET=faucet
export SSH_PUB_KEY=gitlab-ci.pub
export SSH_PRIV_KEY=gitlab-ci
export FUND_MEMO="fund-bond-wallet"
export TENDERMINT_NODE="testnet-binance.thorchain.info:26657"
export CHAIN_ID=Binance-Chain-Nile
export SIGNER_NAME=thorchain
export BOND_AMOUNT=100000000:RUNE-A1F
export GAS_FEE=75001
export DISK_SIZE=${DISK_SIZE:=100}
export AWS_INSTANCE_TYPE=${AWS_INSTANCE_TYPE:=c5.2xlarge}

cleanup () {
    echo "performing cleanup"
    if [ ! -z "${CI}" ]; then
        echo "no need to unset docker variables"
        export AWS_REGION=${AWS_CI_REGION}
        export USER=$CI_PIPELINE_ID
    else
        eval $(docker-machine env -u)
    fi
    docker-machine rm -f $1 > /dev/null 2>&1
    sleep $2
}

##########################################################
# ENSURE DOCKER-MACHINE AND DOCKER_COMPOSE ARE INSTALLED #
##########################################################
which docker-machine
if [ $? != 0 ]; then
    echo "installing docker-machine"
    base=https://github.com/docker/machine/releases/download/v0.16.2
    curl -sL $base/docker-machine-$(uname -s)-$(uname -m) >/tmp/docker-machine &&  install /tmp/docker-machine /usr/local/bin/docker-machine
fi

which docker-compose
if [ $? != 0 ]; then
    echo "installing docker-compose"
    curl -L "https://github.com/docker/compose/releases/download/1.25.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
fi

####################
# START THE STACK  #
####################
create_server() {
    cleanup ${DOCKER_SERVER} 20
    if [ ! -z "${CI}" ]; then
        aws s3 cp s3://${BUCKET_NAME}/$SSH_PUB_KEY /tmp/.
        aws s3 cp s3://${BUCKET_NAME}/$SSH_PRIV_KEY /tmp/.
        chmod 0600 /tmp/$SSH_PRIV_KEY
	    echo "creating server node on AWS"
	    docker-machine create --driver amazonec2 \
            --amazonec2-vpc-id=${AWS_VPC_ID} \
            --amazonec2-region ${AWS_REGION} \
            --amazonec2-instance-type ${AWS_INSTANCE_TYPE} \
            --amazonec2-root-size ${DISK_SIZE} \
            --amazonec2-ssh-keypath /tmp/$SSH_PRIV_KEY \
            --amazonec2-userdata ./${THORNODE_ENV}/ec2-userdata.sh \
            --amazonec2-tags Environment,${THORNODE_ENV} \
            --amazonec2-iam-instance-profile ${THORNODE_ENV}-secrets \
            ${DOCKER_SERVER}
    else
        docker-machine create --driver amazonec2 \
            --amazonec2-vpc-id=${AWS_VPC_ID} \
            --amazonec2-region ${AWS_REGION} \
            --amazonec2-instance-type ${AWS_INSTANCE_TYPE} \
            --amazonec2-root-size ${DISK_SIZE} \
            --amazonec2-userdata ./${THORNODE_ENV}/ec2-userdata.sh \
            --amazonec2-tags Environment,${THORNODE_ENV} \
            --amazonec2-iam-instance-profile ${THORNODE_ENV}-secrets \
            ${DOCKER_SERVER}
    fi
    if [ $? != 0 ]; then
        echo "server could not be created"
        exit 1
    fi
}

start_the_stack () {
    echo "waiting for server to be ready"
    sleep 60 # give setup script enough time install required packages
    eval $(docker-machine env ${DOCKER_SERVER} --shell bash)
    docker-machine ssh ${DOCKER_SERVER} sudo bash $BOOTSTRAP
}

# checks if there is are any standby/ready nodes, exits if there are
check_for_slots() {
    echo "PEER: $PEER"
    curl -s $PEER:1317/thorchain/nodeaccounts
    if [ $? -gt 0 ]; then
      echo "unable to detect if we should add a new node"
      exit 0
    fi
    standby=$(curl -s $PEER:1317/thorchain/nodeaccounts | jq -r '.[] | select(.status | inside("standby ready")) | select(.bond | contains("100000000")) | .status')
    if [[ $(echo $standby | sed '/^$/d' | wc -l) -gt 0 ]]; then
        echo "A node is already waiting to be churned in.... exiting"
        exit 0
    fi
}

verify_stack () {
    if [ "${THORNODE_SERVICE}" != binance ]; then
        export PUB_KEY=$(docker exec thor-daemon thorcli keys show thorchain --pubkey)
        export VALIDATOR=$(docker exec thor-daemon thord tendermint show-validator)
    fi
    echo "allow sufficient time for stack to be up"
    sleep 60
    echo "performing healthchecks"
    export IP=$(docker-machine ip ${DOCKER_SERVER})
    if [ "${THORNODE_SERVICE}" == binance ]; then
        HEALTHCHECK_CMD=$(nc -z $IP 8080)
        if  [ $? == 0 ]; then
	        echo "HEALTHCHECK PASSED"
	    else
	        echo "HEALTHCHECK FAILED"
	        exit 1
        fi
    else
        HEALTHCHECK_URL="http://${IP}:8080/v1/thorchain/pool_addresses"
        HEALTHCHECK_CMD=$(curl -s -o /dev/null -w "%{http_code}" ${HEALTHCHECK_URL})
        if  [ "${HEALTHCHECK_CMD}" == 200 ]; then
	        echo "HEALTHCHECK PASSED"
        else
	        echo "HEALTHCHECK FAILED"
	        exit 1
        fi
    fi
}

churn () {
echo "churning"

eval $(docker-machine env ${DOCKER_SERVER} --shell bash)

cat <<EOD > ${FAUCET_FILE}
${FAUCET_PASSWORD}
${FAUCET_PASSWORD}
${FAUCET_MNEMONIC}
EOD

echo "restore faucet wallet"
while read -r password password_confirmation mnemonic
do
        tbnbcli keys add $FAUCET_WALLET --recover 2>/dev/null
done < $FAUCET_FILE

echo "restore bond wallet locally"
MNEMONIC=$(docker exec thor-daemon cat /root/.bond/mnemonic.txt)

# first delete the wallet if it does exist
echo $BOND_WALLET_PASSWORD| tbnbcli keys delete bond-wallet 2>/dev/null

cat <<EOF > $BOND_FILE
${BOND_WALLET_PASSWORD}
${BOND_WALLET_PASSWORD}
${MNEMONIC}
EOF

cat <<EOF > $RESTORE_FILE
${MNEMONIC}
EOF
sudo chown root $RESTORE_FILE && sudo chmod 400 $RESTORE_FILE

while read -r password password_confirmation mnemonic
do
        tbnbcli keys add $BOND_WALLET --recover 2>/dev/null
done < $BOND_FILE

export BOND_ADDRESS=$(tbnbcli keys list --output json | jq '.[] | select(.name | contains(env.BOND_WALLET))'.address | sed -e 's/"//g')

sleep 15
echo "fund bond wallet from faucet"
if [ ! -z "${FAUCET_PASSWORD}" ]; then
    echo $FAUCET_PASSWORD | tbnbcli token multi-send \
                                --from $FAUCET_WALLET \
                                --chain-id=$CHAIN_ID \
                                --node=$TENDERMINT_NODE \
                                --memo=$FUND_MEMO \
                                --transfers "[{\"to\":\"$BOND_ADDRESS\",\"amount\":\"$BOND_AMOUNT\"}, {\"to\":\"$BOND_ADDRESS\",\"amount\":\"$GAS_FEE:BNB\"}]" --json
else
    echo "please supply your FAUCET_PASSWORD"
    exit 1
fi

sleep 15
echo "make bond to Asgard"

# fetch list of peers
export PEERS=$(curl -sL testnet-seed.thorchain.info/node_ip_list.json | jq -r '.[]')

# find PEER with highest block height
rm -f /tmp/peers.txt
for ip in $PEERS; do
    echo "$(curl -s http://${ip}:26657/status | jq -r '.result.sync_info.latest_block_height') $ip" >> /tmp/peers.txt
done
export PEER=$(cat /tmp/peers.txt | sort -n -r | head -n 1 | awk '{print $NF}')

NODE_ACCOUNT=$(docker exec thor-daemon thorcli keys show thorchain -a)
BOND_MEMO=BOND:$NODE_ACCOUNT

ASGARD=$(curl -s http://${PEER}:1317/thorchain/pool_addresses | jq -r '.current[0].address')

echo ${BOND_WALLET_PASSWORD} | tbnbcli send \
                                --from $BOND_WALLET \
                                --to $ASGARD \
                                --amount "$BOND_AMOUNT" \
                                --chain-id=$CHAIN_ID \
                                --node=$TENDERMINT_NODE \
                                --memo $BOND_MEMO \
                                --json \

echo "just finished making bond"

echo "setting node keys"
sleep 30 # wait for thorchain to register the new node account
export SIGNER_PASSWD=$(aws secretsmanager get-secret-value --secret-id ${THORNODE_ENV}-signer-passwd --region $AWS_REGION  | jq -r .SecretString | awk -F'[:]' '{print $2}' | sed -e 's/}//' | sed -e 's/"//g')
docker exec thor-daemon ash -c "echo $SIGNER_PASSWD | thorcli tx thorchain set-node-keys $PUB_KEY $PUB_KEY $VALIDATOR --node tcp://$PEER:26657 --from $SIGNER_NAME --yes"

# delete local bond-wallet
# echo ${BOND_WALLET_PASSWORD} | tbnbcli keys delete $BOND_WALLET

# delete local faucet-wallet
if [ ! -z "${CI}" ]; then
    echo ${FAUCET_PASSWORD} | tbnbcli keys delete $FAUCET_WALLET
fi
}

final_cleanup () {
    echo "performing final cleanup"
    rm -f $FAUCET_FILE $BOND_FILE
    eval $(docker-machine env ${DOCKER_SERVER} --shell bash)
    echo "removing bootstrap script"
    docker-machine ssh ${DOCKER_SERVER} rm -f /opt/testnet/*-bootstrap

    echo "clean recovered wallets and disconnect from docker socket"
    if [ ! -z "${CI}" ]; then
        echo "no need to unset docker variables"
    else
        eval $(docker-machine env -u)
    fi
}

#########
# START #
#########
if [ ! -z "${AWS_VPC_ID}" ] && [ ! -z "${AWS_REGION}" ] && [ ! -z "${AWS_INSTANCE_TYPE}" ] && [ ! -z "${THORNODE_ENV}" ] && [ ! -z "${THORNODE_SERVICE}" ]; then
    if [ "${THORNODE_SERVICE}" == churn ]; then
        check_for_slots
    fi
    create_server
    start_the_stack
    verify_stack
    if [ "${THORNODE_SERVICE}" == churn ]; then
        churn
    fi
    #final_cleanup
else
    echo "you have not provided all the required environment variables"
	exit 1
fi
