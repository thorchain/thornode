
setup:
	@bundle install

test:
	@rspec --fail-fast

wait-for-blockchain:
	@./wait-for-blockchain.sh

install:
	make -C ../.. install setup
