FROM registry.access.redhat.com/ubi9/ubi-minimal:latest as conjur-cli-go
LABEL org.opencontainers.image.authors="CyberArk Software Ltd."

ENTRYPOINT [ "/usr/local/bin/conjur" ]

# Install 'tar' so we can use 'kubectl cp' to copy policy files to the container
RUN microdnf install -y tar

COPY dist/goreleaser/binaries/conjur_linux_amd64_v1 /usr/local/bin/conjur
