FROM golang:1.13-stretch
MAINTAINER Conjur Inc.

RUN apt-get update && \
    apt-get install -y jq less vim

RUN go get -u \
  github.com/jstemmer/go-junit-report \
  github.com/golang/mock/gomock \
  github.com/smartystreets/goconvey \
  golang.org/x/lint/golint \
  github.com/derekparker/delve/cmd/dlv

RUN go install \
  github.com/golang/mock/mockgen \
  golang.org/x/lint/golint

RUN mkdir -p /go/src/github.com/cyberark/conjur-cli-go/output
WORKDIR /go/src/github.com/cyberark/conjur-cli-go

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN ./generate-mocks
