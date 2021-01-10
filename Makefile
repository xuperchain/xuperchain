all: build

.PHONY: all test clean

export GO111MODULE=on
export GOPROXY=https://goproxy.io,direct
XCHAIN_ROOT := ${PWD}/core
export XCHAIN_ROOT
build: build-release

build-release: contractsdk wasm2c
	bash core/scripts/build.sh

build-debug:contractsdk wasm2c
	XCHAIN_BUILD_FLAG=-gcflags\ \"all\=-N-l\"  bash core/scripts/build.sh

wasm2c:
	make -C core/xvm/compile/wabt -j 8 &&cp core/xvm/compile/wabt/build/wasm2c ./

install: build
	echo set env to xchain 
	echo TBD

test:contractsdk-test
	go test -coverprofile=coverage.txt -covermode=atomic ./...
	GOOS=js GOARCH=wasm go build github.com/xuperchain/xuperchain/core/contractsdk/go/driver 这个测试测的啥

clean:
	rm -rf core/xvm/compile/wabt/build
	find . -name '*.so.*' -exec rm {} \;

contractsdk:
	echo TBD
	# make -C core/contractsdk build

contractsdk-test:contractsdk
	echo TBD
	# make -C core/contractsdk test

# contract:
	# docker build -t xuper/xuperchain-local . && docker run -it --name xchain --rm xuper/xuperchain-dev && docker exec -it xchain bash ../core/scripts/start.sh 
#  build by docker and output to local storage
# docker-build:

# docker-test:
# #  build docker image 
# build-image:
# 	# 