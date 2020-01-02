#!/bin/sh

set -x

USER=$(whoami)
DISK_SIZE=100
cd ../../
LOCAL_VOLUME=$(pwd)

###########
# CLEANUP #
###########
cleanup () {
    echo "performing cleanup"
    eval $(docker-machine env -u)
    docker-machine rm -f $1 > /dev/null 2>&1
    sleep $2
}

####################
# START THE STACK  #
####################
start_the_stack () {
    cd build/docker/
    echo "waiting for server to be ready"
    PROVISIONING_TIME=$1
    sleep ${PROVISIONING_TIME}
    eval $(docker-machine env ${DOCKER_SERVER})
    export NET=${THORNODE_ENV}
    if [ "$THORNODE_ENV" = "mocknet" ]; then
        docker-compose -p thornode \
            -f components/genesis.base.yml \
            -f components/frontend.yml \
            -f components/midgard.yml \
            -f components/mock-binance.yml \
            -f ${THORNODE_ENV}/genesis.yml up --force-recreate --remove-orphans -d
    else
        docker-compose -p thornode \
            -f components/genesis.base.yml \
            -f components/frontend.yml \
            -f components/midgard.yml \
            -f ${THORNODE_ENV}/genesis.yml up --force-recreate --remove-orphans -d
    fi
    sleep 60
}

#####################
# VERIFY THE STACK  #
#####################
verify_the_stack () {
    eval $(docker-machine env -u)
    docker-machine ssh ${DOCKER_SERVER} sudo docker ps
    echo "allow a few mins for docker services to come up"
    sleep 180
    echo "performing healthchecks"
    HEALTHCHECK_URL="http://localhost:8080/v1/thorchain/pool_addresses"
    HEALTHCHECK=$(docker-machine ssh ${DOCKER_SERVER} curl -s -o /dev/null -w "%{http_code}" ${HEALTHCHECK_URL})
    if  [ "$HEALTHCHECK" == 200 ]; then
	    echo "HEALTHCHECK PASSED"
    else
	    echo "HEALTHCHECK FAILED"
    fi
}

########################
# CREATE DOCKER SERVER #
########################
if [ ! -z "${AWS_VPC_ID}" ] && [ ! -z "${AWS_REGION}" ] && [ ! -z "${AWS_INSTANCE_TYPE}" ]; then
    DOCKER_SERVER="${USER}-aws-thornode-${THORNODE_ENV}-server"
    cleanup ${DOCKER_SERVER} 30
	echo "creating docker node on AWS"
	docker-machine create --driver amazonec2 \
        --amazonec2-vpc-id=${AWS_VPC_ID} \
        --amazonec2-region ${AWS_REGION} \
        --amazonec2-instance-type ${AWS_INSTANCE_TYPE} \
        --amazonec2-root-size ${DISK_SIZE} \
        ${DOCKER_SERVER}
    if [ $? != 0 ]; then
        echo "server could not be created"
        exit 1
    fi
    ##################
    # MOUNT VOLUMES ##
    ##################
    docker-machine ssh ${DOCKER_SERVER} sudo mkdir -p ${LOCAL_VOLUME}
    docker-machine ssh ${DOCKER_SERVER} sudo chmod 777 -R ${LOCAL_VOLUME}
    docker-machine scp -r ${LOCAL_VOLUME}/. ${DOCKER_SERVER}:${LOCAL_VOLUME}
    start_the_stack 180
    verify_the_stack
else
	echo "you have not provided all the required environment variables"
	echo "creating docker node using virtualbox"
	DOCKER_SERVER="${USER}-local-thornode-${THORNODE_ENV}-server"
    cleanup ${DOCKER_SERVER} 10
	docker-machine create --driver virtualbox \
        ${DOCKER_SERVER}
    if [ $? != 0 ]; then
        echo "server could not be created"
        exit 1
    fi
    exit_status "Virtualbox server could not be created"
    ##################
    # MOUNT VOLUMES ##
    ##################
    docker-machine ssh ${DOCKER_SERVER} mkdir -p ${LOCAL_VOLUME}
    docker-machine scp -r ${LOCAL_VOLUME}/. ${DOCKER_SERVER}:${LOCAL_VOLUME}
    start_the_stack 60
    verify_the_stack
fi




