.PHONY: run-containers
run-containers:
	podman run --rm -d -p 9001:80 --name server1 docker.io/kennethreitz/httpbin
	podman run --rm -d -p 9002:80 --name server2 docker.io/kennethreitz/httpbin
	podman run --rm -d -p 9003:80 --name server3 docker.io/kennethreitz/httpbin

## stop: stops all demo services
.PHONY: stop
stop:
	podman stop server1
	podman stop server2
	podman stop server3

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## run: starts demo http services
.PHONY: run-proxy-server
run-proxy-server:
	go run cmd/main.go
