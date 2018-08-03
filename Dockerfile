FROM golang:1.10
MAINTAINER Conjur Inc

RUN apt-get update && apt-get install -y jq less 
RUN go get -u \
  github.com/jstemmer/go-junit-report \
  github.com/golang/dep/cmd/dep \
  github.com/golang/mock/gomock \
  github.com/smartystreets/goconvey \
  golang.org/x/lint/golint \ 
  github.com/derekparker/delve/cmd/dlv 

RUN go install \
  github.com/golang/mock/mockgen \
  golang.org/x/lint/golint  

RUN mkdir -p /go/src/github.com/cyberark/conjur-cli-go/output
WORKDIR /go/src/github.com/cyberark/conjur-cli-go

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

COPY . .
RUN ./generate-mocks

ENV GOOS=linux
ENV GOARCH=amd64

EXPOSE 8080
