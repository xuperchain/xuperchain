FROM golang:1.13 AS builder
WORKDIR /home/xchain

RUN apt update && apt install -y unzip

# small trick to take advantage of  docker build cache
RUN ls
COPY go.* ./
COPY Makefile .
RUN make prepare

COPY . .
RUN make

# ---
FROM ubuntu:18.04
WORKDIR /home/xchain
RUN apt update&& apt install -y build-essential
COPY --from=builder /home/xchain/output .
EXPOSE 37101 47101
CMD bash control.sh start -f
