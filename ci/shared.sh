#!/usr/bin/env bash

# Retry given cmd every second until success or timeout.
#
# Args:
#   - timeout_secs: Given as an env var.
#   - All remaining args make up the command to test, beginning with the cmd
#     name, and including any flags and arguments to that command.
wait_for_cmd() {
  : "${timeout_secs:=120}"
  local cmd=("$@")

  for _ in $(seq "$timeout_secs"); do
    if "${cmd[@]}"; then
      return 0
    fi
    sleep 1
  done

  return 1
}

_wait_for_pg() {
  local svc=$1
  local pg_cmd=(psql -U postgres -c "select 1" -d postgres)
  local dc_cmd=(docker-compose exec -T "$svc" "${pg_cmd[@]}")

  echo "Waiting for pg to come up..."

  if ! timeout_secs=120 wait_for_cmd "${dc_cmd[@]}"; then
    echo "ERROR: pg service '$svc' failed to come up."
    exit 1
  fi

  echo "Done."
}

wait_for_conjur() {
  docker-compose exec -T conjur conjurctl wait
}

client_init_and_login() {
  local conjur_account=$1
  local admin_password=$2

  docker-compose exec -T cli conjur init -u http://conjur -a "$conjur_account"
  docker-compose exec -T cli conjur login -u admin -p "$admin_password"
}

load_policy_file() {
  local policy_file=$1
  # Loading policy via the CLI doesn't work here for some reason
  docker-compose exec -T conjur conjurctl policy load test "$policy_file"
}

client_add_secret() {
  local variable=$1
  local value=$2
  docker-compose exec -T cli conjur variable set -i "$variable" -v "$value"
}

init_okta() {
  echo "Loading Okta policy..."
  load_policy_file "/opt/conjur-server/okta/okta-2.yml"

  echo "Adding Okta secrets..."
  client_add_secret 'conjur/authn-oidc/okta-2/provider-uri' "${OKTA_PROVIDER_URI}oauth2/default"
  client_add_secret 'conjur/authn-oidc/okta-2/client-id' "$OKTA_CLIENT_ID"
  client_add_secret 'conjur/authn-oidc/okta-2/client-secret' "$OKTA_CLIENT_SECRET"
  client_add_secret 'conjur/authn-oidc/okta-2/claim-mapping' 'preferred_username'
  client_add_secret 'conjur/authn-oidc/okta-2/nonce' '1656b4264b60af659fce'
  client_add_secret 'conjur/authn-oidc/okta-2/state' '4f413476ef7e2395f0af'
  client_add_secret 'conjur/authn-oidc/okta-2/redirect_uri' 'http://127.0.0.1:8888/callback'
}
