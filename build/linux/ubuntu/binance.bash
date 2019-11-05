#!/bin/bash

set -ex

BNCHOME=${BNCHOME:=~/.bnbchaind}
BNET=${BNET:=prod}
VERSION=${VERSION:=0.6.2}

if [ ! -f /etc/debian_version ]; then
	echo "This script is Ubuntu only"
	exit 1
fi

apt-get update
apt-get -y upgrade
apt install -y curl
curl -s https://packagecloud.io/install/repositories/github/git-lfs/script.deb.sh | bash
apt-get install -y git-lfs
git lfs install

git lfs clone https://github.com/binance-chain/node-binary.git

cd node-binary/fullnode/$BNET/$VERSION

mkdir -p $BNCHOME/config
cp config/* $BNCHOME/config/
cp linux/bnbchaind /usr/bin/

bnbchaind start --home $BNCHOME &

echo "Binance started. Check logs at $BNCHOME/bnc.log"
