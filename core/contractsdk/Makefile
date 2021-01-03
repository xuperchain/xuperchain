all:build

build:
	make -C go build
	make -C cpp build
	make -C java build 
	make -C evm build

test:build
	make -C go test
	make -C cpp test
	make -C java test
	make -C evm test

clean:
	make -C go clean
	make _C cpp test
	make -C java test
	make -C evm test