How to Become a THORNode
==============================

## System setup
To setup your server to run full THORNode

This documentation explains how to setup a Linux server as a THORNode
manually.

Before THORNode do anything, lets get the source code and build our binaries.
```bash
git clone git@gitlab.com:thorchain/thornode.git
make install
```

Next, THORNode will need to setup the Binance full node....
#### Binance
Binance themselves provide documentation. Please follow their
[documentation](https://docs.binance.org/fullnode.html).

Alternatively, you can use a pre made [docker
image](https://github.com/varnav/binance-node-docker) to simplfy it. 

We've provided a ready to use command to start Binance. It is recommended that
this is run on a separate server to THORNode
```bash
make -C build/docker run-testnet-binance
```

Wait until your Binance full node is caught up before continuing onto the next
sections.

#### Thornode
To setup your node, we'll need to run the following commands.

```bash
export PEER=$(curl -sL testnet-seed.thorchain.info/node_ip_list.json | jq -r '.[]' | sort -n -r | head -n 1 | awk '{print $NF}')
export BINANCE_HOST=<IP Address>
make -C build/docker run-testnet-validator
``` 

## Bonding
In order to become a validator, you must bond the minimum amount of rune to a
`thor` address. 

To do so, send your rune to the Thorchain with a specific memo. You will need
your thor address to do so. You can retrieve that via...
```bash
docker exec -it thor-daemon thorcli keys show operator --address
```

Once you have your address, include in your memo to Thorchain
```
BOND:<address>
```

Once you have done that, you can then use the `thorcli` to
register your other addresses.

```bash
docker exec -it thor-daemon /bin/sh
echo password | thorcli tx thorchain set-node-keys $(thorcli keys show thorchain --pubkey) $(thorcli keys show thorchain --pubkey) $(thord tendermint show-validator) --from thorchain --yes
```

Once you have done this, your node is ready to be rotated into the active
group of validators.
