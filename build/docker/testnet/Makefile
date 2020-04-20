mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
path := $(dir $(mkfile_path))

ifeq ($(OS),Windows_NT)
    uname_S := Windows
else
    uname_S := $(shell uname -s)
endif


run-testnet-standalone:
	TAG=${TAG} NET=testnet docker-compose -p thornode -f ${path}../components/standalone.base.yml -f ${path}../components/midgard.yml -f ${path}../components/midgard.${uname_S}.yml -f ${path}../components/standalone.${uname_S}.yml up --force-recreate --remove-orphans --no-build -d

reset-testnet-standalone: stop clean run-testnet-standalone

run-testnet-genesis:
	TAG=testnet NET=testnet docker-compose -p thornode -f ${path}../components/genesis.base.yml -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../components/midgard.${uname_S}.yml -f ${path}../testnet/genesis.yml up --force-recreate --remove-orphans -d

run-testnet-genesis-ci:
	TAG=testnet NET=testnet docker-compose -p thornode -f ${path}../components/genesis.base.yml -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../components/midgard.ci.yml -f ${path}../testnet/genesis.yml up --force-recreate --remove-orphans -d

reset-testnet-genesis: stop clean run-testnet-genesis

run-testnet-genesis-churn:
	TAG=testnet NET=testnet docker-compose -p thornode -f ${path}../components/genesis.base.yml -f ${path}../components/fifthnode.yml -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../components/midgard.${uname_S}.yml -f ${path}../testnet/genesis.yml up --force-recreate --remove-orphans -d

reset-testnet-genesis-churn: stop clean run-testnet-genesis-churn

run-testnet-validator:
	TAG=${TAG} NET=testnet PEER=$(shell curl -sL testnet-seed.thorchain.info/node_ip_list.json | jq -r .[] | shuf -n 1) docker-compose -p thornode -f ${path}../components/midgard.yml -f ${path}../components/midgard.${uname_S}.yml -f ${path}../components/validator.yml -f ${path}../components/validator.${uname_S}.yml up --force-recreate --remove-orphans --no-build -d

run-testnet-validator-ci:
	TAG=testnet NET=testnet docker-compose -p thornode -f ${path}../components/frontend.yml -f ${path}../components/midgard.yml -f ${path}../components/midgard.ci.yml -f ${path}../components/validator.yml -f ${path}../components/validator.Darwin.yml up --force-recreate --remove-orphans -d

reset-testnet-validator: stop clean run-testnet-validator

run-testnet-binance:
	@mkdir -p ${HOME}/.bnbchaind && chown 1000:1000 ${HOME}/.bnbchaind
	@mkdir -p ${HOME}/.bnbcli && chown 1000:1000 ${HOME}/.bnbcli
	docker-compose -p thornode -f ${path}../components/binance-testnet.yml up --force-recreate --remove-orphans -d

clean-testnet-binance:
	@rm -rf ${HOME}/.bnbchaind
	@rm -rf ${HOME}/.bnbcli

reset-testnet-binance: stop clean-testnet-binance run-testnet-binance

refund-testnet:
	TAG=testnet NET=testnet docker-compose -p thornode -f ${path}../components/standalone.base.yml -f ${path}../components/standalone.${uname_S}.yml exec thor-api /docker/scripts/refund.sh localhost:1317
