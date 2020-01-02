mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
path := $(dir $(mkfile_path))

run-testnet-standalone: 
	NET=testnet docker-compose -p thornode -f ${path}../components/standalone.base.yml -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml up --force-recreate --remove-orphans -d

reset-testnet-standalone: stop clean run-testnet-standalone

run-testnet-genesis: 
	NET=testnet docker-compose -p thornode -f ${path}../components/genesis.base.yml -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../testnet/genesis.yml up --force-recreate --remove-orphans -d

reset-testnet-genesis: stop clean run-testnet-genesis

run-testnet-genesis-churn: 
	NET=testnet docker-compose -p thornode -f ${path}../components/genesis.base.yml -f ${path}../components/fifthnode.yml -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../testnet/genesis.yml up --force-recreate --remove-orphans -d

reset-testnet-genesis-churn: stop clean run-testnet-genesis-churn
