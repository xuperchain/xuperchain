#!/bin/bash

set -e

cd `dirname $0`
protoc --cpp_out=xchain -I ../pb ../pb/contract.proto
