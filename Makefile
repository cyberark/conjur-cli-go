build:
	go build -o ./dev/tmp/ conjur.go

install:
	go install conjur.go

run:
	go run conjur.go

all: build
