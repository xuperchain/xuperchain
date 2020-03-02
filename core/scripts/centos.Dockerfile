FROM centos:6.6

RUN yum install -y make git curl tar
RUN curl -L http://people.centos.org/tru/devtools-2/devtools-2.repo -o /etc/yum.repos.d/devtools-2.repo
RUN rpm --rebuilddb && yum install -y devtoolset-2-gcc devtoolset-2-binutils devtoolset-2-gcc-c++

RUN curl -L https://dl.google.com/go/go1.13.8.linux-amd64.tar.gz | tar xzf - -C /opt/
ENV GOPROXY=https://goproxy.cn
ENV PATH="/opt/go/bin:${PATH}"
ENV PATH=/opt/rh/devtoolset-2/root/usr/bin${PATH:+:${PATH}}