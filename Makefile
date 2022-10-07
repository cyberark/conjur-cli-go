test:
	go test -count=1 -v ./...

build:
	go build -o ./dev/tmp/ ./cmd/conjur

install:
	go install ./cmd/conjur

run:
	go run ./cmd/conjur

all: build
