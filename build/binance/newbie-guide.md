# Binance node with Docker newbie guide

## What is this?

This thing allows you to start Binance full node in just a single command, this one:

`docker run --rm -it varnav/binance-node`

This means you don't need to read, understand and follow all this [pages long manual](https://docs.binance.org/fullnode.html#run-full-node-to-join-binance-chain).

## What is Binance full node?

A machine that processes Binance blockchain.

> A full node of Binance Chain is a witness, which observes the consensus messaging, downloads blocks from data seed nodes and executes business logic to achieve the consistent state as validator node (and other full node). Full nodes also help the network by accepting transactions from other nodes and then relaying them to the core Binance network.

## Ok and how do I do that?

You will need Docker. Don't have it? No worries, it's very simple to get it. First, install, or even run [Ubuntu](https://www.ubuntu.com/download/desktop) from live CD/USB. Open terminal with `Ctrl + Alt + T`, then become root with `sudo -i` and run this:

`curl -sSL https://get.docker.com/ | /bin/sh`

And that's it. You now have Docker installed!

## Running production node in background

Ok, and if you want to run node in background with production network, use this command:

`docker run -d --rm --name binance -v /opt/binance-data:/opt/bnbchaind -e "BNET=prod" -p 27146:27146 -p 127.0.0.1:27147:27147 --security-opt no-new-privileges --ulimit nofile=16000:16000 varnav/binance-node`

All data and config files will be stored in `/opt/binance-data/`, while everything else will be running inside container and will be automatically deleted if you stop the container with `docker stop binance`. Data and configs will be, of course, preserved - and you will be easily able to run container again using command above.

To get into console and use `bnbcli` run this:

`docker exec -it binance-devel /bin/bash`

And to see logs run this:

`docker logs -f binance-testnet`

## Is all this free?

Yes, all this is free software.
