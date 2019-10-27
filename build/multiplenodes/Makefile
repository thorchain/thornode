up: clean
	@docker-compose up

down:
	@docker-compose down

clean:
	@rm -rf /tmp/genesis
	@mkdir /tmp/genesis

build:
	docker build -t thornode -f Dockerfile ../..

sh: clean
	@docker-compose run --rm --no-deps -e NODES=1 node1 /bin/sh

single: clean
	@docker-compose run --rm --no-deps -e NODES=1 node1
