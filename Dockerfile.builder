FROM registry.access.redhat.com/ubi7/go-toolset:latest as conjur-cli-go-builder

ENV VERSION=""

COPY --chmod=+x builder_entrypoint.sh /builder_entrypoint.sh

ENTRYPOINT ["/builder_entrypoint.sh"]
