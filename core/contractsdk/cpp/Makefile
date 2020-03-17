.PHONY: build test

export XDEV_ROOT=$(shell pwd)
export PATH := $(shell pwd)/../../../output/:$(PATH)

build:
	echo ${XDEV_ROOT}
	./build.sh

test:
	mkdir -p build
	[ ! -f build/features ] || xdev build -o build/features.wasm example/features.cc
	xdev test	
