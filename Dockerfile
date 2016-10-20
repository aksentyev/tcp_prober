FROM golang:1.7-alpine

MAINTAINER Grigory Aksentyev <grigory.aksentiev@gmail.com>

RUN mkdir -p /go/src/github.com/aksentyev/tcp_prober
COPY . /go/src/github.com/aksentyev/tcp_prober
RUN apk add --no-cache git build-base

ENV GOROOT /usr/local/go
RUN cd /go/src/github.com/aksentyev/tcp_prober \
    && go get -d \
    && go test -v \
    && go build -o /bin/tcp_prober \
    && rm -rf /go/src/github.com/aksentyev/tcp_prober

RUN apk del --purge git build-base

RUN mkdir /config
WORKDIR /config

ENTRYPOINT ["/bin/tcp_prober", "-log.level", "debug"]
CMD ["-update-interval", "3", "-scrape-interval", "3"]
