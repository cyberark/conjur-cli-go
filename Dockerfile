FROM golang:1.10
MAINTAINER Conjur Inc

RUN apt-get update &&\
    apt-get install -y jq less vim

# RUN mkdir -p /go/src/github.com/cyberark/conjur-cli-go/output
WORKDIR /go/src/github.com/cyberark/conjur-cli-go

COPY ci/get-packages ci/
RUN ci/get-packages

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

COPY . .
RUN ci/generate-mocks

ENV GOOS=linux
ENV GOARCH=amd64

EXPOSE 8080
