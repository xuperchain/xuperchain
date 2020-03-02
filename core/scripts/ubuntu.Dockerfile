FROM ubuntu:14.04

RUN apt-get update && apt-get install -y gcc-4.8 g++-4.8 make curl git
RUN update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-4.8 60 --slave /usr/bin/g++ g++ /usr/bin/g++-4.8
RUN update-alternatives --install /usr/bin/cc cc /usr/bin/gcc 60

RUN curl -L https://dl.google.com/go/go1.13.8.linux-amd64.tar.gz | tar xzf - -C /opt/
ENV GOPROXY=https://goproxy.cn
ENV PATH="/opt/go/bin:${PATH}"