#!/bin/bash

# install docker in precondition

docker run -u $UID --rm -v $(pwd):/src hub.baidubce.com/xchain/emcc emmake make
