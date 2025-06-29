.PHONY: run-containers
run-containers:
	docker run --rm -d -e APP_NAME="Google" -p 9001:1233 --name server1 docker.io/rashree2023/load-balancer-probe-replica:v3
	docker run --rm -d -e APP_NAME="Amazon" -p 9002:1233 --name server2 docker.io/rashree2023/load-balancer-probe-replica:v3
	docker run --rm -d -e APP_NAME="Oracle" -p 9003:1233 --name server3 docker.io/rashree2023/load-balancer-probe-replica:v3
	docker run --rm -d -e APP_NAME="Local" -p 9004:1233 --name server4 docker.io/rashree2023/load-balancer-probe-replica:v3

## stop: stops all demo services
.PHONY: stop
stop:
	docker stop server1
	docker stop server2
	docker stop server3
	docker stop server4

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## run: starts demo http services
.PHONY: run-proxy-server
run-proxy-server:
	go run cmd/main.go
