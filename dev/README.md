# conjur-cli-go dev

`./start` creates a dev environment, including a CLI container that is already
logged into Conjur as the admin user.

`./stop` tears down the environment.

`./exec` connects to a CLI container if running, otherwise it runs `start`.

## Dev Environment with Okta

Note: This assumes you are able to retrieve dev secrets from ConjurOps via Summon.
`summon -p summon-conjur -f ../ci/okta/secrets.yml ./start --authn-oidc --oidc-okta`

## Run Integration Tests

`cd ci && summon -p summon-conjur -f ./okta/secrets.yml ./test_integration`
