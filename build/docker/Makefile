.PHONY: build smoke-test tools

build: 
	docker build -t registry.gitlab.com/thorchain/thornode -f Dockerfile ../..

clean:
	@rm -rf /tmp/genesis
	@rm -rf ~/.thor*
	@rm -rf ~/.signer
	@rm -rf /tmp/shared
	@mkdir /tmp/genesis
	@mkdir /tmp/shared
	@mkdir ~/.thord
	@mkdir ~/.signer
	@mkdir ~/.thornode
	@mkdir ~/.thornode/standalone
	@mkdir ~/.thornode/genesis_1
	@mkdir ~/.thornode/genesis_2
	@mkdir ~/.thornode/genesis_3
	@mkdir ~/.thornode/genesis_4
	@mkdir ~/.thornode/validator

clean-validator:
	@rm -rf ~/.thornode/validator
	@mkdir ~/.thornode/validator

run-standaloneWithMockBinance: clean
	docker-compose -f standalone.yml -f standaloneWithMockBinance.yml up --force-recreate --remove-orphans -d

stop-standaloneWithMockBinance: clean
	docker-compose -f standalone.yml -f standaloneWithMockBinance.yml down

run-standalone: clean
	docker-compose -f standalone.yml up --force-recreate --remove-orphans -d

stop-standalone: clean
	docker-compose -f standalone.yml down

run-genesisWithMockBinance: clean
	docker-compose -f genesis.yml -f genesisWithMockBinance.yml up --force-recreate --remove-orphans -d

stop-genesisWithMockBinance: clean
	docker-compose -f genesis.yml -f genesisWithMockBinance.yml down

run-genesis: clean
	docker-compose -f genesis.yml up --force-recreate --remove-orphans -d

stop-genesis: clean
	docker-compose -f genesis.yml down

run-validator: clean-validator
	docker-compose -f validator.yml --project-name validator up --force-recreate -d
