include mocknet/Makefile testnet/Makefile
.PHONY: build 

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

stop: 
	BUILD_FLAG="" docker-compose -p thornode down --remove-orphans

run-validator: 
	docker-compose -f validator.yml --project-name validator up --force-recreate -d

reset-validator: clean-validator run-validator
