FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY . .

RUN go build -o /conjur ./cmd/conjur

WORKDIR /src/vhs

RUN go build -o /uniq_frames .

FROM ghcr.io/charmbracelet/vhs:latest
LABEL org.opencontainers.image.authors="CyberArk Software Ltd."
LABEL id="conjur-cli-go-vhs-runner"

COPY --from=builder /conjur /usr/local/bin/conjur
COPY --from=builder /uniq_frames /usr/local/bin/uniq_frames

ENTRYPOINT ["./run_tests.sh"]