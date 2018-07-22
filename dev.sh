#!/usr/bin/env bash -x

# function finish {
#   echo 'Removing environment'
#   echo '-----'
#   docker-compose down -v
# }
# trap finish EXIT
#

export CONJUR_ACCOUNT=cucumber
export CONJUR_AUTHN_LOGIN=admin

source functions.sh

function main() {
  startConjur
  initEnvironment
  runDevelopment
}

function runDevelopment() {
  local keys=( $(getKeys) )
  local api_key=${keys[0]}
  local api_key_v4=${keys[1]}
  local ssl_cert_v4="$(getCert)"

  export CONJUR_AUTHN_API_KEY="$api_key"
  docker-compose up -d --no-deps cli

  docker-compose build --pull dev
  
  CONJUR_V4_AUTHN_API_KEY="$api_key_v4" \
  CONJUR_V4_SSL_CERTIFICATE="$ssl_cert_v4" \
    docker-compose up --no-deps -d dev
}

main
