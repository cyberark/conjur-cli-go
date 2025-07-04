FROM golang:1.24-alpine AS builder

COPY . .

RUN go build -o /usr/local/bin/conjur ./cmd/conjur

FROM ghcr.io/charmbracelet/vhs:latest
LABEL org.opencontainers.image.authors="CyberArk Software Ltd."
LABEL id="conjur-cli-go-vhs-runner"

COPY --from=builder /usr/local/bin/conjur /usr/local/bin/conjur

ENTRYPOINT ["./run_tests.sh"]