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
# export GOFLAGS=-mod=vendor
XCHAIN_ROOT := ${PWD}/core
export XCHAIN_ROOT
#PATH := ${PWD}/core/xvm/compile/wabt/build:$(PATH)

build:
	PLATFORM=$(PLATFORM) ./core/scripts/build.sh

test:
	go test -coverprofile=coverage.txt -covermode=atomic ./...

clean:
	rm -rf output
	rm -f xchain-cli
	rm -f xchain
	rm -f dump_chain
	rm -f event_client

.PHONY: all test clean
