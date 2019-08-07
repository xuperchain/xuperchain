#!/bin/bash

# install docker in precondition

function test() {
    docker run -u `id -u` -it --rm -v $PWD:/source hub.baidubce.com/xch/contract-dev sh -c 'cd /source && make clean && make test'
}

function build() {
    docker run -u `id -u` --rm -v $(pwd):/src hub.baidubce.com/xchain/emcc emmake make 
}

if [[ -z "$1" ]]; then build; else $1; fi
