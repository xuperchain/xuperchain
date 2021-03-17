FROM golang:1.12.5 AS builder
RUN apt update
WORKDIR /go/src/github.com/xuperchain/xuperchain
COPY . .
RUN make clean && make

# ---
FROM ubuntu:16.04
WORKDIR /home/work/xuperunion/
COPY --from=builder /go/src/github.com/xuperchain/xuperchain/output/ .
EXPOSE 37101 47101
CMD ./xchain-cli createChain && ./xchain
