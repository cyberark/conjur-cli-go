FROM mcr.microsoft.com/oss/go/microsoft/golang:1.24-fips-bullseye as conjur-cli-go-builder

ENV VERSION=""

COPY builder_entrypoint.sh /builder_entrypoint.sh
RUN chmod +x /builder_entrypoint.sh

ENTRYPOINT ["/builder_entrypoint.sh"]
