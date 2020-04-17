#!/bin/sh

set -x

USER=$(hostname)
export DISK_SIZE=${DISK_SIZE:=100}
export AWS_INSTANCE_TYPE=${AWS_INSTANCE_TYPE:=c5.xlarge}

cd ../../
LOCAL_VOLUME=$(pwd)

if [ $1 == "ci" ]; then
    export CI="true"
    export AWS_REGION=$AWS_CI_REGION
    export USER=$CI_JOB_ID
fi

###########
# CLEANUP #
###########
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
    curl -L $base/docker-machine-$(uname -s)-$(uname -m) >/tmp/docker-machine &&  install /tmp/docker-machine /usr/local/bin/docker-machine
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
start_the_stack () {
    cd build/docker/${THORNODE_ENV}
    echo "waiting for server to be ready"
    PROVISIONING_TIME=$1
    sleep ${PROVISIONING_TIME}
    export NET=${THORNODE_ENV}
    export TAG=${THORNODE_ENV}
    eval $(docker-machine env ${DOCKER_SERVER} --shell bash)
    make run-${THORNODE_ENV}-genesis-ci
    sleep 60
}

#####################
# VERIFY THE STACK  #
#####################
verify_the_stack () {
    if [ ! -z "${CI}" ]; then
        echo "no need to unset docker variables"
    else
        eval $(docker-machine env -u)
    fi
    docker-machine ssh ${DOCKER_SERVER} sudo docker ps -a
    echo "allow a few mins for docker services to come up"
    sleep 70
    echo "performing healthchecks"
    IP=$(docker-machine ip ${DOCKER_SERVER})
    HEALTHCHECK_URL="http://${IP}:8080/v1/thorchain/pool_addresses"
    HEALTHCHECK_CMD=$(curl -s -o /dev/null -w "%{http_code}" ${HEALTHCHECK_URL})
    if  [ "${HEALTHCHECK_CMD}" == 200 ]; then
	    echo "HEALTHCHECK PASSED"
    else
	    echo "HEALTHCHECK FAILED"
	    exit 1
    fi
}

########################
# CREATE DOCKER SERVER #
########################
if [ -z "${THORNODE_ENV}" ]; then
    echo "please provide \$THORNODE_ENV environment variable"
    exit 1
fi

if [ ! -z "${AWS_VPC_ID}" ] && [ ! -z "${AWS_REGION}" ] && [ ! -z "${AWS_INSTANCE_TYPE}" ]; then
    DOCKER_SERVER="${USER}-aws-${THORNODE_ENV}"
    cleanup ${DOCKER_SERVER} 20
	echo "creating server node on AWS"
	docker-machine create --driver amazonec2 \
        --amazonec2-vpc-id=${AWS_VPC_ID} \
        --amazonec2-region ${AWS_REGION} \
        --amazonec2-instance-type ${AWS_INSTANCE_TYPE} \
        --amazonec2-root-size ${DISK_SIZE} \
        --amazonec2-security-group ${SECURITY_GROUP} \
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
    start_the_stack 60
    verify_the_stack
else
	echo "you have not provided all the required environment variables"
	exit 1
fi
