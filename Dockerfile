FROM golang:1.7
RUN apk add --no-cache --virtual .build-deps git make
EXPOSE 8101 8101
WORKDIR /go/src/github.com/dcos/dcos-oauth
COPY . /go/src/github.com/dcos/dcos-oauth
RUN make
