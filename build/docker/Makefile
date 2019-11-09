.PHONY: build smoke tools

build: 
	docker build -t registry.gitlab.com/thorchain/bepswap/thornode -f Dockerfile ../..

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

clean-validator:
	@rm -rf ~/.thornode/validator
	@mkdir ~/.thornode/validator

run-standalone: clean
	docker-compose -f standalone.yml up --force-recreate --remove-orphans -d

run-genesis: clean
	docker-compose -f genesis.yml up --force-recreate --remove-orphans -d

run-validator: clean-validator
	docker-compose -f validator.yml --project-name validator up --force-recreate -d

tools:
	@make -C ../../ tools

smoke-test: run-standalone tools
	@make PRIV_KEY=$(shell cat ~/.thornode/standalone/keys/private_key.txt) ENV=local -C ../../ smoke-test

validate-smoke-test: smoke-test
	@echo "$(shell cat ../../test/smoke/checksum) /tmp/smoke-test.json" | sha256sum --check
