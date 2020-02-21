mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
path := $(dir $(mkfile_path))
TAG ?= testnet

ifeq ($(OS),Windows_NT)
    uname_S := Windows
else
    uname_S := $(shell uname -s)
endif


run-testnet-standalone:
	TAG=${TAG} NET=testnet docker-compose -p thornode -f ${path}../components/standalone.base.yml -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../components/midgard.${uname_S}.yml -f ${path}../components/standalone.${uname_S}.yml up --force-recreate --remove-orphans -d

reset-testnet-standalone: stop clean run-testnet-standalone

run-testnet-genesis:
	TAG=${TAG} NET=testnet docker-compose -p thornode -f ${path}../components/genesis.base.yml -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../components/midgard.${uname_S}.yml -f ${path}../testnet/genesis.yml up --force-recreate --remove-orphans -d

run-testnet-genesis-ci:
	TAG=${TAG} NET=testnet docker-compose -p thornode -f ${path}../components/genesis.base.yml -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../components/midgard.ci.yml -f ${path}../testnet/genesis.yml up --force-recreate --remove-orphans -d

reset-testnet-genesis: stop clean run-testnet-genesis

run-testnet-genesis-churn:
	TAG=${TAG} NET=testnet docker-compose -p thornode -f ${path}../components/genesis.base.yml -f ${path}../components/fifthnode.yml -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../components/midgard.${uname_S}.yml -f ${path}../testnet/genesis.yml up --force-recreate --remove-orphans -d

reset-testnet-genesis-churn: stop clean run-testnet-genesis-churn

run-testnet-validator:
	TAG=${TAG} NET=testnet docker-compose -p thornode -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../components/midgard.${uname_S}.yml -f ${path}../components/validator.yml -f ${path}../components/validator.${uname_S}.yml up --force-recreate --remove-orphans -d

run-testnet-validator-ci:
	TAG=${TAG} NET=testnet docker-compose -p thornode -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../components/midgard.ci.yml -f ${path}../components/validator.yml -f ${path}../components/validator.Darwin.yml up --force-recreate --remove-orphans -d

reset-testnet-validator: stop clean run-testnet-validator

run-testnet-binance:
	@mkdir -p ${HOME}/.binance && chown 1000:1000 ${HOME}/.binance
	docker-compose -p thornode -f ${path}../components/binance-testnet.yml -f ${path}../components/binance-testnet.${uname_S}.yml up --force-recreate --remove-orphans -d

clean-testnet-binance:
	@rm -rf ${HOME}/.binance

reset-testnet-binance: stop clean-testnet-binance run-testnet-binance
