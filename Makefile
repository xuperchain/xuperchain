ifeq ($(OS),Windows_NT)
  PLATFORM="Windows"
else
  ifeq ($(shell uname),Darwin)
    PLATFORM="MacOS"
  else
    PLATFORM="Linux"
  endif
endif

all: build

export  GOPROXY=https://goproxy.cn,direct

XCHAIN_ROOT := ${PWD}/core
export XCHAIN_ROOT
PATH := ${PWD}/core/xvm/compile/wabt/build:$(PATH)

build:
	PLATFORM=$(PLATFORM) ./core/scripts/build.sh

test:
	go test -coverprofile=coverage.txt -covermode=atomic ./...
	# test wasm sdk
	GOOS=js GOARCH=wasm go build github.com/xuperchain/xuperchain/core/contractsdk/go/driver

contractsdk:
	make -C core/contractsdk/cpp build
	make -C core/contractsdk/cpp test

clean:
	rm -rf output
	rm -f xchain-cli
	rm -f xchain
	rm -f dump_chain
	rm -f event_client
	rm -rf ./core/xvm/compile/wabt/build/

.PHONY: all test clean
