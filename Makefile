.PHONY: run-containers
run-containers:
	podman run --rm -d -p 9001:1233 --name server1 localhost/rashree2023/load-balancer-probe-replica:v3
	podman run --rm -d -p 9002:1233 --name server2 localhost/rashree2023/load-balancer-probe-replica:v3
	podman run --rm -d -p 9003:1233 --name server3 localhost/rashree2023/load-balancer-probe-replica:v3
	podman run --rm -d -p 9004:1233 --name server4 localhost/rashree2023/load-balancer-probe-replica:v3

## stop: stops all demo services
.PHONY: stop
stop:
	podman stop server1
	podman stop server2
	podman stop server3
	podman stop server4

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## run: starts demo http services
.PHONY: run-proxy-server
run-proxy-server:
	go run cmd/main.go

## run: starts rabbitmq service
.PHONY: run-rabbitmq
run-rabbitmq:
	podman run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 docker.io/rabbitmq:4.0-management
