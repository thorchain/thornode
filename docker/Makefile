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
	@rm -rf ~/.thornode/standalone
	@mkdir /tmp/genesis
	@mkdir ~/.thornode/standalone

run-standalone: clean
	docker-compose -f standalone.yml up --force-recreate --remove-orphans
