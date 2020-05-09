#!/bin/bash

set -x

start-docker-thornode () {
    ssh -oStrictHostKeyChecking=no root@$1 "curl -s https://gist.githubusercontent.com/cbarraford/97622088230e10db8a9b9af26b938747/raw/7410ab031c9d9c6b4bc9594d13bbb99ef2b2d61f/docker-setup.bash | BRANCH=mocknet10 PEER=67.205.166.241 BINANCE_HOST=http://67.205.166.241:26660 bash"
}

start-thornode () {
    ssh -oStrictHostKeyChecking=no root@$1 "curl -s https://gist.githubusercontent.com/cbarraford/97622088230e10db8a9b9af26b938747/raw/cbab98326cb6aadad808310df4a0f466ae581b05/setup.bash | BRANCH=mocknet10 PEER=67.205.166.241 BINANCE_HOST=http://67.205.166.241:26660 bash"
}

stop-thornode () {
    ssh -oStrictHostKeyChecking=no root@$1 "docker rm -f \$(docker ps -q)"
}

leave-thornode () {
    ssh -oStrictHostKeyChecking=no root@$1 "bash thornode/build/scripts/mock-leave.sh 67.205.166.241 \$(cat ~/.thornode/validator/.bond/address.txt)"
}

reboot-thornode () {
    ssh -oStrictHostKeyChecking=no root@$1 "reboot"
}

pull-thornode () {
    ssh -oStrictHostKeyChecking=no root@$1 "make -C thornode pull"
}

ps-thornode () {
    ssh -oStrictHostKeyChecking=no root@$1 "docker ps"
}

tss-thornode () {
    ssh -oStrictHostKeyChecking=no root@$1 "curl -s https://gist.githubusercontent.com/cbarraford/97622088230e10db8a9b9af26b938747/raw/48e527f4a52cd928eeba9193cff7383ff427f778/go-tss.bash | BRANCH=master bash"
}

build-thornode () {
    # ssh -oStrictHostKeyChecking=no root@$1 "cd thornode; TAGS=sandbox make -C build/docker build"
    ssh -oStrictHostKeyChecking=no root@$1 "curl -s https://gist.githubusercontent.com/cbarraford/97622088230e10db8a9b9af26b938747/raw/cbab98326cb6aadad808310df4a0f466ae581b05/build.bash | BRANCH=mocknet10 bash"
}

thorid-thornode () {
    ssh -oStrictHostKeyChecking=no root@$1 "curl -s https://gist.githubusercontent.com/cbarraford/97622088230e10db8a9b9af26b938747/raw/d58cef53045aa568684874f3b51a2b2e3d015484/thorchainID.bash | bash"
}

prune-thornode () {
    ssh -oStrictHostKeyChecking=no root@$1 "docker system prune --volumes -a -f"
}

p2pid-thornode () {
    curl $1:4040/p2pid
}

$1 64.227.13.31
$1 64.227.26.96
$1 64.227.19.100
$1 64.227.19.148
$1 167.99.166.120
$1 64.225.35.29
$1 167.99.166.215
$1 64.225.42.50
$1 64.225.41.177
$1 167.99.170.92
$1 167.99.166.63
