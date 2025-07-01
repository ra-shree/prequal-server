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

r ?= 1000
c ?= 10
## run: send large number of requests
.PHONY: send-load
send-load:
	echo "Running ab with r=$(r) and c=$(c)"
	ab -n $(r) -c $(c) http://localhost:8000/test

## run: show statistics table for instances
.PHONY: show-stats
show-stats:
	go run cmd/stats.go
