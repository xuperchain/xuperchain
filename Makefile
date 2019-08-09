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
export GO111MODULE=on
export GOFLAGS=-mod=vendor
XCHAIN_ROOT := ${PWD}
export XCHAIN_ROOT
PATH := ${PWD}/xvm/compile/wabt/build:$(PATH)

build:
	PLATFORM=$(PLATFORM) ./build.sh

test:
	go test `go list ./... | egrep -v 'test'`
	# test wasm sdk
	GOOS=js GOARCH=wasm go build github.com/xuperchain/xuperunion/contractsdk/go/driver
	cd xvm/spectest && go run main.go core

clean:
	rm -rf output
	rm -f xchain-cli
	rm -f xchain
	rm -f dump_chain

.PHONY: all test clean
