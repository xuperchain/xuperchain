FROM golang:1.13.2
RUN sed -i 's#http://deb.debian.org#https://mirrors.163.com#g' /etc/apt/sources.list && apt update
RUN apt update && apt install -y  openjdk-11-jre gdbserver cmake make vim

WORKDIR /go/src/github.com/xuperchain/xuperchain
COPY . .
RUN make clean
ENV GOPROXY=https://goproxy.cn
RUN go get github.com/go-delve/delve/cmd/dlv

RUN go build -mod=vendor -o core/xchain-cli github.com/xuperchain/xuperchain/core/cmd/cli
RUN go build -mod=vendor -o core/xchain github.com/xuperchain/xuperchain/core/cmd/xchain
RUN go build -mod=vendor -o core/xdev github.com/xuperchain/xuperchain/core/cmd/xdev
RUN make -C core/xvm/compile/wabt -j 8 && cp core/xvm/compile/wabt/build/wasm2c /bin

RUN mkdir -p core/plugins/kv core/plugins/crypto core/plugins/consensus core/plugins/contract
RUN go build -mod=vendor  --buildmode=plugin --tags multi -o core/plugins/kv/kv-ldb-multi.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-ldb
RUN go build -mod=vendor --buildmode=plugin --tags single -o core/plugins/kv/kv-ldb-single.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-ldb
RUN go build -mod=vendor --buildmode=plugin --tags cloud -o core/plugins/kv/kv-ldb-cloud.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-ldb
RUN go build -mod=vendor --buildmode=plugin -o core/plugins/kv/kv-badger.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-badger
RUN go build -mod=vendor --buildmode=plugin -o core/plugins/crypto/crypto-default.so.1.0.0 github.com/xuperchain/xuperchain/core/crypto/client/xchain/plugin_impl
RUN go build -mod=vendor --buildmode=plugin -o core/plugins/crypto/crypto-schnorr.so.1.0.0 github.com/xuperchain/xuperchain/core/crypto/client/schnorr/plugin_impl
RUN go build -mod=vendor --buildmode=plugin -o core/plugins/crypto/crypto-gm.so.1.0.0 github.com/xuperchain/xuperchain/core/crypto/client/gm/gmclient/plugin_impl
RUN go build -mod=vendor --buildmode=plugin -o core/plugins/consensus/consensus-pow.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/pow
RUN go build -mod=vendor --buildmode=plugin -o core/plugins/consensus/consensus-single.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/single
RUN go build -mod=vendor --buildmode=plugin -o core/plugins/consensus/consensus-tdpos.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/tdpos/main

RUN go build -mod=vendor --buildmode=plugin -o core/plugins/consensus/consensus-xpoa.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/xpoa/main
RUN go build -mod=vendor --buildmode=plugin -o core/plugins/p2p/p2p-p2pv1.so.1.0.0 github.com/xuperchain/xuperchain/core/p2p/p2pv1/plugin_impl
RUN go build -mod=vendor --buildmode=plugin -o core/plugins/p2p/p2p-p2pv2.so.1.0.0 github.com/xuperchain/xuperchain/core/p2p/p2pv2/plugin_impl
RUN go build -mod=vendor --buildmode=plugin -o core/plugins/xendorser/xendorser-default.so.1.0.0 github.com/xuperchain/xuperchain/core/server/xendorser/plugin-default
RUN go build -mod=vendor --buildmode=plugin -o core/plugins/xendorser/xendorser-proxy.so.1.0.0 github.com/xuperchain/xuperchain/core/server/xendorser/plugin-proxy
EXPOSE 40000 40000

WORKDIR /go/src/github.com/xuperchain/xuperchain/core
RUN ./xchain-cli createChain
CMD ["dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "./xchain"]

