FROM golang:1.19-alpine

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go install ./cmd/conjur
