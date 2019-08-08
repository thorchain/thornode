
setup:
	@bundle install

test-internal:
	@rspec --fail-fast

test:
	@docker-compose up

wait-for-blockchain:
	@./wait-for-blockchain.sh

install:
	make -C ../.. setup

run:
	make -C ../.. start-daemon > /var/log/daemon.log 2>&1 &
	make -C ../.. start-rest > /var/log/rest.log 2>&1 &
