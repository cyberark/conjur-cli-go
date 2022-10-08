build:
	go build -o ./dev/tmp/ ./cmd/conjur

test:
	go test -count=1 -v ./...

install:
	go install ./cmd/conjur

# Example usage of run: make run ARGS="variable get -i path/to/var" 
run:
	go run ./cmd/conjur $(ARGS)
