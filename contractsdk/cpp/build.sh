#!/bin/bash

# install docker in precondition

function test() {
    verify
    docker run -u $UID -it --rm -v $PWD:/source hub.baidubce.com/xch/contract-dev sh -c 'cd /source && make clean && make test'
}

function build() {
    verify
    docker run -u $UID --rm -v $(pwd):/src hub.baidubce.com/xchain/emcc emmake make 
}

function verify {
    [ ! `id -u` ] && echo "$UID not exist" && exit 1
}

if [[ -z "$1" ]]; then build; else $1; fi
