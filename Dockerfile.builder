FROM mcr.microsoft.com/oss/go/microsoft/golang:1.22-fips-bullseye as conjur-cli-go-builder

ENV VERSION=""

COPY --chmod=+x builder_entrypoint.sh /builder_entrypoint.sh

ENTRYPOINT ["/builder_entrypoint.sh"]
