FROM registry.access.redhat.com/ubi9/ubi-minimal:latest as conjur-cli-go-builder

ENV VERSION=""

RUN microdnf update -y
RUN microdnf install -y go-toolset git

# Add the WORKDIR as a safe directory so git commands
# can be run in containers using this image
RUN git config --global --add safe.directory /github.cyberng.com/Conjur-Enterprise/conjur-cli-go


COPY builder_entrypoint.sh /builder_entrypoint.sh
RUN chmod +x /builder_entrypoint.sh

ENTRYPOINT ["/builder_entrypoint.sh"]
