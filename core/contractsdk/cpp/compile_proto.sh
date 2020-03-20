#!/bin/bash

set -e

cd `dirname $0`
protoc --cpp_out=src/xchain -I ../pb ../pb/contract.proto
