#!/usr/bin/env bash

set -ex

THORNODE_ENV=${THORNODE_ENV:-testnet}
THORNODE_REPO=${THORNODE_REPO:-https://gitlab.com/thorchain/thornode.git}
BRANCH=${BRANCH:-master}
GIT_PATH=${GIT_PATH:-/opt/thornode}
LOGFILE=${LOGFILE:-/var/log/thornode.log}
SEED_ENDPOINT=${SEED_ENDPOINT:-https://${THORNODE_ENV}-seed.thorchain.info}

# install apt repositories
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable"

# install essential packages
echo "installing essential packages"
apt-get update -qy
apt-get install -qy \
    build-essential \
    jq \
    make \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg2 \
    cron \
    software-properties-common \
    unzip \
    python3-pip \
    docker-ce

systemctl enable cron # enable cron

# setup ebs volume
if [ -f /dev/xvdb ]; then
    mkfs -t ext4 /dev/xvdb
    mkdir /opt
    mount /dev/xvdb /opt
    echo "/dev/xvdb /opt ext4 defaults 0 0" >> /etc/fstab
fi

# install docker-compose
echo "installing docker-compose" >> $LOGFILE
curl -L "https://github.com/docker/compose/releases/download/1.25.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
docker-compose version >> $LOGFILE

systemctl enable docker # enable docker at boot

echo "installing bnbcli >> $LOGFILE"
wget https://media.githubusercontent.com/media/binance-chain/node-binary/master/cli/testnet/0.6.2/linux/tbnbcli
chmod +x tbnbcli
mv tbnbcli /usr/local/bin/.
tbnbcli version >> $LOGFILE

# clone and start thornode
echo "cloning thornode" >> $LOGFILE
rm -rf $GIT_PATH
mkdir -p $GIT_PATH
git clone --single-branch --branch ${BRANCH} ${THORNODE_REPO} ${GIT_PATH} >> $LOGFILE

mkdir -p /opt/${THORNODE_ENV}
chmod -R 777 /opt/${THORNODE_ENV}

cat <<EOF > /opt/${THORNODE_ENV}/binance-bootstrap
#!/bin/sh

cd $GIT_PATH/build/docker
make run-${THORNODE_ENV}-binance >> $LOGFILE
EOF

cat <<EOF > /opt/${THORNODE_ENV}/genesis-bootstrap
#!/bin/sh

cd $GIT_PATH/build/docker
export TAG=${THORNODE_ENV}
export BINANCE_HOST="http://testnet-binance.thorchain.info:26657"
make run-${THORNODE_ENV}-standalone >> $LOGFILE
EOF

cat <<EOF > /opt/${THORNODE_ENV}/validator-bootstrap
#!/bin/sh

cd $GIT_PATH/build/docker
export TAG=${THORNODE_ENV}
export BINANCE_HOST="http://testnet-binance.thorchain.info:26657"
export PEER=\$(curl -sL ${THORNODE_ENV}-seed.thorchain.info/node_ip_list.json | jq -r .[] | shuf -n 1)
make run-${THORNODE_ENV}-validator >> $LOGFILE
EOF
