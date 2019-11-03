#!/bin/bash

apt-get update
apt-get -y upgrade
apt install -y curl vim git build-essential jq

wget https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz
tar -xvf go1.13.3.linux-amd64.tar.gz
mv go /usr/local
rm go1.13.3.linux-amd64.tar.gz # cleanup

export GOROOT=/usr/local/go
export GOPATH=~/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

git clone https://gitlab.com/thorchain/bepswap/thornode.git ~/go/src/gitlab.com/thorchain/bepswap/thornode

cd ~/go/src/gitlab.com/thorchain/bepswap/thornode
go get -v
make install tools
