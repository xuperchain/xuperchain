# Building xchain using docker

## Builder image

`centos.Dockerfile` and `ubuntu.Dockerfile` is the Dockerfile of builder.

## Use builder image

``` bash
$ cd xuperchain
$ # centos environment
$ docker run --rm -u `id -u`:`id -g` -v `pwd`:`pwd` -w `pwd` xuper/centos-builder:0.1 make
$ # ubuntu environment
$ # docker run --rm -u `id -u`:`id -g` -v `pwd`:`pwd` -w `pwd` xuper/ubuntu-builder:0.1 make
```