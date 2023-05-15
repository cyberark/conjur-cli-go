# conjur-cli-go dev

`./start` creates a dev environment, including a CLI container that is already
logged into Conjur as the admin user. You can optionally pass `--oidc-keycloak` to enable
OIDC authentication with Keycloak.

`./stop` tears down the environment.

`./exec` connects to a CLI container if running, otherwise it runs `start`.

## Dev Environment with Okta

Note: This assumes you are able to retrieve dev secrets from ConjurOps via Summon.

`summon -p summon-conjur -f ../ci/okta/secrets.yml -e development ./start --oidc-okta`

## Run Integration Tests

`cd ci && summon -p summon-conjur -f ./okta/secrets.yml -e ci ./test_integration`
