#!/bin/bash

cd `dirname $0`

protoc --go_out=pb --go_opt=paths=source_relative -I../pb ../pb/contract.proto
protoc --go_out=plugins=grpc:pbrpc --go_opt=paths=source_relative -I../pb ../pb/contract_service.proto
