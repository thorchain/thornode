How to Become a THORNode
==============================

## System setup
To setup your server to run full THORNode

### Docker Setup
TODO: Docker documentation

### Linux Setup
This documentation explains how to setup a Linux server as a THORNode
manually.

Before we do anything, lets get the source code and build our binaries.
```bash
git clone git@gitlab.com:thorchain/bepswap/thornode.git
make install
```

Next, we will need to setup the Binance full node....
#### Binance
Binance themselves provide documentation. Please follow their
[documentation](https://docs.binance.org/fullnode.html).

Alternatively, you can use a pre made [docker
image](https://github.com/varnav/binance-node-docker) to simplfy it. 

Wait until your Binance full node is caught up before continuing onto the next
sections.

#### Observer
TODO

#### Signer
TODO

#### Thord
To setup `thord`, we'll need to run the following commands.

```bash
thord init local --chain-id statechain

thorcli keys add operator
thorcli keys add observer

thorcli config chain-id statechain
thorcli config output json
thorcli config indent true
thorcli config trust-node true
```

Next, we need to get the genesis file of Thorchain.
For testnet, run...
```bash
curl
https://gitlab.com/thorchain/bepswap/thornode/raw/master/genesis/testnet.json -o ~/.thord/config/genesis.json
```

For mainnet, run...
```bash
curl
https://gitlab.com/thorchain/bepswap/thornode/raw/master/genesis/mainnet.json -o ~/.thord/config/genesis.json
```

Validate your genesis file is valid.
```bash
thord validate-genesis
```

You can now start your `thord` process

```bash
thord start --rpc.laddr tcp://0.0.0.0:26657
```

#### Rest Server
To start the rest API of your `thord` daemon, run the following...

```bash
thorcli rest-server --laddr tcp://0.0.0.0:1317
```


## Bonding
In order to become a validator, you must bond the minimum amount of rune to a
`thor` address. 

To do so, send your rune to the Thorchain with a specific memo. You will need
your thor address to do so. You can retrieve that via...
```bash
thorcli keys show operator --address
```

Once you have your address, include in your memo to Thorchain
```
BOND:<address>
```

Once you have done that, you can then use the `thorcli` to
register your other addresses.

```bash
thorcli tx swapservice set-trust-account $(thorcli keys show observer --address) $(thord tendermint show-validator)
```

Once you have done this, your node is ready to be rotated into the active
group of validators.
