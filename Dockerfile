FROM registry.access.redhat.com/ubi9/ubi-minimal:latest AS conjur-cli-go
LABEL org.opencontainers.image.authors="CyberArk Software Ltd."

ENTRYPOINT [ "/usr/local/bin/conjur" ]

# Install 'tar' so we can use 'kubectl cp' to copy policy files to the container
RUN microdnf install -y tar && microdnf clean all

# Create a non-root user with a home directory for storing the .conjurrc file
RUN groupadd -r cli && useradd --no-log-init -r -g cli cli && \
    mkdir -p /home/cli && chown cli:cli /home/cli

COPY dist/goreleaser/binaries/conjur_linux_amd64_v1 /usr/local/bin/conjur

USER cli
