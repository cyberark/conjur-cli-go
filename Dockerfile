FROM golang:1.10
MAINTAINER Conjur Inc

RUN apt-get update && apt-get install -y jq less 
RUN go get -u github.com/jstemmer/go-junit-report
RUN go get -u github.com/golang/dep/cmd/dep
RUN go get github.com/golang/mock/gomock &&\
  go install github.com/golang/mock/mockgen
RUN go get github.com/smartystreets/goconvey
RUN go get -u github.com/derekparker/delve/cmd/dlv

RUN mkdir -p /go/src/github.com/cyberark/conjur-cli-go/output
WORKDIR /go/src/github.com/cyberark/conjur-cli-go

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

COPY . .
RUN mockgen -package mocks -destination action/mocks/client.go github.com/cyberark/conjur-cli-go/action ConjurClient

ENV GOOS=linux
ENV GOARCH=amd64

EXPOSE 8080
