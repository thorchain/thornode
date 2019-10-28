up: clean
	@docker-compose up

down:
	@docker-compose down

build:
	docker build -t thornode -f Dockerfile ../genesis_ceremony

sh:
	@docker-compose run --rm --no-deps thornode /bin/sh
