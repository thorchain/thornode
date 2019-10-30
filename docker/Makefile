.PHONY: build

build: build-observer build-signer build-statechain

build-observer:
	docker build -t thornode-observer -f dockerfiles/observer.Dockerfile ..

build-signer:
	docker build -t thornode-signer -f dockerfiles/signer.Dockerfile ..

build-statechain:
	docker build -t thornode-statechain -f dockerfiles/statechain.Dockerfile ..

clean:
	@rm -rf /tmp/genesis
	@rm -rf ~/.thornode
	@mkdir /tmp/genesis
	@mkdir ~/.thornode
	@mkdir ~/.thornode/standalone
	@mkdir ~/.thornode/genesis_1
	@mkdir ~/.thornode/genesis_2
	@mkdir ~/.thornode/genesis_3
	@mkdir ~/.thornode/genesis_4
	@mkdir ~/.thornode/validator

run-standalone: clean
	docker-compose -f standalone.yml up --force-recreate --remove-orphans

run-genesis: clean
	docker-compose -f genesis.yml up --force-recreate --remove-orphans

run-validator: clean
	docker-compose -f validator.yml up --force-recreate --remove-orphans
