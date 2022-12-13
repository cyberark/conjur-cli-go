build:
	go build -o ./dev/tmp/ ./cmd/conjur

build_dev:
	go build -tags=dev -o ./dev/tmp/ ./cmd/conjur

test:
	go test -count=1 -v -tags=dev ./...

install:
	go install ./cmd/conjur

install_dev:
	go install -tags=dev ./cmd/conjur

build_integration:
	go test -tags=dev -count=1 -c -v --tags=integration ./cmd/integration/...

integration: install
	go test -tags=dev -p=1 -count=1 -v --tags=integration ./cmd/integration/...

# Example usage of run: make run ARGS="variable get -i path/to/var"
run:
	go run ./cmd/conjur $(ARGS)
