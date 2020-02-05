[![pipeline status](https://gitlab.com/thorchain/thornode/badges/master/pipeline.svg)](https://gitlab.com/thorchain/thornode/commits/master)
[![coverage report](https://gitlab.com/thorchain/thornode/badges/master/coverage.svg)](https://gitlab.com/thorchain/thornode/commits/master)
[![Build Status](https://gitlab.com/thorchain/thornode/badges/master/build.svg)](https://gitlab.com/thorchain/thornode/commits/master)

# THORChain
======================================

THORChain is a decentralised liquidity network built with [CosmosSDK](cosmos.network). 

### THORNodes
The THORNode software allows a node to join and service the network, which will run with a minimum of four nodes. The only limitation to the number of nodes that can participate is set by the `minimumBondAmount`, which is the minimum amount of capital required to join. Nodes are not permissioned; any node that can bond the required amount of capital can be scheduled to churn in. 

THORChain comes to consensus about events observed on external networks via witness transactions from nodes. Swap and stake logic is then applied to these finalised events. Each event causes a state change in THORChain, and some events generate an output transaction which require assets to be moved (outgoing swaps or bond/liquidity withdrawals). These output transactions are then batched, signed by a threshold signature scheme protocol and broadcast back to the respective external network. The final gas fee on the network is then accounted for and the transaction complete. 

This is described as a "1-way state peg", where only state enters the system, derived from external networks. There are no pegged tokens or 2-way pegs, because they are not necessary. On-chain Bitcoin can be swapped with on-chain Ethereum in the time it takes to finalise the confirmed event. 

All funds in the system are fully accounted for and can be audited. All logic is fully transparent. 

### Churn
THORChain actively churns its validator set to prevent stagnation and capture, and ensure liveness in signing committees. Churning is also the mechanism by which the THORNode software can safely facilitate non-contentious upgrades. 

Every 50000 blocks (3 days) THORChain will schedule the oldest and the most unreliable node to leave, and rotate in two new nodes. The next two nodes chosen are simply the nodes with the highest bond. 

During a churn event the following happens:
* The incoming nodes participate in a TSS key-generation event to create new Asgard vault addresses
* When successful, the new vault is tested with a on-chain challenge-response. 
* If successful, the vaults are rolled forward, moving all assets from the old vault to the new vault. 
* The outgoing nodes are refunded their bond and removed from the system. 

### Bifröst
The Bifröst faciliates connections with external networks, such as Binance Chain, Ethereum and Bitcoin. The Bifröst is generally well-abstracted, needing only minor changes between different chains. The Bifröst handles observations of incoming transactions, which are passed into THORChain via special witness transactions. The Bifröst also handles multi-party computation to sign outgoing transactions via a Genarro-Goldfeder TSS scheme. Only 2/3rds of nodes are required to be in each signing ceremony on a first-come-first-serve basis, and there is no log of who is present. In this way, each node maintains plausible deniabilty around involvement with every transaction. 

To add a new chain, adapt one of the existing modules to the new chain, and submit a merge request to be tested and validated. Once merged, new nodes can start signalling support for the new chain. Once a super-majority (67%) of nodes support the new chain it will be added to the network. 

To remove a chain, nodes can stop witnessing it. If a super-majority of nodes do not promptly follow suit, the non-witnessing nodes will attract penalties during the time they do not witness it. If a super-majority of nodes stop witnessing a chain it will invoke a chain-specific Ragnörok, where all funds attributed to that chain will be returned and the chain delisted. 

### Transactions 
The THORChain facilitates the following transactions, which are made on external networks and replayed into the THORChain via witness transactions:
- **STAKE**: Anyone can stake assets in pools. If the asset hasn't been seen before, a new pool is created. 
- **WITHDRAW**: Anyone who is staking can withdraw their claim on the pool.
- **SWAP**: Anyone can send in assets and swap to another, including sending to a destination address, and including optional price protection. 
- **BOND**: Anyone can bond assets and attempt to become a Node. Bonds must be greater than the `minimumBondAmount`, else they will be refunded. 
- **LEAVE**: Nodes can voluntarily leave the system and their bond and rewards will be paid out. Leaving takes 6 hours. 
- **RESERVE**: Anyone can add assets to the Protocol Reserve, which pays out to Nodes and Stakers. 220,447,472 Rune will be funded in this way. 

### Continuous Liquidity Pools
The Staking, Withdrawing and Swapping logic is based on the `CLP` Continuous Liquidity Pool algorithm. 

**Swaps**
The algorithm for processing assets swaps is given by:
`y = (x * Y * X) / (x + X)^2`, where `x = input, X = Input Asset, Y = Output Asset, y = output`

The fee paid by the trader is given by:
`fee = ( x^2 *  Y ) / ( x + X )^2 `

The slip-based fee model has the following benefits:
* Resistant to manipulation
* A proxy for demand of liquidity
* Asymptotes to zero over time, ensuring pool prices match reference prices
* Prevents Impermanent Loss to liquidity providers

**Staking**
The stake units awarded to a liquidity provider is given by:
`stakeUnits = ((R + T) * (r * T + R * t))/(4 * R * T)`, where `r = Rune Staked, R = Rune Balance, T = Token Balance, t = Token Staked`

This allows them to stake asymmetrically since it has no opinion on price. 

### Incentives
The system is safest and most capital-efficient when 67% of Rune is bonded and 33% is staked in pools. At this point, nodes will be paid 67% of the System Income, and liquidity providers will be paid 33% of the income. The Sytem Income is the block rewards (`blockReward = totalReserve / 6 / 6311390`) plus the liquidity fees collected in that block. 

An Incentive Pendulum ensures that liquidity providers receive 100% of the income when 0% is staked (inefficent), and 0% of the income when `totalStaked >= totalBonded` (unsafe).
The Total Reserve accumulates the `transactionFee`, which pays for outgoing gas fees and stabilises long-term value accrual. 

### Governance
There is strictly minimal goverance possible through THORNode software. Each THORNode can only generate valid blocks that is fully-compliant with the binary run by the super-majority. 

The best way to apply changes to the system is to submit a THORChain Improvement Proposal (TIP) for testing, validation and discussion among the THORChain developer community. If the change is beneficial to the network, it can be merged into the binary. New nodes may opt to run this updated binary, signalling via a `semver` versioning scheme. Once the super-majority are on the same binary, the system will update automatically. Schema and logic changes can be applied via this approach. 

Changes to the Bifröst may not need coordination, as long as the changes don't impact THORChain schema or logic, such as adding new chains. 

Emergency changes to the protocol may be difficult to coordinate, since there is no ability to communicate with any of the nodes. The best way to handle an emergency is to invoke Ragnarök, simply by leaving the system. When the system falls below 4 nodes all funds are paid out and the system can be shut-down. 

======================================

## Setup
Ensure you have a recent version of go (ie `1.121) and enabled go modules
```
export GO111MODULE=on
```
And have `GOBIN` in your `PATH`
```
export GOBIN=$GOPATH/bin
```

### Automated Install
To install easily, run the following command...
```bash
make setup
```

### Manual Install
Install via this `make` command.

```bash
make install
```

Once you've installed `thorcli` and `thord`, check that they are there.

```bash
thorcli help
thord help
```

### Configuration

## Start
There are three services you may want to start.

#### Daemon
This runs the backend
```bash
make start
```

#### REST API Service
Starts an HTTP service to service requests to the backend.
```bash
make start-rest
```

#### CORS Proxy
For making requests in a browser to the API backend, you'll need to start a
proxy in front of the API service to give proper CORS headers. This is because
CORS support in Cosmos was removed. In the meantime, use
[cors.io](http://cors.io) as a proxy.

Bifröst
===============

Witness events that take place on external chains.

### Binance Chain Environment
Please create a config.json file
```json
{
  "chain_id": "statechain",
  "pool_address": "pool address",
  "rune_address": "rune address",
  "dex_host": "testnet-dex.binance.org",
  "observer_db_path": "signerdata",
  "block_scanner": {
    "rpc_host": "binance node host",
    "start_block_height": 34912415,
    "block_scan_processors": 2,
    "block_height_discover_back_off": "1s"
  },
  "state_chain": {
    "chain_id": "statechain",
    "chain_host": "localhost:1317",
    "signer_name": "signer name",
    "signer_passwd": "signer password"
  }
}
You could overwrite the above configs using environment variables as well
Export the following environment variables (set based on your environment and/or the net being used [test/prod]):
```bash
export CHAIN_ID=<chain id>
export POOL_ADDRESS=<pool address>
export RUNE_ADDRESS=<rune address>,
export DEX_HOST=<DEX Hostname>
export BLOCK_SCANNER_RPC_HOST=<RPC HOSTNAME>
export BLOCK_SCANNER_START_BLOCK_HEIGHT=34912415
export BLOCK_SCANNER_BLOCK_SCAN_PROCESSORS=2
export BLOCK_SCANNER_BLOCK_HEIGHT_DISCOVER_BACK_OFF=1s
export STATE_CHAIN_CHAIN_ID=STATECHAIN
export STATE_CHAIN_CHAIN_HOST=localhost:1317
export STATE_CHAIN_SIGNER_NAME=signer name
export STATE_CHAIN_SIGNER_PASSWD=signer password
```


### Development
Setup a local server
```bash
make install
make start
```

### Test
Run tests
```bash
make test
```

To run test live when you change a file, use...
```
go get -u github.com/mitranim/gow
make test-watch
```
