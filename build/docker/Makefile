.PHONY: build

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
	docker-compose -f standalone.yml up --force-recreate --remove-orphans

run-genesis: clean
	docker-compose -f genesis.yml up --force-recreate --remove-orphans

run-validator: clean-validator
	docker-compose -f validator.yml --project-name validator up --force-recreate
