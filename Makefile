build:
	go build -tags=dev -buildvcs=false -o ./dev/tmp/ ./cmd/conjur

test:
	go test -tags=dev -count=1 -v ./...

install:
	go install -tags=dev -buildvcs=false ./cmd/conjur

build_integration:
	go test -tags=dev,integration -count=1 -c -v ./cmd/integration/...

integration: install
	go test -tags=dev,integration -p=1 -count=1 -v ./cmd/integration/...

integration_cloud: install
	go test -tags=integration -p=1 -count=1 -v ./cmd/integration_cloud/...

# Example usage of run: make run ARGS="variable get -i path/to/var"
run:
	go run -tags=dev ./cmd/conjur $(ARGS)
