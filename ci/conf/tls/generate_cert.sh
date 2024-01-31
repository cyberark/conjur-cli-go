#!/usr/bin/env bash

# Navigate to the ci/conf/tls directory (where this script lives) to ensure we
# can run this script from anywhere.
cd "$(dirname "$0")"

rm -f nginx.crt nginx.key

openssl genrsa -out nginx.key 2048
openssl req -x509 -key nginx.key -sha256 -days 365 -out nginx.crt -config tls.conf -extensions v3_ca
