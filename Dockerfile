FROM platform9/go-rpm-build:v1.0
USER root
RUN rm -rf /go
RUN yum install -y wget
RUN wget https://dl.google.com/go/go1.16.6.linux-amd64.tar.gz && tar xf go*.tar.gz
WORKDIR /root/go
ENV GOROOT=/go
ENV GOPATH=/root/go
ENV PATH=${PATH}:${GOPATH}/bin
RUN go get -u 'github.com/jteeuwen/go-bindata/...'
