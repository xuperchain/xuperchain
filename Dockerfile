FROM golang:1.12.5 AS builder
RUN apt update
WORKDIR /go/src/github.com/xuperchain/xuperunion
COPY . .
RUN make clean&&make

# ---
FROM ubuntu:16.04
RUN apt-get update && apt-get install  host -y
WORKDIR /home/work/xuperunion/
COPY --from=builder /go/src/github.com/xuperchain/xuperunion/output/ .
EXPOSE 37101 47101
RUN ./xchain-cli createChain && ./xchain-cli netURL gen
CMD  ./xchain
