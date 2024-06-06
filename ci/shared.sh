#!/usr/bin/env bash

# CC servers can't find it for some reason.  Local shellcheck is fine.
# shellcheck disable=SC1091
source "./keycloak/keycloak_functions.sh"

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
  local dc_cmd=(docker compose exec -T "$svc" "${pg_cmd[@]}")

  echo "Waiting for pg to come up..."

  if ! timeout_secs=120 wait_for_cmd "${dc_cmd[@]}"; then
    echo "ERROR: pg service '$svc' failed to come up."
    exit 1
  fi

  echo "Done."
}

wait_for_conjur() {
  docker compose exec -T conjur conjurctl wait
}


configure_oidc_providers() {
  wait_for_keycloak_server
  fetch_keycloak_certificate
  create_keycloak_users
  echo "keycloak admin console url: http://0.0.0.0:7777/auth/admin"
}

generate_identity_policy() {  
  echo "Generating policy for AuthnOIDC V2 service 'identity' and user '$IDENTITY_USERNAME'"
  policy_dir="./identity"
  rm -f "$policy_dir/users.yml"
  sed -e "s#{{ IDENTITY_USERNAME }}#$IDENTITY_USERNAME#g" "$policy_dir/users.template.yml" > "$policy_dir/users.yml"
}

# Retrieves a JWT from keycloak and saves it to the file `/jwt` in the cli container.
fetch_jwt() {
  TOKEN=$(echo $(docker compose exec cli bash -c 'curl -X POST "http://keycloak:8080/auth/realms/master/protocol/openid-connect/token" \
    --header "Content-Type: application/x-www-form-urlencoded" \
    --data-urlencode "grant_type=client_credentials" \
    --data-urlencode "client_id=conjurClient" \
    --data-urlencode "client_secret=1234"') | jq -r ".access_token")

  docker compose exec cli bash -c "echo $TOKEN > /jwt"
}
