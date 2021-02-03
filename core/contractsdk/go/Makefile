proto:
	protoc --go_out=pb --go_opt=paths=source_relative -I../pb ../pb/contract.proto
	protoc --go_out=plugins=grpc:pbrpc --go_opt=paths=source_relative -I../pb ../pb/contract_service.proto
build:
	make -C example build

test: test-example test-sdk

test-sdk:
	go test ./...
test-example:
	make -C example test

clean:
	make -C example clean
