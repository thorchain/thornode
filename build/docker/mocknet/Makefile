mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
path := $(dir $(mkfile_path))

run-mocknet-standalone: 
	NET=mocknet docker-compose -p thornode -f ${path}../components/standalone.base.yml -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../components/mock-binance.yml -f ${path}../mocknet/standalone.yml up --force-recreate --remove-orphans -d

reset-mocknet-standalone: stop clean run-mocknet-standalone

run-mocknet-genesis: 
	NET=mocknet docker-compose -p thornode -f ${path}../components/genesis.base.yml -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../components/mock-binance.yml -f ${path}../mocknet/genesis.yml up --force-recreate --remove-orphans -d

reset-mocknet-genesis: stop clean run-mocknet-genesis

run-mocknet-genesis-churn: 
	NET=mocknet docker-compose -p thornode -f ${path}../components/genesis.base.yml -f ${path}../components/fifthnode.yml -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../components/mock-binance.yml -f ${path}../mocknet/genesis.yml -f ${path}../mocknet/churn.yml up --force-recreate --remove-orphans -d

reset-mocknet-genesis-churn: stop clean run-mocknet-genesis-churn

run-mocknet-validator: 
	NET=mocknet docker-compose -p thornode -f ${path}../components/validator.yml up --force-recreate --remove-orphans -d

reset-mocknet-validator: stop clean run-mocknet-validator
