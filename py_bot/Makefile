# Makefile for Docker

IMAGE_NAME = habitry
ENV_FILE = .env

# Настройки прокси для Docker контейнера
SOCKS_PROXY = socks5h://localhost:10808

build:
	docker build --tag $(IMAGE_NAME) .

run:
	docker run --env-file $(ENV_FILE) --network="host" -it $(IMAGE_NAME)

run-proxy:
	docker run --env-file $(ENV_FILE) --network="host" \
		-e HTTP_PROXY=$(SOCKS_PROXY) \
		-e HTTPS_PROXY=$(SOCKS_PROXY) \
		-e ALL_PROXY=$(SOCKS_PROXY) \
		-it $(IMAGE_NAME)

clean:
	docker image rm $(IMAGE_NAME)

runn:
	docker build --tag $(IMAGE_NAME) .
	docker run --env-file $(ENV_FILE) --network="host" -it $(IMAGE_NAME)

runn-proxy:
	docker build --tag $(IMAGE_NAME) .
	docker run --env-file $(ENV_FILE) --network="host" \
		-e HTTP_PROXY=$(SOCKS_PROXY) \
		-e HTTPS_PROXY=$(SOCKS_PROXY) \
		-e ALL_PROXY=$(SOCKS_PROXY) \
		-it $(IMAGE_NAME)

test-proxy:
	docker run --env-file $(ENV_FILE) --network="host" \
		-e HTTP_PROXY=$(SOCKS_PROXY) \
		-e HTTPS_PROXY=$(SOCKS_PROXY) \
		-e ALL_PROXY=$(SOCKS_PROXY) \
		--rm -it $(IMAGE_NAME) \
		python -c "import requests; print('Testing proxy...'); r = requests.get('https://httpbin.org/ip', timeout=10); print('IP:', r.json())"