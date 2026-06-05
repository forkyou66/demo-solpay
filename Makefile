build:
	docker build -t solana-pay .

run:
	docker run -p 7542:7542 solana-pay

remove:
	docker rm -f solana-pay
	docker image rm -f solana-pay

.PHONY: build run remove
