build:
	docker build -t binance-fullnode .

run:
	docker run -d --name binance -v /opt/binance-data:/opt/bnbchaind -e "BNET=prod" -p 27146:27146 -p 27147:27147 -p 26660:26660 --restart unless-stopped --security-opt no-new-privileges --ulimit nofile=16000:16000 varnav/binance-node

