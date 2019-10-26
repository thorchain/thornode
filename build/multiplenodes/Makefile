
build:
	docker build -t thornode -f ../../Genesis.Dockerfile ../..

clean:
	@rm -rf /tmp/genesis
	@mkdir /tmp/genesis

run: clean
	@docker-compose up

sh: clean
	@docker-compose run --rm --no-deps -e NODES=1 node1 /bin/sh
