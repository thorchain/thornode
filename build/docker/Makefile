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

run-standalone: clean
	docker-compose -f standalone.yml up --force-recreate --remove-orphans -d

run-genesis: clean
	docker-compose -f genesis.yml up --force-recreate --remove-orphans -d

run-validator: clean-validator
	docker-compose -f validator.yml --project-name validator up --force-recreate -d

smoke-test:
	@bsinner -f ${FAUCET_KEY} -p ${PRIV_KEY} -e local -c ../../test/smoke/scenarios/flow-test-audit-1p.json -l /tmp/smoke-test.json -d true

validate-smoke-test: smoke-test
	@echo "$(shell cat ../../test/smoke/checksum) /tmp/smoke-test.json" | sha256sum --check
