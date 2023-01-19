FROM alpine:3.17.1 as conjur-cli-go
LABEL org.opencontainers.image.authors="CyberArk Software Ltd."

ENTRYPOINT [ "/usr/local/bin/conjur" ]

COPY dist/goreleaser/binaries/conjur_linux_amd64_v1 /usr/local/bin/conjur
