# conjur-cli-go dev

`./start` creates a dev environment, including a CLI container that is already
logged into Conjur as the admin user.

`./stop` tears down the environment.

`./exec` connects to a CLI container if running, otherwise it runs `start`.

## Dev Environment with Okta
`summon -p summon-conjur -f ./okta/secrets.yml -e development ./start --authn-oidc --oidc-okta`


`conjur logout && conjur init --force -u http://conjur -a dev -t oidc --service-id okta-2 && conjur login -u $OKTA_USERNAME -p $OKTA_PASSWORD && conjur whoami`


## Run Integration Tests
`summon -p summon-conjur -f ./okta/secrets.yml -e ci ./test_integration`

