all: build

.PHONY: all test clean

export GO111MODULE=on
export GOFLAGS=-mod=vendor
XCHAIN_ROOT := ${PWD}/core
export XCHAIN_ROOT
PATH := ${PWD}/core/xvm/compile/wabt/build:$(PATH)

build:contractsdk
	./core/scripts/build.sh
	make -C core/xvm/compile/wabt -j 8 &&cp core/xvm/compile/wabt/build/wasm2c ./

install: build
	echo set env to xchain 



test:contractsdk-test
	go test -coverprofile=coverage.txt -covermode=atomic ./...
	make -C 
	# GOOS=js GOARCH=wasm go build github.com/xuperchain/xuperchain/core/contractsdk/go/driver 这个测试测的啥

clean:
	rm -rf core/xvm/compile/wabt/build
	find . -name '*.so.*' -exec rm {} \;

contractsdk:
	make -C core/contractsdk build

contractsdk-test:contractsdk
	make -C core/contractsdk test

contract:
	docker build -t xuper/xuperchain-local . && docker run -it --name xchain --rm xuper/xuperchain-dev && docker exec -it xchain bash ../core/scripts/start.sh 

docker-build:


docker-test: