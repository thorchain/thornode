#!/bin/sh

cleanup () {
    echo "performing cleanup"
    if [ ! -z "${CI}" ]; then
        echo "no need to unset docker variables"
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
    cd ../../
    LOCAL_VOLUME=$(pwd)
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
            ${DOCKER_SERVER}
    else
        docker-machine create --driver amazonec2 \
            --amazonec2-vpc-id=${AWS_VPC_ID} \
            --amazonec2-region ${AWS_REGION} \
            --amazonec2-instance-type ${AWS_INSTANCE_TYPE} \
            --amazonec2-root-size ${DISK_SIZE} \
            ${DOCKER_SERVER}
    fi
    if [ $? != 0 ]; then
        echo "server could not be created"
        exit 1
    fi
    echo "mounting volumes"
    docker-machine ssh ${DOCKER_SERVER} sudo mkdir -p ${LOCAL_VOLUME}
    docker-machine ssh ${DOCKER_SERVER} sudo chmod 777 -R ${LOCAL_VOLUME}
    docker-machine scp -r ${LOCAL_VOLUME}/. ${DOCKER_SERVER}:${LOCAL_VOLUME}
}

start_the_stack () {
    cd build/docker/${THORNODE_ENV}
    echo "waiting for server to be ready"
    PROVISIONING_TIME=$1
    sleep ${PROVISIONING_TIME}
    export NET=${THORNODE_ENV}
    export TAG=${THORNODE_ENV}
    eval $(docker-machine env ${DOCKER_SERVER} --shell bash)
    PEER=$(curl -s http://thorchain.net.s3-website-us-east-1.amazonaws.com/net/)
    if [ ! -z "${CI}" ]; then
        export PEER=$PEER && make run-${THORNODE_ENV}-validator-ci
    elif [ ! -z "${NON_CI}" ]; then
        export PEER=$PEER && make run-${THORNODE_ENV}-validator
    else
        make run-${THORNODE_ENV}-standalone
    fi
    sleep 60
}

##################
# CHURN
##################
churn () {
    echo "starting churning"
    export FAUCET_PASSWORD=$FAUCET_PASSWORD && . ./../../scripts/make-testnet-bond.sh
    #################################################################
    # wait for bond transaction and for node account to be registered
    #################################################################
    eval $(docker-machine env ${DOCKER_SERVER} --shell bash)
    PUB_KEY=$(docker exec thor-daemon thorcli keys show thorchain --pubkey)
    VALIDATOR=$(docker exec thor-daemon thord tendermint show-validator)
    if [ ! -z "${CI}" ]; then
        export SIGNER_PASSWD=${CI_SIGNER_PASSWD}
    fi
    echo "setting node keys"
    sleep 60 # wait for thorchain to register the new node account
    docker exec thor-daemon ash -c "echo $SIGNER_PASSWD | thorcli tx thorchain set-node-keys $PUB_KEY $PUB_KEY $VALIDATOR --node tcp://$PEER:26657 --from $SIGNER_NAME --yes"
}

#####################
# VERIFY THE STACK  #
#####################
verify_the_stack () {
    echo "allow a few mins for docker services to come up"
    sleep 60
    docker ps -a
    echo "performing healthchecks"
    export IP=$(docker-machine ip ${DOCKER_SERVER})
    HEALTHCHECK_URL="http://${IP}:8080/v1/thorchain/pool_addresses"
    HEALTHCHECK_CMD=$(curl -s -o /dev/null -w "%{http_code}" ${HEALTHCHECK_URL})
    if  [ "${HEALTHCHECK_CMD}" == 200 ]; then
	    echo "HEALTHCHECK PASSED"
    else
	    echo "HEALTHCHECK FAILED"
	    exit 1
    fi
    if [ ! -z "${CHURN}" ]; then
        echo "starting churning"
        churn
    else
        update_ip
    fi
}

################################
# Register IP on S3 Web Endpoint
################################
update_ip() {
    echo $IP > /tmp/ip_address
    sed -e 's/"//g' -e "s/null//g" /tmp/ip_address > /tmp/${S3_FILE}
    aws s3 cp /tmp/${S3_FILE} s3://${BUCKET_NAME}/net/
}

#########
# START #
#########
THORNODE_ENV=testnet
BUCKET_NAME="thorchain.net"
S3_FILE="${THORNODE_ENV}.json"
USER=computer
DISK_SIZE=100
SSH_PUB_KEY=gitlab-ci.pub
SSH_PRIV_KEY=gitlab-ci

if [ $1 == "ci" ]; then
    export CI="true"
    export AWS_REGION=$AWS_CI_REGION
    export USER=$CI_JOB_ID
    export CHURN="true"
elif [ $1 == "churn" ]; then
    export CHURN="true"
    export NON_CI="true"
fi

export DOCKER_SERVER="${USER}-${THORNODE_ENV}$1"
if [ ! -z "${AWS_VPC_ID}" ] && [ ! -z "${AWS_REGION}" ] && [ ! -z "${AWS_INSTANCE_TYPE}" ]; then
    cleanup ${DOCKER_SERVER} 20
    create_server
    start_the_stack 60
    verify_the_stack
else
	echo "you have not provided all the required environment variables"
	exit 1
fi

if [ ! -z "${CI}" ]; then
    echo "no need to unset docker variables"
    rm -rf /tmp/gitlab-ci*
else
    eval $(docker-machine env -u)
fi
