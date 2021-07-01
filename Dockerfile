FROM golang:1.13 AS builder
WORKDIR /home/xchain

RUN apt update && apt install -y unzip git

# small trick to take advantage of  docker build cache
COPY go.* ./
COPY Makefile .
RUN  make prepare

COPY . .
RUN make

# ---
FROM ubuntu:18.04
WORKDIR /home/xchain
COPY --from=builder /home/xchain/output .
EXPOSE 37101 37101
CMD bash control.sh start -f
