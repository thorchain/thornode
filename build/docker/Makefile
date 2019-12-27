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
	@mkdir -p ~/.thornode/validator

run-standaloneWithMockBinance: 
	docker-compose -f standalone.yml -f standaloneWithMockBinance.yml up --force-recreate --remove-orphans -d

stop-standaloneWithMockBinance: 
	docker-compose -f standalone.yml -f standaloneWithMockBinance.yml down

reset-standaloneWithMockBinance: stop-standaloneWithMockBinance clean run-standaloneWithMockBinance

run-standaloneFullStack: 
	docker-compose -f standalone.yml -f standaloneFullStack.yml up --force-recreate --remove-orphans -d

stop-standaloneFullStack: 
	docker-compose -f standalone.yml -f standaloneFullStack.yml down

reset-standaloneFullStack: stop-standaloneFullStack clean run-standaloneFullStack

run-standalone: 
	docker-compose -f standalone.yml up --force-recreate --remove-orphans -d

stop-standalone: 
	docker-compose -f standalone.yml down

reset-standalone: stop-standalone clean run-standalone

run-genesisFullStack: 
	docker-compose -f genesis.yml -f genesisFullStack.yml up --force-recreate --remove-orphans -d

stop-genesisFullStack: 
	docker-compose -f genesis.yml -f genesisFullStack.yml down

reset-genesisFullStack: stop-genesisFullStack clean run-genesisFullStack

run-genesisWithMockBinance: 
	docker-compose -f genesis.yml -f genesisWithMockBinance.yml up --force-recreate --remove-orphans -d

stop-genesisWithMockBinance: 
	docker-compose -f genesis.yml -f genesisWithMockBinance.yml down

reset-genesisWithMockBinance: stop-genesisWithMockBinance clean run-genesisWithMockBinance

run-genesisWithChurn: 
	docker-compose -f genesis.yml -f genesisWithMockBinance.yml -f genesisWithChurnNode.yml up --force-recreate --remove-orphans -d

stop-genesisWithChurn: 
	docker-compose -f genesis.yml -f genesisWithMockBinance.yml -f genesisWithChurnNode.yml down

reset-genesisWithChurn: stop-genesisWithChurn clean run-genesisWithChurn

run-genesis: 
	docker-compose -f genesis.yml up --force-recreate --remove-orphans -d

stop-genesis: 
	docker-compose -f genesis.yml down

reset-genesis: stop-genesis clean run-genesis

run-validator: 
	docker-compose -f validator.yml --project-name validator up --force-recreate -d

reset-validator: clean-validator run-validator
